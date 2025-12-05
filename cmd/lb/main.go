package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	URL   string
	Alive bool
	mux   sync.RWMutex
}

func (s *Server) SetAlive(alive bool) {
	s.mux.Lock()
	s.Alive = alive
	s.mux.Unlock()
}

func (s *Server) IsAlive() bool {
	s.mux.RLock()
	alive := s.Alive
	s.mux.RUnlock()
	return alive
}

var (
	servers = []*Server{
		{URL: "localhost:8000", Alive: true},
		{URL: "localhost:8001", Alive: true},
		{URL: "localhost:8002", Alive: true},
	}
	counter uint64
)

func healthCheck() {
	t := time.NewTicker(10 * time.Second)

	for {
		<-t.C
		log.Println("Starting server health check...")

		for _, server := range servers {
			conn, err := net.DialTimeout("tcp", server.URL, 2*time.Second)

			if err != nil {
				log.Printf("Server %s is down: %v", server.URL, err)
				server.SetAlive(false)
			} else {
				server.SetAlive(true)
				conn.Close()
			}
		}
		log.Println("Health Check Completed")
	}
}

func chooseBackend() *Server {
	start := atomic.AddUint64(&counter, 1)

	len := uint64(len(servers))
	for i := 0; i < int(len); i++ {
		index := (start + uint64(i)) % len

		if servers[index].Alive {
			return servers[index]
		}
	}
	return nil
}

func proxy(clientConn net.Conn, backendConn net.Conn) {
	defer clientConn.Close()
	defer backendConn.Close()
	done := make(chan struct{})

	go func() {
		_, err := io.Copy(backendConn, clientConn)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Printf("Error copying files from client to backend, %v", err)
		}

		done <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log.Printf("Error copying files from backend to client, %v", err)
		}

		done <- struct{}{}
	}()

	<-done
}

func handleConnection(clientConn net.Conn) {
	backend := chooseBackend()
	if backend == nil {
		log.Printf("Error: All servers are down!")
		return
	}
	log.Printf("Balancing connection to: %s", backend.URL)
	backendConn, err := net.Dial("tcp", backend.URL)
	if err != nil {
		log.Printf("Failed to dial %s : %v", backend.URL, err)
		clientConn.Close()
		return
	}
	proxy(clientConn, backendConn)
}

func main() {
	PORT := "8080"
	ADDR := ":" + PORT
	listener, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Fatalf("Failed to run the load balancer, ERROR: %v", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on PORT: ", PORT)

	for {
		go healthCheck()

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}
