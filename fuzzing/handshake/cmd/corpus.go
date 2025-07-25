package main

import (
	"context"
	"log"
	"net"

	tls "github.com/nukilabs/utls"

	fuzzhandshake "github.com/nukilabs/quic-go/fuzzing/handshake"
	"github.com/nukilabs/quic-go/fuzzing/internal/helper"
	"github.com/nukilabs/quic-go/internal/handshake"
	"github.com/nukilabs/quic-go/internal/protocol"
	"github.com/nukilabs/quic-go/internal/testdata"
	"github.com/nukilabs/quic-go/internal/utils"
	"github.com/nukilabs/quic-go/internal/wire"
)

const alpn = "fuzz"

func main() {
	client := handshake.NewCryptoSetupClient(
		protocol.ConnectionID{},
		&wire.TransportParameters{ActiveConnectionIDLimit: 2},
		&tls.Config{
			MinVersion:         tls.VersionTLS13,
			ServerName:         "localhost",
			NextProtos:         []string{alpn},
			RootCAs:            testdata.GetRootCA(),
			ClientSessionCache: tls.NewLRUClientSessionCache(1),
		},
		false,
		&utils.RTTStats{},
		nil,
		utils.DefaultLogger.WithPrefix("client"),
		protocol.Version1,
	)

	config := testdata.GetTLSConfig()
	config.NextProtos = []string{alpn}
	server := handshake.NewCryptoSetupServer(
		protocol.ConnectionID{},
		&net.UDPAddr{IP: net.IPv6loopback, Port: 1234},
		&net.UDPAddr{IP: net.IPv6loopback, Port: 4321},
		&wire.TransportParameters{ActiveConnectionIDLimit: 2},
		config,
		false,
		&utils.RTTStats{},
		nil,
		utils.DefaultLogger.WithPrefix("server"),
		protocol.Version1,
	)

	if err := client.StartHandshake(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := server.StartHandshake(context.Background()); err != nil {
		log.Fatal(err)
	}

	var clientHandshakeComplete, serverHandshakeComplete bool
	var messages [][]byte
	for {
	clientLoop:
		for {
			ev := client.NextEvent()
			//nolint:exhaustive // only need to process a few events
			switch ev.Kind {
			case handshake.EventNoEvent:
				break clientLoop
			case handshake.EventWriteInitialData:
				messages = append(messages, ev.Data)
				if err := server.HandleMessage(ev.Data, protocol.EncryptionInitial); err != nil {
					log.Fatal(err)
				}
			case handshake.EventWriteHandshakeData:
				messages = append(messages, ev.Data)
				if err := server.HandleMessage(ev.Data, protocol.EncryptionHandshake); err != nil {
					log.Fatal(err)
				}
			case handshake.EventHandshakeComplete:
				clientHandshakeComplete = true
			}
		}

	serverLoop:
		for {
			ev := server.NextEvent()
			//nolint:exhaustive // only need to process a few events
			switch ev.Kind {
			case handshake.EventNoEvent:
				break serverLoop
			case handshake.EventWriteInitialData:
				messages = append(messages, ev.Data)
				if err := client.HandleMessage(ev.Data, protocol.EncryptionInitial); err != nil {
					log.Fatal(err)
				}
			case handshake.EventWriteHandshakeData:
				messages = append(messages, ev.Data)
				if err := client.HandleMessage(ev.Data, protocol.EncryptionHandshake); err != nil {
					log.Fatal(err)
				}
			case handshake.EventHandshakeComplete:
				serverHandshakeComplete = true
			}
		}

		if serverHandshakeComplete && clientHandshakeComplete {
			break
		}
	}

	ticket, err := server.GetSessionTicket()
	if err != nil {
		log.Fatal(err)
	}
	if ticket == nil {
		log.Fatal("expected a session ticket")
	}
	messages = append(messages, ticket)

	for _, m := range messages {
		if err := helper.WriteCorpusFileWithPrefix("corpus", m, fuzzhandshake.PrefixLen); err != nil {
			log.Fatal(err)
		}
	}
}
