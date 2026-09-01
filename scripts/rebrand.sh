#!/usr/bin/env bash
# Reproduces the nukilabs fork's dependency swap mechanically.
# Run from repo root after taking a clean upstream tag:  bash scripts/rebrand.sh
#
# The fork re-points quic-go at the nukilabs fingerprinting stack:
#   crypto/tls                     -> github.com/nukilabs/utls   (aliased tls)
#   net/http (+ sub-pkgs)          -> github.com/nukilabs/http
#   golang.org/x/net/http2/hpack   -> github.com/nukilabs/http/http2/hpack
#   golang.org/x/net/http/httpguts -> github.com/nukilabs/http/httpguts
#   github.com/quic-go/quic-go     -> github.com/nukilabs/quic-go  (module path)
set -euo pipefail

# Pinned versions of the fork dependencies (bump as needed).
HTTP_VER="v1.3.1"
UTLS_VER="v1.3.3"
QPACK_VER="v0.7.0"

# Rewrite import paths in all Go sources. Order matters: the more specific
# net/http sub-packages must be rewritten before the bare "net/http".
find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 perl -pi -e '
  s{"crypto/tls"}{tls "github.com/nukilabs/utls"}g;
  s{"golang.org/x/net/http2/hpack"}{"github.com/nukilabs/http/http2/hpack"}g;
  s{"golang.org/x/net/http/httpguts"}{"github.com/nukilabs/http/httpguts"}g;
  s{"net/http/httptrace"}{"github.com/nukilabs/http/httptrace"}g;
  s{"net/http/httptest"}{"github.com/nukilabs/http/httptest"}g;
  s{"net/http"}{"github.com/nukilabs/http"}g;
  s{github.com/quic-go/qpack}{github.com/nukilabs/qpack}g;
  s{github.com/quic-go/quic-go}{github.com/nukilabs/quic-go}g;
'

# net/http/pprof (a debug-only blank import in the example) has no counterpart
# in the nukilabs/http fork; drop the blank import line, matching the fork.
find . -name '*.go' -not -path './vendor/*' -print0 \
  | xargs -0 perl -ni -e 'print unless m{^\s*_ "net/http/pprof"\s*$}'

# Retarget //go:linkname pulls of TLS 1.3 internals from the standard library to
# the nukilabs/utls fork, which actually provides the TLS implementation used
# here. Without this, SetCipherSuite would mutate the unused crypto/tls slices
# (no effect on utls handshakes), and Go 1.27 rejects the nextTrafficSecret pull
# from the standard library outright. The FIPS-only aeadAESGCMTLS13 pull is left
# on crypto/tls intentionally: utls does not re-export it, and any TLS 1.3
# AES-GCM AEAD is equivalent for quic-go's own packet protection.
# Order matters: the NoAES suffix must be rewritten before its prefix.
find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 perl -pi -e '
  s{crypto/tls\.\(\*cipherSuiteTLS13\)\.nextTrafficSecret}{github.com/nukilabs/utls.(*cipherSuiteTLS13).nextTrafficSecret}g;
  s{crypto/tls\.defaultCipherSuitesTLS13NoAES}{github.com/nukilabs/utls.defaultCipherSuitesTLS13NoAES}g;
  s{crypto/tls\.defaultCipherSuitesTLS13}{github.com/nukilabs/utls.defaultCipherSuitesTLS13}g;
  s{crypto/tls\.cipherSuitesTLS13}{github.com/nukilabs/utls.cipherSuitesTLS13}g;
'

# Module path + go directive, then let the toolchain resolve requires.
go mod edit -module github.com/nukilabs/quic-go
go mod edit -require "github.com/nukilabs/http@${HTTP_VER}"
go mod edit -require "github.com/nukilabs/utls@${UTLS_VER}"
go mod edit -require "github.com/nukilabs/qpack@${QPACK_VER}"

gofmt -w .
go mod tidy

echo "rebrand complete"
