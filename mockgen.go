//go:build gomock || generate

package quic

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_send_conn_test.go github.com/nukilabs/quic-go SendConn"
type SendConn = sendConn

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_raw_conn_test.go github.com/nukilabs/quic-go RawConn"
type RawConn = rawConn

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_sender_test.go github.com/nukilabs/quic-go Sender"
type Sender = sender

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_stream_sender_test.go github.com/nukilabs/quic-go StreamSender"
type StreamSender = streamSender

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_stream_control_frame_getter_test.go github.com/nukilabs/quic-go StreamControlFrameGetter"
type StreamControlFrameGetter = streamControlFrameGetter

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_stream_frame_getter_test.go github.com/nukilabs/quic-go StreamFrameGetter"
type StreamFrameGetter = streamFrameGetter

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_frame_source_test.go github.com/nukilabs/quic-go FrameSource"
type FrameSource = frameSource

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_ack_frame_source_test.go github.com/nukilabs/quic-go AckFrameSource"
type AckFrameSource = ackFrameSource

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_sealing_manager_test.go github.com/nukilabs/quic-go SealingManager"
type SealingManager = sealingManager

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_unpacker_test.go github.com/nukilabs/quic-go Unpacker"
type Unpacker = unpacker

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_packer_test.go github.com/nukilabs/quic-go Packer"
type Packer = packer

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_mtu_discoverer_test.go github.com/nukilabs/quic-go MTUDiscoverer"
type MTUDiscoverer = mtuDiscoverer

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_conn_runner_test.go github.com/nukilabs/quic-go ConnRunner"
type ConnRunner = connRunner

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/nukilabs/quic-go -destination mock_packet_handler_test.go github.com/nukilabs/quic-go PacketHandler"
type PacketHandler = packetHandler

//go:generate sh -c "go run go.uber.org/mock/mockgen -typed -package quic -self_package github.com/nukilabs/quic-go -self_package github.com/nukilabs/quic-go -destination mock_packetconn_test.go net PacketConn"
