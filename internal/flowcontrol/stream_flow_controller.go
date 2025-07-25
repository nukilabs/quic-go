package flowcontrol

import (
	"fmt"
	"time"

	"github.com/nukilabs/quic-go/internal/protocol"
	"github.com/nukilabs/quic-go/internal/qerr"
	"github.com/nukilabs/quic-go/internal/utils"
)

type streamFlowController struct {
	baseFlowController

	streamID protocol.StreamID

	connection connectionFlowControllerI

	receivedFinalOffset bool
}

var _ StreamFlowController = &streamFlowController{}

// NewStreamFlowController gets a new flow controller for a stream
func NewStreamFlowController(
	streamID protocol.StreamID,
	cfc ConnectionFlowController,
	receiveWindow protocol.ByteCount,
	maxReceiveWindow protocol.ByteCount,
	initialSendWindow protocol.ByteCount,
	rttStats *utils.RTTStats,
	logger utils.Logger,
) StreamFlowController {
	return &streamFlowController{
		streamID:   streamID,
		connection: cfc.(connectionFlowControllerI),
		baseFlowController: baseFlowController{
			rttStats:             rttStats,
			receiveWindow:        receiveWindow,
			receiveWindowSize:    receiveWindow,
			maxReceiveWindowSize: maxReceiveWindow,
			sendWindow:           initialSendWindow,
			logger:               logger,
		},
	}
}

// UpdateHighestReceived updates the highestReceived value, if the offset is higher.
func (c *streamFlowController) UpdateHighestReceived(offset protocol.ByteCount, final bool, now time.Time) error {
	// If the final offset for this stream is already known, check for consistency.
	if c.receivedFinalOffset {
		// If we receive another final offset, check that it's the same.
		if final && offset != c.highestReceived {
			return &qerr.TransportError{
				ErrorCode:    qerr.FinalSizeError,
				ErrorMessage: fmt.Sprintf("received inconsistent final offset for stream %d (old: %d, new: %d bytes)", c.streamID, c.highestReceived, offset),
			}
		}
		// Check that the offset is below the final offset.
		if offset > c.highestReceived {
			return &qerr.TransportError{
				ErrorCode:    qerr.FinalSizeError,
				ErrorMessage: fmt.Sprintf("received offset %d for stream %d, but final offset was already received at %d", offset, c.streamID, c.highestReceived),
			}
		}
	}

	if final {
		c.receivedFinalOffset = true
	}
	if offset == c.highestReceived {
		return nil
	}
	// A higher offset was received before. This can happen due to reordering.
	if offset < c.highestReceived {
		if final {
			return &qerr.TransportError{
				ErrorCode:    qerr.FinalSizeError,
				ErrorMessage: fmt.Sprintf("received final offset %d for stream %d, but already received offset %d before", offset, c.streamID, c.highestReceived),
			}
		}
		return nil
	}

	// If this is the first frame received for this stream, start flow-control auto-tuning.
	if c.highestReceived == 0 {
		c.startNewAutoTuningEpoch(now)
	}
	increment := offset - c.highestReceived
	c.highestReceived = offset

	if c.checkFlowControlViolation() {
		return &qerr.TransportError{
			ErrorCode:    qerr.FlowControlError,
			ErrorMessage: fmt.Sprintf("received %d bytes on stream %d, allowed %d bytes", offset, c.streamID, c.receiveWindow),
		}
	}
	return c.connection.IncrementHighestReceived(increment, now)
}

func (c *streamFlowController) AddBytesRead(n protocol.ByteCount) (hasStreamWindowUpdate, hasConnWindowUpdate bool) {
	c.mutex.Lock()
	c.addBytesRead(n)
	hasStreamWindowUpdate = c.shouldQueueWindowUpdate()
	c.mutex.Unlock()
	hasConnWindowUpdate = c.connection.AddBytesRead(n)
	return
}

func (c *streamFlowController) Abandon() {
	c.mutex.Lock()
	unread := c.highestReceived - c.bytesRead
	c.bytesRead = c.highestReceived
	c.mutex.Unlock()
	if unread > 0 {
		c.connection.AddBytesRead(unread)
	}
}

func (c *streamFlowController) AddBytesSent(n protocol.ByteCount) {
	c.baseFlowController.AddBytesSent(n)
	c.connection.AddBytesSent(n)
}

func (c *streamFlowController) SendWindowSize() protocol.ByteCount {
	return min(c.baseFlowController.SendWindowSize(), c.connection.SendWindowSize())
}

func (c *streamFlowController) IsNewlyBlocked() bool {
	blocked, _ := c.baseFlowController.IsNewlyBlocked()
	return blocked
}

func (c *streamFlowController) shouldQueueWindowUpdate() bool {
	return !c.receivedFinalOffset && c.hasWindowUpdate()
}

func (c *streamFlowController) GetWindowUpdate(now time.Time) protocol.ByteCount {
	// If we already received the final offset for this stream, the peer won't need any additional flow control credit.
	if c.receivedFinalOffset {
		return 0
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	oldWindowSize := c.receiveWindowSize
	offset := c.getWindowUpdate(now)
	if c.receiveWindowSize > oldWindowSize { // auto-tuning enlarged the window size
		c.logger.Debugf("Increasing receive flow control window for stream %d to %d", c.streamID, c.receiveWindowSize)
		c.connection.EnsureMinimumWindowSize(protocol.ByteCount(float64(c.receiveWindowSize)*protocol.ConnectionFlowControlMultiplier), now)
	}
	return offset
}
