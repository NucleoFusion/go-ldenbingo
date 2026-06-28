package main

import (
	"bufio"
	"fmt"
	"go-ldenbingo/neto"
	"os"
	"strings"
)

func main() {
	client := neto.NewNetoClient("127.0.0.1", 4501, "")

	client.OnStatus = func(s string) { fmt.Println("[status]", s) }
	client.OnError = func(err error) { fmt.Println("[error]", err) }
	client.OnDisconnected = func(err error) { fmt.Println("[disconnected]", err) }
	client.OnKicked = func(reason string) { fmt.Println("[kicked]", reason) }
	client.OnConnected = func() {
		fmt.Println("[connected] guid:", client.ClientGuid())
	}

	client.On(&neto.ServerJoinRoomAccepted{}, func(obj interface{}) {
		fmt.Printf("[ServerJoinRoomAccepted] %+v\n", obj)
	})
	client.On(&neto.ServerJoinRoomDenied{}, func(obj interface{}) {
		fmt.Printf("[ServerJoinRoomDenied] %+v\n", obj)
	})
	client.On(&neto.ServerUserJoinedRoom{}, func(obj interface{}) {
		fmt.Printf("[ServerUserJoinedRoom] %+v\n", obj)
	})
	client.On(&neto.ServerEntireBingoBoardUpdate{}, func(obj interface{}) {
		fmt.Printf("[ServerEntireBingoBoardUpdate] %+v\n", obj)
	})

	if err := client.Connect(); err != nil {
		fmt.Println("connect failed:", err)
		os.Exit(1)
	}

	fmt.Println("Type 'create <roomname> <nick>' or 'join <roomname> <nick>', then Enter:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 3 {
			fmt.Println("usage: create|join <roomname> <nick>")
			continue
		}
		cmd, room, nick := parts[0], parts[1], parts[2]

		switch cmd {
		case "create":
			err := client.Send(neto.ClientRequestCreateRoom{
				RoomName: room,
				Nick:     nick,
				Team:     0,
				Settings: neto.BingoGameSettings{BoardSize: 5},
			})
			fmt.Println("send create err:", err)
		case "join":
			err := client.Send(neto.ClientRequestJoinRoom{
				RoomName: room,
				Nick:     nick,
				Team:     0,
			})
			fmt.Println("send join err:", err)
		default:
			fmt.Println("unknown command")
		}
	}
}
