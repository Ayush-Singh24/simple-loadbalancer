package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	port := os.Args[1]
	fmt.Fprintf(w, "Hello from Backend Server on port: %s\n", port)
}

func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: Expected 1 arguement, but received %d\n", len(args)-1)
		fmt.Fprintf(os.Stderr, "Usage %s <PORT>\n", args[0])
		os.Exit(1)
	}

	http.HandleFunc("/", handler)

	addr := ":" + os.Args[1]
	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
