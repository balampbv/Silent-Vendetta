package websocket

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	Conn     *websocket.Conn
	GameID   string
	PlayerID string
}

type Message struct {
	Type     string      `json:"type"`
	GameID   string      `json:"gameId"`
	PlayerID string      `json:"playerId"`
	Data     interface{} `json:"data"`
}

type Manager struct {
	clients    map[*Client]bool
	Broadcast  chan Message
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (m *Manager) Start() {
	for {
		select {
		case client := <-m.Register:
			m.mu.Lock()
			m.clients[client] = true
			m.mu.Unlock()

		case client := <-m.Unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Conn.Close()
			}
			m.mu.Unlock()

		case message := <-m.Broadcast:
			m.mu.RLock()
			for client := range m.clients {
				if client.GameID == message.GameID {
					err := client.Conn.WriteJSON(message)
					if err != nil {
						log.Printf("error: %v", err)
						client.Conn.Close()
						delete(m.clients, client)
					}
				}
			}
			m.mu.RUnlock()
		}
	}
}

func (m *Manager) SendToGame(gameID string, message Message) {
	m.Broadcast <- message
}

func (m *Manager) SendToPlayer(playerID string, message Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for client := range m.clients {
		if client.PlayerID == playerID {
			err := client.Conn.WriteJSON(message)
			if err != nil {
				log.Printf("error sending to player: %v", err)
			}
			return
		}
	}
}

// GetGameClients returns all clients in a specific game
func (m *Manager) GetGameClients(gameID string) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var gameClients []*Client
	for client := range m.clients {
		if client.GameID == gameID {
			gameClients = append(gameClients, client)
		}
	}
	return gameClients
}
