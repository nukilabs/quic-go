module github.com/nukilabs/quic-go

go 1.25

require (
	github.com/quic-go/qpack v0.6.0
	github.com/stretchr/testify v1.11.1
	go.uber.org/mock v0.5.2
	golang.org/x/crypto v0.42.0
	golang.org/x/net v0.44.0
	golang.org/x/sync v0.17.0
	golang.org/x/sys v0.36.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jordanlewis/gcassert v0.0.0-20250430164644-389ef753e22e // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/nukilabs/http v1.1.1
	github.com/nukilabs/utls v1.1.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool (
	github.com/jordanlewis/gcassert/cmd/gcassert
	go.uber.org/mock/mockgen
)
