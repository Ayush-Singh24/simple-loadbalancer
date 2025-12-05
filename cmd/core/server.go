package core

import "sync"

type Server struct {
	URL   string
	Alive bool
	mux   sync.RWMutex
}

func NewServer(url string) *Server {
	return &Server{
		URL:   url,
		Alive: true,
	}
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
