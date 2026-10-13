package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	ws "go-boilerplate/pkg/websocket"
)

func main() {
	address := flag.String("url", "http://localhost:8080", "running HTTP app")
	origin := flag.String("origin", "https://example.com", "allowed Origin")
	wait := flag.Bool("wait-shutdown", false, "require close code 1001 after the app stops")
	flag.Parse()
	if err := run(*address, *origin, *wait); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(address, origin string, wait bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := register(ctx, address)
	if err != nil {
		return err
	}
	url := strings.Replace(address, "http", "ws", 1) + "/v1/ws"
	conn, handshake, err := (&websocket.Dialer{HandshakeTimeout: 3 * time.Second}).DialContext(ctx, url, http.Header{"Authorization": {"Bearer " + token}, "Origin": {origin}, "X-Request-ID": {"ws-runtime"}, "Accept-Encoding": {"gzip"}})
	if handshake != nil {
		defer handshake.Body.Close()
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return err
	}
	if handshake.Header.Get("X-Request-ID") != "ws-runtime" || handshake.Header.Get("Content-Encoding") != "" {
		return fmt.Errorf("invalid handshake middleware headers")
	}
	if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return err
	}
	var message ws.Message
	if err := conn.ReadJSON(&message); err != nil {
		return err
	}
	if message.Type != "notification" || message.RequestID != "ws-runtime" {
		return fmt.Errorf("invalid notification")
	}
	if err := conn.WriteJSON(ws.Message{Type: "echo", Data: "runtime"}); err != nil {
		return err
	}
	if err := conn.ReadJSON(&message); err != nil {
		return err
	}
	if message != (ws.Message{Type: "echo", Data: "runtime", RequestID: "ws-runtime"}) {
		return fmt.Errorf("invalid echo")
	}
	fmt.Println("WebSocket connected")
	if wait {
		return waitForShutdown(conn)
	}
	return nil
}

func waitForShutdown(conn *websocket.Conn) error {
	_, _, err := conn.ReadMessage()
	if err == nil {
		return fmt.Errorf("expected shutdown close 1001, received a message")
	}
	if !websocket.IsCloseError(err, 1001) {
		return fmt.Errorf("expected shutdown close 1001: %w", err)
	}
	fmt.Println("WebSocket shutdown verified")
	return nil
}

func register(ctx context.Context, address string) (string, error) {
	body := fmt.Sprintf(`{"email":"ws-%d@example.com","password":"RuntimeTest123!","name":"WebSocket runtime"}`, time.Now().UnixNano())
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/v1/auth/register", bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("registration status %d", response.StatusCode)
	}
	var registration struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&registration); err != nil {
		return "", err
	}
	return registration.Data.AccessToken, nil
}
