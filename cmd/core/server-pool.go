package core

import (
	"log"
	"net"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	servers []*Server
	current uint64
}

func NewServerPool(urls []string) *ServerPool {
	pool := &ServerPool{
		servers: make([]*Server, 0),
		current: 0,
	}

	for _, url := range urls {
		pool.servers = append(pool.servers, NewServer(url))
	}
	return pool
}

func (sp *ServerPool) GetNextPeer() *Server {
	next := atomic.AddUint64(&sp.current, 1)

	l := uint64(len(sp.servers))

	for i := 0; i < int(l); i++ {
		index := (uint64(i) + next) % l

		if sp.servers[index].IsAlive() {
			if i != 0 {
				atomic.StoreUint64(&sp.current, next+uint64(i))
			}
			return sp.servers[index]
		}
	}
	return nil
}

func (sp *ServerPool) StartHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			<-ticker.C
			log.Printf("Starting health check!")
			for _, server := range sp.servers {
				conn, err := net.DialTimeout("tcp", server.URL, 2*time.Second)
				if err != nil {
					log.Printf("Server %s DOWN", server.URL)
					server.SetAlive(false)
				} else {
					server.SetAlive(true)
					conn.Close()
				}
			}
		}
	}()
}
