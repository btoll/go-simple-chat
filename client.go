package main

import (
	"bufio"
	"context"
	"net"
	"sync"
)

type Username string

type Client struct {
	conn net.Conn
	user Username
	ch   chan string
	once sync.Once
}

func NewClient(conn net.Conn, username Username) *Client {
	c := &Client{
		conn: conn,
		user: username,
		ch:   make(chan string, clientBuffer),
	}
	register(c)
	return c
}

func (c *Client) Broadcast(ctx context.Context) {
	scanner := bufio.NewScanner(c.conn)
ScannerLoop:
	for scanner.Scan() {
		select {
		case chat.broadcast <- ChatMessage{
			sender: c,
			msg:    scanner.Text(),
		}:
		case <-ctx.Done():
			break ScannerLoop
		default:
			// The broadcast channel is full so drop the message.
			// Not great, but it "gracefully" handles backpressure
			// and avoids a panic during a graceful shutdown.
			//
			// Without the default, this will block trying to send
			// when the buffered broadcast channel is at capacity.
			// Then, when gracefully shutting down, the broadcast
			// channel is closed which will cause this to panic
			// b/c at that point we'd be attempting to send on a
			// closed channel.
		}
	}
	// If we got here, the client disconnected and we need to remove
	// them from the connection map.
	deregister(c)
}

func (c *Client) Listen(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.ch:
			if !ok {
				return
			}
			_, err := c.conn.Write([]byte(msg))
			if err != nil {
				// We're assuming here that the reason it failed is b/c
				// it's a dead connection.
				deregister(c)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Start(ctx context.Context) {
	go c.Broadcast(ctx)
	go c.Listen(ctx)
}
