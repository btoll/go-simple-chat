package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	broadcastBuffer int
	clientBuffer    int
	port            string
	chat            *Chat = NewChat()
)

type Chat struct {
	c         ChatRoom
	broadcast chan ChatMessage
	mu        *sync.Mutex
}
type ChatMessage struct {
	sender *Client
	msg    string
}
type ChatRoom map[Username]*Client

func NewChat() *Chat {
	return &Chat{
		c:         make(ChatRoom),
		broadcast: make(chan ChatMessage, broadcastBuffer),
		mu:        &sync.Mutex{},
	}
}

func handleNewConnection(ctx context.Context, conn net.Conn) {
	_, err := conn.Write([]byte("What's your name?: "))
	if err != nil {
		fmt.Printf("err=%+v\n", err)
		conn.Close()
		return
	}

	var username Username
	for {
		select {
		case <-ctx.Done():
			conn.Close()
			return
		default:
		}
		b := make([]byte, 64)
		n, err := conn.Read(b)
		if err != nil {
			// This will handle a SIGINT at first connection and before a name is entered at the prompt.
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "Error: %s from connection %s\n", err.Error(), conn.RemoteAddr())
			}
			conn.Close()
			return
		}
		username = Username(string(b[:n-1]))
		chat.mu.Lock()
		_, found := chat.c[username]
		chat.mu.Unlock()
		if found {
			_, err = fmt.Fprintf(conn, "Username %s is already taken, choose another: ", username)
			if err != nil {
				conn.Close()
				return
			}
			continue
		}
		break
	}

	_, err = fmt.Fprintf(conn, "Welcome to the the simple chat server, %s!\n", username)
	if err != nil {
		conn.Close()
		return
	}
	c := NewClient(conn, username)
	c.Start(ctx)
}

func (c *Client) closeChannel() {
	c.once.Do(func() {
		close(c.ch)
	})
}

func deregister(client *Client) {
	// Maybe want different reasons as to why a connection was deleted.
	chat.mu.Lock()
	if _, found := chat.c[client.user]; found {
		client.conn.Close()
		client.closeChannel()
		delete(chat.c, client.user)
		fmt.Printf("User %s just left the chat!\n", client.user)
	}
	chat.mu.Unlock()
}

func register(client *Client) {
	chat.mu.Lock()
	chat.c[client.user] = client
	fmt.Printf("User %s just joined the chat!\n", client.user)
	chat.mu.Unlock()
}

func shutdown(listener net.Listener) {
	// This is here to handle shutdown gracefully:
	// 1. Close listener.
	// 2. Close each client connection.  This will
	//    trigger deregister() via Client.Broadcast().
	// 3. The WaitGroup will wait for every goroutine
	//    to finish its business.
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	<-s
	// We'll handle both signals the same.
	err := listener.Close()
	if err != nil {
		fmt.Printf("err=%+v\n", err)
	}
	q := []*Client{}
	chat.mu.Lock()
	for _, client := range chat.c {
		q = append(q, client)
	}
	chat.mu.Unlock()
	for _, client := range q {
		client.conn.Close()
	}
}

func main() {
	flag.StringVar(&port, "port", "9999", "The port the chat server listens on.")
	flag.IntVar(&broadcastBuffer, "broadcastBuffer", 32, "Size of the buffered broadcast channel.")
	flag.IntVar(&clientBuffer, "clientBuffer", 16, "Size of the buffered client channel.")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Simple chat server started and listening on :" + port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case message := <-chat.broadcast:
				q := []*Client{}
				chat.mu.Lock()
				for _, client := range chat.c {
					if client != message.sender {
						q = append(q, client)
					}
				}
				chat.mu.Unlock()
				// We're just implementing simple backpressure here by skipping the send to a channel
				// if it will block for a slow client.  Dropping an occasional message isn't the end
				// of the world.
				for _, client := range q {
					select {
					case client.ch <- fmt.Sprintf("(%s) %s\n", message.sender.user, message.msg):
					default:
					}
				}
			case <-ctx.Done():
				fmt.Println("Broadcast gorouting shutting down...")
				close(chat.broadcast)
				return
			}
		}
	}()

	go func() {
		// Note the we don't need to listen for a Done event in this loop because the goroutine will be
		// exited when the listener is closed in shutdown().
		//
		// In other words, Accept() is a blocking system call on the socket. When you close that socket
		// (via listener.Close()), the OS wakes up that blocked call and returns an error (net.ErrClosed).
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				fmt.Println(err)
				continue
			}
			go handleNewConnection(ctx, conn)
		}
	}()

	shutdown(listener)
}
