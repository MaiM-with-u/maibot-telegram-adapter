package maibot

import (
	"errors"
	"fmt"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"net/http"
	"sync"
	"time"

	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/config"
	"github.com/gorilla/websocket"
)

type MessageHandler func(*MessageBase)

type Client struct {
	conn              *websocket.Conn
	endpoint          string
	platform          string
	authToken         string
	done              chan struct{}
	reconnectInterval time.Duration
	startOnce         sync.Once
	mu                sync.Mutex
	messageHandler    MessageHandler
}

func NewClient(endpoint, platform, authToken string) *Client {
	return &Client{
		endpoint:          endpoint,
		platform:          platform,
		authToken:         authToken,
		done:              make(chan struct{}),
		reconnectInterval: 5 * time.Second,
	}
}

func (c *Client) Connect() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("platform", c.platform)
	if c.authToken != "" {
		headers.Set("authorization", c.authToken)
	}

	conn, _, err := dialer.Dial(c.endpoint, headers)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	logger.Info("Connected to WebSocket endpoint: %s", c.endpoint)

	c.setupHeartbeat()

	return nil
}

func (c *Client) setupHeartbeat() {
	c.conn.SetPongHandler(func(appData string) error {
		return nil
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.done:
				return
			case <-ticker.C:
				c.mu.Lock()
				if c.conn != nil {
					err := c.conn.WriteMessage(websocket.PingMessage, nil)
					if err != nil {
						logger.Error("Ping failed: %v", err)
						c.conn.Close()
						c.conn = nil
					}
				}
				c.mu.Unlock()
			}
		}
	}()
}

func (c *Client) SendMessage(message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return errors.New("WebSocket not connected")
	}

	err := c.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logger.Error("SendMessage failed: %v", err)
		c.conn.Close()
		c.conn = nil
		return err
	}

	return nil
}

// SendMessageBase sends a MessageBase message
func (c *Client) SendMessageBase(msg *MessageBase) error {
	jsonData, err := msg.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert message to JSON: %w", err)
	}

	return c.SendMessage([]byte(jsonData))
}

// SendTextMessage sends a simple text message
func (c *Client) SendTextMessage(messageID, userID, text string, groupID ...string) error {
	var msg *MessageBase
	if len(groupID) > 0 && groupID[0] != "" {
		msg = NewGroupTextMessage(c.platform, messageID, userID, groupID[0], text)
	} else {
		msg = NewSimpleTextMessage(c.platform, messageID, userID, text)
	}

	return c.SendMessageBase(msg)
}

// SetMessageHandler sets the message handler function
func (c *Client) SetMessageHandler(handler MessageHandler) {
	c.messageHandler = handler
}

func (c *Client) Listen() {
	for {
		select {
		case <-c.done:
			return
		default:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()

			if conn == nil {
				logger.Info("Attempting to reconnect...")
				if err := c.Connect(); err != nil {
					logger.Error("Reconnection failed: %v", err)
					time.Sleep(c.reconnectInterval)
					continue
				}
				continue
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				logger.Error("Error reading message: %v", err)
				c.mu.Lock()
				if c.conn != nil {
					c.conn.Close()
					c.conn = nil
				}
				c.mu.Unlock()
				time.Sleep(c.reconnectInterval)
				continue
			}

			logger.Info("Received message: %s", string(message))
			c.handleReceivedMessage(message)
		}
	}
}

func (c *Client) Close() {
	close(c.done)

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// handleReceivedMessage processes incoming messages
func (c *Client) handleReceivedMessage(data []byte) {
	msg, err := FromJSON(string(data))
	if err != nil {
		logger.Error("Failed to parse received message: %v", err)
		return
	}

	logger.Info("Received MessageBase from MaiBot server: %s", msg.GetTextContent())

	if c.messageHandler != nil {
		c.messageHandler(msg)
	}
}

func StartMaiBot() {
	cfg := config.Get()
	client := NewClient(cfg.MaiBot.URL, "telegram", "")

	client.startOnce.Do(func() {
		go client.Listen()
		logger.Info("MaiBot WebSocket client started with auto-reconnect")
	})
}

// GetDefaultClient returns the default client instance
var defaultClient *Client

func GetDefaultClient() *Client {
	if defaultClient == nil {
		cfg := config.Get()
		defaultClient = NewClient(cfg.MaiBot.URL, "telegram", "")

		go func() {
			if err := defaultClient.Connect(); err != nil {
				logger.Error("Failed to connect to MaiBot server: %v", err)
			}
			defaultClient.Listen()
		}()
	}
	return defaultClient
}

func InitDefaultClient() error {
	cfg := config.Get()
	defaultClient = NewClient(cfg.MaiBot.URL, "telegram", "")

	if err := defaultClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to MaiBot server: %w", err)
	}

	go defaultClient.Listen()
	return nil
}
