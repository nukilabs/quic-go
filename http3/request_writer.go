package http3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/nukilabs/http"
	"github.com/nukilabs/http/httptrace"

	"github.com/nukilabs/http/http2/hpack"
	"github.com/nukilabs/http/httpguts"
	"golang.org/x/net/idna"

	"github.com/quic-go/qpack"
)

const bodyCopyBufferSize = 8 * 1024

type requestWriter struct {
	mutex             sync.Mutex
	encoder           *qpack.Encoder
	headerBuf         *bytes.Buffer
	pseudoHeaderOrder []string
}

func newRequestWriter(pseudoHeaderOrder []string) *requestWriter {
	headerBuf := &bytes.Buffer{}
	encoder := qpack.NewEncoder(headerBuf)
	return &requestWriter{
		encoder:           encoder,
		headerBuf:         headerBuf,
		pseudoHeaderOrder: pseudoHeaderOrder,
	}
}

func (w *requestWriter) WriteRequestHeader(wr io.Writer, req *http.Request, gzip bool) error {
	// TODO: figure out how to add support for trailers
	buf := &bytes.Buffer{}
	if err := w.writeHeaders(buf, req, gzip); err != nil {
		return err
	}
	if _, err := wr.Write(buf.Bytes()); err != nil {
		return err
	}
	trace := httptrace.ContextClientTrace(req.Context())
	traceWroteHeaders(trace)
	return nil
}

func (w *requestWriter) writeHeaders(wr io.Writer, req *http.Request, gzip bool) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	defer w.encoder.Close()
	defer w.headerBuf.Reset()

	if err := w.encodeHeaders(req, gzip, "", actualContentLength(req)); err != nil {
		return err
	}

	b := make([]byte, 0, 128)
	b = (&headersFrame{Length: uint64(w.headerBuf.Len())}).Append(b)
	if _, err := wr.Write(b); err != nil {
		return err
	}
	_, err := wr.Write(w.headerBuf.Bytes())
	return err
}

func isExtendedConnectRequest(req *http.Request) bool {
	return req.Method == http.MethodConnect && req.Proto != "" && req.Proto != "HTTP/1.1"
}

// copied from net/transport.go
// Modified to support Extended CONNECT:
// Contrary to what the godoc for the http.Request says,
// we do respect the Proto field if the method is CONNECT.
func (w *requestWriter) encodeHeaders(req *http.Request, addGzipHeader bool, trailers string, contentLength int64) error {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	host, err := httpguts.PunycodeHostPort(host)
	if err != nil {
		return err
	}
	if !httpguts.ValidHostHeader(host) {
		return errors.New("http3: invalid Host header")
	}

	// http.NewRequest sets this field to HTTP/1.1
	isExtendedConnect := isExtendedConnectRequest(req)

	var path string
	if req.Method != http.MethodConnect || isExtendedConnect {
		path = req.URL.RequestURI()
		if !validPseudoPath(path) {
			orig := path
			path = strings.TrimPrefix(path, req.URL.Scheme+"://"+host)
			if !validPseudoPath(path) {
				if req.URL.Opaque != "" {
					return fmt.Errorf("invalid request :path %q from URL.Opaque = %q", orig, req.URL.Opaque)
				} else {
					return fmt.Errorf("invalid request :path %q", orig)
				}
			}
		}
	}

	// Check for any invalid headers and return an error before we
	// potentially pollute our hpack state. (We want to be able to
	// continue to reuse the hpack encoder for future requests)
	for k, vv := range req.Header {
		if !httpguts.ValidHeaderFieldName(k) {
			if k == http.HeaderOrderKey {
				continue
			}
			return fmt.Errorf("invalid HTTP header name %q", k)
		}
		for _, v := range vv {
			if !httpguts.ValidHeaderFieldValue(v) {
				return fmt.Errorf("invalid HTTP header value %q for header %q", v, k)
			}
		}
	}

	// Add content-length if we know it, or if it's a method that must have it.
	if shouldSendReqContentLength(req.Method, contentLength) {
		req.Header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}

	// Add accept-encoding if we're adding a gzip header.
	if addGzipHeader {
		req.Header.Set("Accept-Encoding", "gzip")
	}

	// Sort the headers.
	kvs, sorter := req.Header.SortedKeyValues(make(map[string]bool))
	defer http.HeaderSorterPool.Put(sorter)

	enumerateHeaders := func(f func(name, value string)) {
		// 8.1.2.3 Request Pseudo-Header Fields
		// The :path pseudo-header field includes the path and query parts of the
		// target URI (the path-absolute production and optionally a '?' character
		// followed by the query production (see Sections 3.3 and 3.4 of
		// [RFC3986]).
		if w.pseudoHeaderOrder != nil {
			for _, name := range w.pseudoHeaderOrder {
				switch name {
				case ":authority":
					f(":authority", host)
				case ":method":
					f(":method", req.Method)
				case ":path":
					if req.Method != http.MethodConnect || isExtendedConnect {
						f(":path", path)
					}
				case ":scheme":
					if req.Method != http.MethodConnect || isExtendedConnect {
						f(":scheme", req.URL.Scheme)
					}
				}
			}
		} else {
			f(":authority", host)
			f(":method", req.Method)
			if req.Method != http.MethodConnect || isExtendedConnect {
				f(":path", path)
				f(":scheme", req.URL.Scheme)
			}
			if isExtendedConnect {
				f(":protocol", req.Proto)
			}
		}
		if trailers != "" {
			f("trailer", trailers)
		}

		var didUA bool
		for _, kv := range kvs {
			if strings.EqualFold(kv.Key, "host") || strings.EqualFold(kv.Key, "content-length") {
				// Host is :authority, already sent.
				// Content-Length is automatic, set below.
				continue
			} else if strings.EqualFold(kv.Key, "connection") || strings.EqualFold(kv.Key, "proxy-connection") ||
				strings.EqualFold(kv.Key, "transfer-encoding") || strings.EqualFold(kv.Key, "upgrade") ||
				strings.EqualFold(kv.Key, "keep-alive") {
				// Per 8.1.2.2 Connection-Specific Header
				// Fields, don't send connection-specific
				// fields. We have already checked if any
				// are error-worthy so just ignore the rest.
				continue
			} else if strings.EqualFold(kv.Key, "user-agent") {
				// Match Go's http1 behavior: at most one
				// User-Agent. If set to nil or empty string,
				// then omit it. Otherwise if not mentioned,
				// include the default (below).
				didUA = true
				if len(kv.Values) < 1 {
					continue
				}
				kv.Values = kv.Values[:1]
				if kv.Values[0] == "" {
					continue
				}

			}

			for _, v := range kv.Values {
				f(kv.Key, v)
			}
		}
		if !didUA {
			f("user-agent", defaultUserAgent)
		}
	}

	// Do a first pass over the headers counting bytes to ensure
	// we don't exceed cc.peerMaxHeaderListSize. This is done as a
	// separate pass before encoding the headers to prevent
	// modifying the hpack state.
	hlSize := uint64(0)
	enumerateHeaders(func(name, value string) {
		hf := hpack.HeaderField{Name: name, Value: value}
		hlSize += uint64(hf.Size())
	})

	// TODO: check maximum header list size
	// if hlSize > cc.peerMaxHeaderListSize {
	// 	return errRequestHeaderListSize
	// }

	trace := httptrace.ContextClientTrace(req.Context())
	traceHeaders := traceHasWroteHeaderField(trace)

	// Header list size is ok. Write the headers.
	enumerateHeaders(func(name, value string) {
		name = strings.ToLower(name)
		w.encoder.WriteField(qpack.HeaderField{Name: name, Value: value})
		if traceHeaders {
			traceWroteHeaderField(trace, name, value)
		}
	})

	return nil
}

// authorityAddr returns a given authority (a host/IP, or host:port / ip:port)
// and returns a host:port. The port 443 is added if needed.
func authorityAddr(authority string) (addr string) {
	host, port, err := net.SplitHostPort(authority)
	if err != nil { // authority didn't have a port
		port = "443"
		host = authority
	}
	if a, err := idna.ToASCII(host); err == nil {
		host = a
	}
	// IPv6 address literal, without a port:
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host + ":" + port
	}
	return net.JoinHostPort(host, port)
}

// validPseudoPath reports whether v is a valid :path pseudo-header
// value. It must be either:
//
//	*) a non-empty string starting with '/'
//	*) the string '*', for OPTIONS requests.
//
// For now this is only used a quick check for deciding when to clean
// up Opaque URLs before sending requests from the Transport.
// See golang.org/issue/16847
//
// We used to enforce that the path also didn't start with "//", but
// Google's GFE accepts such paths and Chrome sends them, so ignore
// that part of the spec. See golang.org/issue/19103.
func validPseudoPath(v string) bool {
	return (len(v) > 0 && v[0] == '/') || v == "*"
}

// actualContentLength returns a sanitized version of
// req.ContentLength, where 0 actually means zero (not unknown) and -1
// means unknown.
func actualContentLength(req *http.Request) int64 {
	if req.Body == nil {
		return 0
	}
	if req.ContentLength != 0 {
		return req.ContentLength
	}
	return -1
}

// shouldSendReqContentLength reports whether the http2.Transport should send
// a "content-length" request header. This logic is basically a copy of the net/http
// transferWriter.shouldSendContentLength.
// The contentLength is the corrected contentLength (so 0 means actually 0, not unknown).
// -1 means unknown.
func shouldSendReqContentLength(method string, contentLength int64) bool {
	if contentLength > 0 {
		return true
	}
	if contentLength < 0 {
		return false
	}
	// For zero bodies, whether we send a content-length depends on the method.
	// It also kinda doesn't matter for http2 either way, with END_STREAM.
	switch method {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}
