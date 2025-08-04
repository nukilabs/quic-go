package http3

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/nukilabs/quic-go"
	"github.com/nukilabs/quic-go/quicvarint"
)

// FrameType is the frame type of a HTTP/3 frame
type FrameType uint64

type unknownFrameHandlerFunc func(FrameType, error) (processed bool, err error)

type frame interface{}

var errHijacked = errors.New("hijacked")

type frameParser struct {
	r                   io.Reader
	closeConn           func(quic.ApplicationErrorCode, string) error
	unknownFrameHandler unknownFrameHandlerFunc
}

func (p *frameParser) ParseNext() (frame, error) {
	qr := quicvarint.NewReader(p.r)
	for {
		t, err := quicvarint.Read(qr)
		if err != nil {
			if p.unknownFrameHandler != nil {
				hijacked, err := p.unknownFrameHandler(0, err)
				if err != nil {
					return nil, err
				}
				if hijacked {
					return nil, errHijacked
				}
			}
			return nil, err
		}
		// Call the unknownFrameHandler for frames not defined in the HTTP/3 spec
		if t > 0xd && p.unknownFrameHandler != nil {
			hijacked, err := p.unknownFrameHandler(FrameType(t), nil)
			if err != nil {
				return nil, err
			}
			if hijacked {
				return nil, errHijacked
			}
			// If the unknownFrameHandler didn't process the frame, it is our responsibility to skip it.
		}
		l, err := quicvarint.Read(qr)
		if err != nil {
			return nil, err
		}

		switch t {
		case 0x0:
			return &dataFrame{Length: l}, nil
		case 0x1:
			return &headersFrame{Length: l}, nil
		case 0x4:
			return parseSettingsFrame(p.r, l)
		case 0x3: // CANCEL_PUSH
		case 0x5: // PUSH_PROMISE
		case 0x7:
			return parseGoAwayFrame(qr, l)
		case 0xd: // MAX_PUSH_ID
		case 0x2, 0x6, 0x8, 0x9:
			p.closeConn(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "")
			return nil, fmt.Errorf("http3: reserved frame type: %d", t)
		}
		// skip over unknown frames
		if _, err := io.CopyN(io.Discard, qr, int64(l)); err != nil {
			return nil, err
		}
	}
}

type dataFrame struct {
	Length uint64
}

func (f *dataFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x0)
	return quicvarint.Append(b, f.Length)
}

type headersFrame struct {
	Length uint64
}

func (f *headersFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x1)
	return quicvarint.Append(b, f.Length)
}

const (
	// QPACK maximum table capacity, RFC 9204
	SettingQpackMaxTableCapacity = 0x1
	// Maximum field section size, RFC 9114
	SettingMaxFieldSectionSize = 0x6
	// QPACK blocked streams, RFC 9204
	SettingQpackBlockedStreams = 0x7
	// Extended CONNECT, RFC 9220
	SettingExtendedConnect = 0x8
	// HTTP Datagrams, RFC 9297
	SettingH3Datagram = 0x33
	// Enable Metadata, draft-beky-httpbis-metadata-02
	SettingEnableMetadata = 0x4d44
	// GREASE identifier, draft-edm-protocol-greasing-05
	SettingGrease = 0x1f*1 + 0x21
)

type Setting struct {
	ID  uint64
	Val uint64
}

type settingsFrame struct {
	Datagram        bool // HTTP Datagrams, RFC 9297
	ExtendedConnect bool // Extended CONNECT, RFC 9220

	Other map[uint64]uint64 // all settings that we don't explicitly recognize
	Order []uint64          // order of settings
}

func parseSettingsFrame(r io.Reader, l uint64) (*settingsFrame, error) {
	if l > 8*(1<<10) {
		return nil, fmt.Errorf("unexpected size for SETTINGS frame: %d", l)
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}
	frame := &settingsFrame{}
	b := bytes.NewReader(buf)
	var readDatagram, readExtendedConnect bool
	for b.Len() > 0 {
		id, err := quicvarint.Read(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}
		val, err := quicvarint.Read(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}

		switch id {
		case SettingExtendedConnect:
			if readExtendedConnect {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			readExtendedConnect = true
			if val != 0 && val != 1 {
				return nil, fmt.Errorf("invalid value for SETTINGS_ENABLE_CONNECT_PROTOCOL: %d", val)
			}
			frame.ExtendedConnect = val == 1
		case SettingH3Datagram:
			if readDatagram {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			readDatagram = true
			if val != 0 && val != 1 {
				return nil, fmt.Errorf("invalid value for SETTINGS_H3_DATAGRAM: %d", val)
			}
			frame.Datagram = val == 1
		default:
			if _, ok := frame.Other[id]; ok {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			if frame.Other == nil {
				frame.Other = make(map[uint64]uint64)
			}
			frame.Other[id] = val
		}
	}
	return frame, nil
}

func (f *settingsFrame) Append(b []byte) []byte {
	if f.Order != nil {
		return f.AppendInOrder(b)
	}

	b = quicvarint.Append(b, 0x4)
	var l int
	var datagramAdded, extendedConnectAdded bool
	var n1, n2 uint64
	for id, val := range f.Other {
		if id == SettingGrease {
			n, err := rand.Int(rand.Reader, big.NewInt(1<<30))
			if err != nil {
				n = big.NewInt(1)
			}
			n1 = n.Uint64()
			if val != 0 {
				n2 = val % (1 << 30)
			} else {
				n2 = val
			}
			l += quicvarint.Len(n1) + quicvarint.Len(n2)
		} else {
			l += quicvarint.Len(id) + quicvarint.Len(val)
		}
		if id == SettingH3Datagram {
			datagramAdded = true
		}
		if id == SettingExtendedConnect {
			extendedConnectAdded = true
		}
	}
	if f.Datagram && !datagramAdded {
		l += quicvarint.Len(SettingH3Datagram) + quicvarint.Len(1)
	}
	if f.ExtendedConnect && !extendedConnectAdded {
		l += quicvarint.Len(SettingExtendedConnect) + quicvarint.Len(1)
	}
	b = quicvarint.Append(b, uint64(l))
	for id, val := range f.Other {
		if id == SettingGrease {
			b = quicvarint.Append(b, n1)
			b = quicvarint.Append(b, n2)
		} else {
			b = quicvarint.Append(b, id)
			b = quicvarint.Append(b, val)
		}
	}
	if f.Datagram && !datagramAdded {
		b = quicvarint.Append(b, SettingH3Datagram)
		b = quicvarint.Append(b, 1)
	}
	if f.ExtendedConnect && !extendedConnectAdded {
		b = quicvarint.Append(b, SettingExtendedConnect)
		b = quicvarint.Append(b, 1)
	}
	return b
}

func (f *settingsFrame) AppendInOrder(b []byte) []byte {
	b = quicvarint.Append(b, 0x4)
	var l int
	var datagramAdded, extendedConnectAdded bool
	var n1, n2 uint64

	// Calculate length first, respecting order
	for _, id := range f.Order {
		if val, exists := f.Other[id]; exists {
			if id == SettingGrease {
				n, err := rand.Int(rand.Reader, big.NewInt(1<<30))
				if err != nil {
					n = big.NewInt(1)
				}
				n1 = n.Uint64()
				if val != 0 {
					n2 = val % (1 << 30)
				} else {
					n2 = val
				}
				l += quicvarint.Len(n1) + quicvarint.Len(n2)
			} else {
				l += quicvarint.Len(id) + quicvarint.Len(val)
			}
			if id == SettingH3Datagram {
				datagramAdded = true
			}
			if id == SettingExtendedConnect {
				extendedConnectAdded = true
			}
		}
	}

	// Add Datagram and ExtendedConnect if they are enabled but not in Order
	if f.Datagram && !datagramAdded {
		l += quicvarint.Len(SettingH3Datagram) + quicvarint.Len(1)
	}
	if f.ExtendedConnect && !extendedConnectAdded {
		l += quicvarint.Len(SettingExtendedConnect) + quicvarint.Len(1)
	}

	b = quicvarint.Append(b, uint64(l))

	// Append settings in the specified order
	for _, id := range f.Order {
		if val, exists := f.Other[id]; exists {
			if id == SettingGrease {
				b = quicvarint.Append(b, n1)
				b = quicvarint.Append(b, n2)
			} else {
				b = quicvarint.Append(b, id)
				b = quicvarint.Append(b, val)
			}
		}
	}

	// Add Datagram and ExtendedConnect if they are enabled but not in Order
	if f.Datagram && !datagramAdded {
		b = quicvarint.Append(b, SettingH3Datagram)
		b = quicvarint.Append(b, 1)
	}
	if f.ExtendedConnect && !extendedConnectAdded {
		b = quicvarint.Append(b, SettingExtendedConnect)
		b = quicvarint.Append(b, 1)
	}
	return b
}

type goAwayFrame struct {
	StreamID quic.StreamID
}

func parseGoAwayFrame(r io.ByteReader, l uint64) (*goAwayFrame, error) {
	frame := &goAwayFrame{}
	cbr := countingByteReader{ByteReader: r}
	id, err := quicvarint.Read(&cbr)
	if err != nil {
		return nil, err
	}
	if cbr.Read != int(l) {
		return nil, errors.New("GOAWAY frame: inconsistent length")
	}
	frame.StreamID = quic.StreamID(id)
	return frame, nil
}

func (f *goAwayFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x7)
	b = quicvarint.Append(b, uint64(quicvarint.Len(uint64(f.StreamID))))
	return quicvarint.Append(b, uint64(f.StreamID))
}
