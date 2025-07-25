package wire

import (
	"github.com/nukilabs/quic-go/internal/protocol"
)

// A HandshakeDoneFrame is a HANDSHAKE_DONE frame
type HandshakeDoneFrame struct{}

func (f *HandshakeDoneFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	return append(b, handshakeDoneFrameType), nil
}

// Length of a written frame
func (f *HandshakeDoneFrame) Length(_ protocol.Version) protocol.ByteCount {
	return 1
}
