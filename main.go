package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quit       chan struct{}
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quit:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)

	if err != nil {
		return err
	}

	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()

	fmt.Printf("Server started on %s\n", s.listenAddr)

	<-s.quit

	return nil
}

func (s *Server) Stop() {
	close(s.quit)
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				fmt.Println("Accept connection error:", err)
				continue
			}
		}

		fmt.Printf("New connection: %s\n", conn.RemoteAddr().String())

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Printf("Received from %v: %s\n", conn.RemoteAddr(), msg)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Read error: %v\n", err)
	}

	fmt.Printf("Connection closed: %s\n", conn.RemoteAddr())
}

func main() {
	server := NewServer(":3000")

	go func() {
		server.Start()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	fmt.Println("Shutting down gracefully...")

	server.Stop()

	fmt.Println("Server stopped.")
}
