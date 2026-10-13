package httphandler

import (
	"net"
	"strings"
	"time"

	fastws "github.com/fasthttp/websocket"
	transport "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/response"
	"go-boilerplate/pkg/websocket"
)

type fiberWebSocketConnection struct {
	*fastws.Conn
	socket net.Conn
}

// fasthttp closes its wrapper after the callback returns. Close the socket to interrupt active reads.
func (c *fiberWebSocketConnection) Close() error { return c.socket.Close() }

func SetupWebSocketRoutes(app *fiber.App, cfg config.WebSocket, service jwt.Service, server *websocket.Server) {
	if !cfg.Enabled {
		return
	}
	upgrade := transport.New(serveFiberWebSocket, transport.Config{HandshakeTimeout: cfg.HandshakeTimeout, Origins: cfg.Origins(), RecoverHandler: func(conn *transport.Conn) {
		if recover() != nil {
			_ = conn.WriteControl(transport.CloseMessage, transport.FormatCloseMessage(1011, "internal error"), time.Now().Add(cfg.WriteTimeout))
			_ = conn.Close()
		}
	}})
	app.Get("/v1/ws", func(c *fiber.Ctx) error {
		claims, err := websocketClaims(service, c.Get("Authorization"))
		if err != nil {
			return response.Unauthorized(c, "Invalid WebSocket credentials")
		}
		if !cfg.AllowsOrigin(c.Get("Origin")) {
			return response.Forbidden(c, "WebSocket origin is not allowed")
		}
		if c.Method() != fiber.MethodGet || !transport.IsWebSocketUpgrade(c) || !websocketHeadersValid(c.Get("Sec-WebSocket-Version"), c.Get("Sec-WebSocket-Key")) {
			c.Set("Sec-WebSocket-Version", "13")
			return response.Error(c, fiber.StatusUpgradeRequired, "UPGRADE_REQUIRED", "WebSocket upgrade required")
		}
		session, ok := server.Reserve()
		if !ok {
			return response.Error(c, fiber.StatusServiceUnavailable, "UNAVAILABLE", "WebSocket connections unavailable")
		}
		upgraded := false
		defer func() {
			if !upgraded {
				session.Release()
			}
		}()
		c.Locals("websocket.session", session)
		c.Locals("websocket.request_id", strings.Clone(c.GetRespHeader("X-Request-ID")))
		c.Locals("websocket.user_id", claims.UserID)
		c.Locals("websocket.expiry", claims.ExpiresAt.Time)
		if err := upgrade(c); err != nil {
			return err
		}
		upgraded = true
		return nil
	})
}

func serveFiberWebSocket(conn *transport.Conn) {
	session, ok := conn.Locals("websocket.session").(*websocket.Session)
	if !ok || session == nil {
		_ = conn.Close()
		return
	}
	defer session.Release()
	requestID, _ := conn.Locals("websocket.request_id").(string)
	userID, _ := conn.Locals("websocket.user_id").(uint)
	expiry, _ := conn.Locals("websocket.expiry").(time.Time)
	socket := conn.UnderlyingConn()
	if underlying, ok := socket.(interface{ UnsafeConn() net.Conn }); ok {
		socket = underlying.UnsafeConn()
	}
	session.Serve(&fiberWebSocketConnection{Conn: conn.Conn, socket: socket}, requestID, userID, expiry)
}
