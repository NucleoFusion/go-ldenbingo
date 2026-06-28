# go-ldenbingo

A cross-platform Go-based client library (`go-neto`) designed to communicate with the `EldenBingoServer`.

This package rewrite provides a fully type-safe, asynchronous event-driven client that handles connection handshake, framing, keepalives, MessagePack serialization mapping (matching .NET's `ContractlessStandardResolver`), and C#'s custom `Lz4BlockArray` compression.

## Installation

Add the package to your `go.mod`:

```bash
go get go-ldenbingo/neto
```

## Features

- **Length-Prefixed Framing**: Automatic processing of TCP headers.
- **Asynchronous Read Loop**: Independent read loop that handles incoming packets concurrently.
- **LZ4 Decompression**: Fully supports C# MessagePack-CSharp's `Lz4BlockArray` (Extension Type 98) and `Lz4Block` (Extension Type 99) compression transparently.
- **Event Dispatching**: Register type-safe callbacks for EldenBingo events (e.g., coordinates, board updates, chat).
- **Heartbeat Watchdog**: Automatic keep-alive detection that drops/reconnects if the server becomes unresponsive.

## Getting Started

Here is a complete example of connecting to the server, joining a lobby, listening for coordinates updates, and sending coordinates:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"go-ldenbingo/neto"
)

func main() {
	// Create the client
	// Pass address, port, and a unique identity token (to recover connection if disconnected)
	client := neto.NewNetoClient("127.0.0.1", 4501, "my_secret_token")

	// Set connection lifecycle callbacks
	client.OnConnected = func() {
		fmt.Println("Handshake complete, successfully registered on server!")

		// Let's send a request to join a room
		joinReq := neto.ClientRequestJoinRoom{
			RoomName:  "TestLobby",
			AdminPass: "",
			Nick:      "GoPlayer",
			Team:      0, // Red Team
		}
		
		if err := client.Send(joinReq); err != nil {
			log.Printf("Failed to send join request: %v", err)
		}
	}

	client.OnDisconnected = func(err error) {
		fmt.Printf("Disconnected from server: %v\n", err)
	}

	client.OnKicked = func(reason string) {
		fmt.Printf("Kicked by server. Reason: %s\n", reason)
	}

	client.OnStatus = func(status string) {
		fmt.Printf("[Status] %s\n", status)
	}

	client.OnError = func(err error) {
		fmt.Printf("[Error] %v\n", err)
	}

	// Register event handlers for specific packet types
	
	// 1. Join Room Response
	client.On(neto.ServerJoinRoomAccepted{}, func(obj interface{}) {
		res := obj.(*neto.ServerJoinRoomAccepted)
		fmt.Printf("Successfully joined lobby: %s. Current active users: %d\n", res.RoomName, len(res.Users))
	})

	// 2. Chat messages
	client.On(neto.ServerUserChat{}, func(obj interface{}) {
		chat := obj.(*neto.ServerUserChat)
		fmt.Printf("<User %s> %s\n", chat.UserGuid, chat.Message)
	})

	// 3. User coordinates stream
	client.On(neto.ServerUserCoordinates{}, func(obj interface{}) {
		coords := obj.(*neto.ServerUserCoordinates)
		fmt.Printf("User %s is at X: %.2f, Y: %.2f (DLC: %v)\n", 
			coords.UserGuid, coords.X, coords.Y, coords.MapInstance == neto.DLC)
	})

	// Connect to the server
	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Periodically send coordinates to simulate gameplay
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if client.IsConnected() {
				coords := neto.ClientCoordinates{
					X:             105.45,
					Y:             -320.12,
					Angle:         1.57,
					IsUnderground: false,
					MapInstance:   neto.MainMap,
				}
				if err := client.Send(coords); err != nil {
					log.Printf("Failed to send coordinates: %v", err)
				}
			}
		}
	}()

	// Keep the main thread alive
	select {}
}
```

## Structure & Architecture

```text
├── go.mod
├── neto
│   ├── client.go      # NetoClient connection, read-loop, and event callbacks
│   ├── lz4.go         # Decompression algorithm for C# Ext 98/99 payloads
│   ├── packets.go     # Packet structures, envelope logic, and MsgPack registry
│   └── types.go       # Core types, constants, and enums from EldenBingoCommon
└── README.md          # Getting started instructions
```
