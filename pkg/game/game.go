package game

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Role = string

// Role constants
const (
	RoleMafia     Role = "mafia"
	RoleVillager  Role = "villager"
	RoleDetective Role = "detective"
	RoleMedic     Role = "medic"
)

type Phase string

const (
	PhaseWaiting  Phase = "waiting"
	PhaseNight    Phase = "night"
	PhaseDiscuss  Phase = "discuss"
	PhaseVote     Phase = "vote"
	PhaseGameOver Phase = "gameover"
)

type Player struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     Role   `json:"role"`
	IsAlive  bool   `json:"isAlive"`
	IsHost   bool   `json:"isHost"`
	VotedFor string `json:"votedFor,omitempty"`
}

type Game struct {
	ID           string             `json:"id"`
	Players      map[string]*Player `json:"players"`
	Phase        Phase              `json:"phase"`
	Round        int                `json:"round"`
	PhaseEndTime time.Time          `json:"phaseEndTime"`
	MinPlayers   int                `json:"minPlayers"`
	MaxPlayers   int                `json:"maxPlayers"`
	MafiaCount   int                `json:"mafiaCount"`
	mu           sync.RWMutex       `json:"-"`
}

func NewGame(id string) *Game {
	return &Game{
		ID:         id,
		Players:    make(map[string]*Player),
		Phase:      PhaseWaiting,
		Round:      0,
		MinPlayers: 4,
		MaxPlayers: 10,
		MafiaCount: 2,
	}
}

func (g *Game) AddPlayer(name, playerID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.Players) >= g.MaxPlayers {
		return ErrGameFull
	}

	// Generate a unique ID for the player if not provided
	if playerID == "" {
		playerID = uuid.New().String()
	}

	// Check if player with same name already exists
	for _, p := range g.Players {
		if p.Name == name {
			return ErrPlayerNameTaken
		}
	}

	isHost := len(g.Players) == 0
	g.Players[playerID] = &Player{
		ID:      playerID,
		Name:    name,
		IsAlive: true,
		IsHost:  isHost,
	}

	return nil
}

func (g *Game) RemovePlayer(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.Players, id)
}

func (g *Game) IsGameReady() bool {
	return len(g.Players) >= g.MinPlayers
}

func (g *Game) GetAlivePlayers() []*Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	alive := make([]*Player, 0)
	for _, p := range g.Players {
		if p.IsAlive {
			alive = append(alive, p)
		}
	}
	return alive
}

func (g *Game) CheckWinCondition() (bool, Role) {
	aliveVillains := 0
	aliveCivilians := 0

	for _, p := range g.Players {
		if !p.IsAlive {
			continue
		}
		if p.Role == RoleMafia {
			aliveVillains++
		} else {
			aliveCivilians++
		}
	}

	if aliveVillains == 0 {
		return true, RoleVillager
	}
	if aliveVillains >= aliveCivilians {
		return true, RoleMafia
	}

	return false, ""
}
