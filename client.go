package main

import (
	"bufio"
	"net"
)

type Username string

type Client struct {
	conn net.Conn
	user Username
	ch   chan string
}

func NewClient(conn net.Conn, username Username) *Client {
	c := &Client{
		conn: conn,
		user: username,
		ch:   make(chan string, 16),
	}
	register(c)
	return c
}

func (c *Client) Broadcast() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		chat.broadcast <- ChatMessage{
			sender: c,
			msg:    scanner.Text(),
		}
	}
	// If we got here, then the client sent a SIGINT (which returns `false`
	// to scanner.Scan()), and we need to remove them from the connection map.
	deregister(c)
}

func (c *Client) Listen() {
	for msg := range c.ch {
		_, err := c.conn.Write([]byte(msg))
		if err != nil {
			// We're assuming here that the reason it failed is b/c
			// it's a dead connection.
			deregister(c)
			return
		}
	}
}
