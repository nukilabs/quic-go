// Code generated by go generate. DO NOT EDIT.
// Source: sys_conn_buffers.go

package quic

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/nukilabs/quic-go/internal/protocol"
	"github.com/nukilabs/quic-go/internal/utils"
)

func setSendBuffer(c net.PacketConn) error {
	conn, ok := c.(interface{ SetWriteBuffer(int) error })
	if !ok {
		return errors.New("connection doesn't allow setting of send buffer size. Not a *net.UDPConn?")
	}

	var syscallConn syscall.RawConn
	if sc, ok := c.(interface {
		SyscallConn() (syscall.RawConn, error)
	}); ok {
		var err error
		syscallConn, err = sc.SyscallConn()
		if err != nil {
			syscallConn = nil
		}
	}
	// The connection has a SetWriteBuffer method, but we couldn't obtain a syscall.RawConn.
	// This shouldn't happen for a net.UDPConn, but is possible if the connection just implements the
	// net.PacketConn interface and the SetWriteBuffer method.
	// We have no way of checking if increasing the buffer size actually worked.
	if syscallConn == nil {
		return conn.SetWriteBuffer(protocol.DesiredSendBufferSize)
	}

	size, err := inspectWriteBuffer(syscallConn)
	if err != nil {
		return fmt.Errorf("failed to determine send buffer size: %w", err)
	}
	if size >= protocol.DesiredSendBufferSize {
		utils.DefaultLogger.Debugf("Conn has send buffer of %d kiB (wanted: at least %d kiB)", size/1024, protocol.DesiredSendBufferSize/1024)
		return nil
	}
	// Ignore the error. We check if we succeeded by querying the buffer size afterward.
	_ = conn.SetWriteBuffer(protocol.DesiredSendBufferSize)
	newSize, err := inspectWriteBuffer(syscallConn)
	if newSize < protocol.DesiredSendBufferSize {
		// Try again with RCVBUFFORCE on Linux
		_ = forceSetSendBuffer(syscallConn, protocol.DesiredSendBufferSize)
		newSize, err = inspectWriteBuffer(syscallConn)
		if err != nil {
			return fmt.Errorf("failed to determine send buffer size: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to determine send buffer size: %w", err)
	}
	if newSize == size {
		return fmt.Errorf("failed to increase send buffer size (wanted: %d kiB, got %d kiB)", protocol.DesiredSendBufferSize/1024, newSize/1024)
	}
	if newSize < protocol.DesiredSendBufferSize {
		return fmt.Errorf("failed to sufficiently increase send buffer size (was: %d kiB, wanted: %d kiB, got: %d kiB)", size/1024, protocol.DesiredSendBufferSize/1024, newSize/1024)
	}
	utils.DefaultLogger.Debugf("Increased send buffer size to %d kiB", newSize/1024)
	return nil
}
