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
HTTP_VER="v1.1.1"
UTLS_VER="v1.1.1"

# Rewrite import paths in all Go sources. Order matters: the more specific
# net/http sub-packages must be rewritten before the bare "net/http".
find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 perl -pi -e '
  s{"crypto/tls"}{tls "github.com/nukilabs/utls"}g;
  s{"golang.org/x/net/http2/hpack"}{"github.com/nukilabs/http/http2/hpack"}g;
  s{"golang.org/x/net/http/httpguts"}{"github.com/nukilabs/http/httpguts"}g;
  s{"net/http/httptrace"}{"github.com/nukilabs/http/httptrace"}g;
  s{"net/http/httptest"}{"github.com/nukilabs/http/httptest"}g;
  s{"net/http"}{"github.com/nukilabs/http"}g;
  s{github.com/quic-go/quic-go}{github.com/nukilabs/quic-go}g;
'

# net/http/pprof (a debug-only blank import in the example) has no counterpart
# in the nukilabs/http fork; drop the blank import line, matching the fork.
find . -name '*.go' -not -path './vendor/*' -print0 \
  | xargs -0 perl -ni -e 'print unless m{^\s*_ "net/http/pprof"\s*$}'

# Module path + go directive, then let the toolchain resolve requires.
go mod edit -module github.com/nukilabs/quic-go
go mod edit -require "github.com/nukilabs/http@${HTTP_VER}"
go mod edit -require "github.com/nukilabs/utls@${UTLS_VER}"

gofmt -w .
go mod tidy

echo "rebrand complete"
