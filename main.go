package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type ChatMessage struct {
	sender *Client
	msg    string
}
type ChatRoom map[net.Conn]*Client
type Chat struct {
	c         ChatRoom
	broadcast chan ChatMessage
	m         *sync.Mutex
}

func NewChat() *Chat {
	return &Chat{
		c:         make(ChatRoom),
		broadcast: make(chan ChatMessage, 32),
		m:         &sync.Mutex{},
	}
}

var chat = NewChat()

func handleNewConnection(conn net.Conn) {
	_, err := conn.Write([]byte("What's your name?: "))
	if err != nil {
		// TODO
		fmt.Printf("err=%+v\n", err)
	}

	b := make([]byte, 64)
	n, err := conn.Read(b)
	if err != nil {
		// This will handle a SIGINT at first connection and before a name is entered at the prompt.
		if errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Error: %s from connection %s\n", err.Error(), conn.RemoteAddr())
			return
		}
		// TODO
	}

	c := NewClient(conn, Username(string(b[:n-1])))
	conn.Write([]byte(
		fmt.Sprintf("Welcome to the the simple chat server, %s!\n", c.user),
	))
	go c.Broadcast()
	go c.Listen()
}

func deregister(client *Client) {
	// Maybe want different reasons as to why a connection was deleted.
	chat.m.Lock()
	if _, found := chat.c[client.conn]; found {
		close(client.ch)
		delete(chat.c, client.conn)
		fmt.Printf("User %s just left the chat!\n", client.user)
	}
	chat.m.Unlock()
}

func register(client *Client) {
	chat.m.Lock()
	chat.c[client.conn] = client
	fmt.Printf("User %s just joined the chat!\n", client.user)
	chat.m.Unlock()
}

func main() {
	// Not necessary but good hygiene.
	defer close(chat.broadcast)

	listener, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Simple chat server started and listening on :9999")

	go func() {
		for {
			select {
			case message := <-chat.broadcast:
				chat.m.Lock()
				for _, client := range chat.c {
					if client != message.sender {
						client.ch <- fmt.Sprintf("(%s) %s\n", message.sender.user, message.msg)
					}
				}
				chat.m.Unlock()
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleNewConnection(conn)
	}
}
