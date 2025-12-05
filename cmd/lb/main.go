package main

import (
	"fmt"
	"io"
	"load-balancer/cmd/core"
	"log"
	"net"
	"os"
	"strings"
)

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

func handleConnection(clientConn net.Conn, pool *core.ServerPool) {
	backend := pool.GetNextPeer()
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
	servers := []string{
		"localhost:8000",
		"localhost:8001",
		"localhost:8002",
	}
	PORT := "8080"
	ADDR := ":" + PORT

	pool := core.NewServerPool(servers)
	pool.StartHealthCheck()

	listener, err := net.Listen("tcp", ADDR)
	if err != nil {
		log.Fatalf("Failed to run the load balancer, ERROR: %v", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on PORT: ", PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn, pool)
	}
}
