package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
)

var (
	servers = []string{
		"localhost:8000",
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}

	counter uint64 = 0
)

func chooseBackend() string {
	current := atomic.AddUint64(&counter, 1)

	len := uint64(len(servers))
	index := current % len
	return servers[index]
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
	backendAddr := chooseBackend()
	log.Printf("Balancing connection to: %s", backendAddr)

	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to dial %s : %v", backendAddr, err)
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
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}
