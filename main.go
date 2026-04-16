package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Username string
type ChatMessage struct {
	c      net.Conn
	sender Username
	msg    string
}
type ChatRoom map[net.Conn]Username
type Chat struct {
	c         ChatRoom
	broadcast chan ChatMessage
	m         *sync.Mutex
}

func NewChat() *Chat {
	return &Chat{
		c:         make(ChatRoom),
		broadcast: make(chan ChatMessage),
		m:         &sync.Mutex{},
	}
}

var chat = NewChat()

func handleNewConnection(conn net.Conn) {
	defer conn.Close()

	_, err := conn.Write([]byte("What's your name?: "))
	if err != nil {
		fmt.Println("got here")
		// TODO
	}

	b := make([]byte, 128)
	n, err := conn.Read(b)
	if err != nil {
		// This will handle a SIGINT at first connection and before a name is entered at the prompt.
		if errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Error: %s from connection %s\n", err.Error(), conn.RemoteAddr())
			return
		}
		// TODO
	}

	username := Username(string(b[:n-1]))
	register(conn, username)
	conn.Write([]byte(
		fmt.Sprintf("Welcome to the the simple chat server, %s!\n", username),
	))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		chat.broadcast <- ChatMessage{
			c:      conn,
			sender: username,
			msg:    scanner.Text(),
		}
	}

	// If we got here, then the client sent a SIGINT (which returns `false` to scanner.Scan()),
	// and we need to remove them from the connection map.
	chat.m.Lock()
	deregister(conn, username)
	chat.m.Unlock()
}

// This must be called with chat.m held b/c deregister is call from multiple paths.
func deregister(conn net.Conn, username Username) {
	delete(chat.c, conn)
	// Probably want to message different reasons as to why a connection was deleted.
	fmt.Printf("User %s just left the chat!\n", username)
}

func register(conn net.Conn, username Username) {
	chat.m.Lock()
	chat.c[conn] = Username(username)
	fmt.Printf("User %s just joined the chat!\n", username)
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
				for c := range chat.c {
					if c != message.c {
						// Might want to write outside of the lock in case it is
						// slow or blocks.
						_, err := c.Write([]byte(
							fmt.Sprintf("(%s) %s\n", message.sender, message.msg),
						))
						if err != nil {
							// We're assuming here that the reason it failed is b/c
							// it's a dead connection.
							deregister(c, message.sender)
						}
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
