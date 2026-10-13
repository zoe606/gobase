package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"

	ws "go-boilerplate/pkg/websocket"
)

func main() {
	address := flag.String("url", "ws://localhost:8080/v1/ws", "WebSocket URL")
	origin := flag.String("origin", "http://localhost:3000", "allowed Origin")
	flag.Parse()
	if err := run(*address, *origin, os.Getenv("WS_ACCESS_TOKEN")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(address, origin, token string) error {
	if token == "" {
		return fmt.Errorf("set WS_ACCESS_TOKEN to an access token")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, response, err := (&websocket.Dialer{HandshakeTimeout: 3 * time.Second}).DialContext(ctx, address, http.Header{"Authorization": {"Bearer " + token}, "Origin": {origin}})
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	var message ws.Message
	if err := conn.ReadJSON(&message); err != nil {
		return err
	}
	fmt.Printf("%s: %s\n", message.Type, message.Data)
	if err := conn.WriteJSON(ws.Message{Type: "echo", Data: "hello"}); err != nil {
		return err
	}
	if err := conn.ReadJSON(&message); err != nil {
		return err
	}
	fmt.Printf("%s: %s\n", message.Type, message.Data)
	return conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
}
