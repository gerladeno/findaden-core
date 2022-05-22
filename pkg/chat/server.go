package chat

import (
	"context"
	"sync"
)

type Store interface {
	SaveChat(ctx context.Context, uuid1, uuid2 string) error
	GetChat(ctx context.Context, uuid1, uuid2 string) error
	GetAllChats(ctx context.Context, uuid string) ([]string, error)
	SaveMessage(ctx context.Context, m *Message) error
	LoadAllMessages(ctx context.Context, uuid1, uuid2 string) ([]*Message, error)
}

type Server struct {
	store Store
	hubs  map[string]map[string]*Hub
	mx    sync.Mutex
}

func NewServer() *Server {
	s := Server{
		hubs:  make(map[string]map[string]*Hub),
		store: fakeStore{},
	}
	return &s
}

func (s *Server) GetDialog(_ context.Context, client, target string) *Hub {
	s.mx.Lock()
	defer s.mx.Unlock()
	m, ok := s.hubs[client]
	if !ok {
		m = make(map[string]*Hub)
		s.hubs[client] = m
	}
	h, ok := m[target]
	if !ok {
		h = newHub()
		go h.run()
		m[target] = h
	}
	m, ok = s.hubs[target]
	if !ok {
		m = make(map[string]*Hub)
		s.hubs[target] = m
	}
	m[client] = h
	return h
}

func (s *Server) GetAllChats(ctx context.Context, uuid string) ([]string, error) {
	return s.store.GetAllChats(ctx, uuid)
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
