package game

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GameManager struct {
	games map[string]*Game
	mu    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*Game),
	}
}

func (m *GameManager) CreateGame() (*Game, error) {
	gameID := uuid.New().String()
	game := NewGame(gameID)

	m.mu.Lock()
	m.games[gameID] = game
	m.mu.Unlock()

	return game, nil
}

func (m *GameManager) GetGame(id string) (*Game, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	game, exists := m.games[id]
	if !exists {
		return nil, ErrGameNotFound
	}

	return game, nil
}

func (m *GameManager) StartGame(gameID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	if !game.IsGameReady() {
		return ErrNotEnoughPlayers
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.Phase != PhaseWaiting {
		return ErrGameAlreadyStarted
	}

	// Assign roles
	players := make([]*Player, 0, len(game.Players))
	for _, p := range game.Players {
		log.Printf("Adding player to role assignment: %s", p.Name)
		players = append(players, p)
	}

	// Shuffle players
	for i := len(players) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()) % (i + 1)
		players[i], players[j] = players[j], players[i]
	}

	// Assign mafias
	mafiaCount := game.MafiaCount
	if mafiaCount > len(players)/3 {
		mafiaCount = len(players) / 3
	}
	for i := 0; i < mafiaCount; i++ {
		players[i].Role = RoleMafia
		log.Printf("Assigned Mafia role to: %s", players[i].Name)
	}

	// Assign special roles (detective and medic)
	currentIndex := mafiaCount
	if len(players) >= 5 {
		players[currentIndex].Role = RoleDetective
		log.Printf("Assigned Detective role to: %s", players[currentIndex].Name)
		currentIndex++

		if len(players) >= 7 {
			players[currentIndex].Role = RoleMedic
			log.Printf("Assigned Medic role to: %s", players[currentIndex].Name)
			currentIndex++
		}
	}

	// Assign remaining players as villagers
	for i := currentIndex; i < len(players); i++ {
		players[i].Role = RoleVillager
		log.Printf("Assigned Villager role to: %s", players[i].Name)
	}

	game.Phase = PhaseNight
	game.Round = 1
	game.PhaseEndTime = time.Now().Add(30 * time.Second)

	// Log final role assignments
	log.Printf("Final role assignments:")
	for _, p := range players {
		log.Printf("Player %s - Role: %s", p.Name, p.Role)
	}

	return nil
}

func (m *GameManager) HandleVote(gameID string, voterID string, targetID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.Phase != PhaseVote {
		return ErrInvalidPhase
	}

	voter, exists := game.Players[voterID]
	if !exists {
		return ErrPlayerNotFound
	}

	if !voter.IsAlive {
		return ErrPlayerNotAlive
	}

	target, exists := game.Players[targetID]
	if !exists {
		return ErrPlayerNotFound
	}

	if !target.IsAlive {
		return ErrPlayerNotAlive
	}

	voter.VotedFor = targetID
	return nil
}

func (m *GameManager) ProcessVotes(gameID string) (string, error) {
	game, err := m.GetGame(gameID)
	if err != nil {
		return "", err
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	votes := make(map[string]int)
	for _, player := range game.Players {
		if player.IsAlive && player.VotedFor != "" {
			votes[player.VotedFor]++
		}
	}

	maxVotes := 0
	var eliminated string
	for playerID, voteCount := range votes {
		if voteCount > maxVotes {
			maxVotes = voteCount
			eliminated = playerID
		}
	}

	if eliminated != "" {
		game.Players[eliminated].IsAlive = false
	}

	// Reset votes
	for _, player := range game.Players {
		player.VotedFor = ""
	}

	return eliminated, nil
}

func (m *GameManager) RemoveGame(id string) {
	m.mu.Lock()
	delete(m.games, id)
	m.mu.Unlock()
}

// HandleMafiaAction records a mafia member's target for the night
func (m *GameManager) HandleMafiaAction(gameID string, mafiaID string, targetID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.Phase != PhaseNight {
		return ErrInvalidPhase
	}

	mafia, exists := game.Players[mafiaID]
	if !exists {
		return ErrPlayerNotFound
	}

	if mafia.Role != RoleMafia {
		return ErrNotMafia
	}

	target, exists := game.Players[targetID]
	if !exists {
		return ErrPlayerNotFound
	}

	if !target.IsAlive {
		return ErrPlayerNotAlive
	}

	// Record the mafia's vote
	mafia.VotedFor = targetID
	log.Printf("Mafia %s voted to kill %s", mafia.Name, target.Name)

	return nil
}

// ProcessNightActions processes all night actions (mafia kills, detective investigations, medic saves)
func (m *GameManager) ProcessNightActions(gameID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Count mafia votes
	votes := make(map[string]int)
	var mafiaMembers int
	for _, player := range game.Players {
		if player.IsAlive && player.Role == RoleMafia {
			mafiaMembers++
			if player.VotedFor != "" {
				votes[player.VotedFor]++
			}
		}
	}

	// Find the target with the most votes
	var targetID string
	maxVotes := 0
	for id, count := range votes {
		if count > maxVotes {
			maxVotes = count
			targetID = id
		}
	}

	// Only kill if at least half of living mafia voted for the same target
	if maxVotes >= (mafiaMembers+1)/2 && targetID != "" {
		target := game.Players[targetID]
		target.IsAlive = false
		log.Printf("Mafia killed player %s", target.Name)
	}

	// Reset all votes
	for _, player := range game.Players {
		player.VotedFor = ""
	}

	return nil
}

func (m *GameManager) AdvancePhase(gameID string) error {
	game, err := m.GetGame(gameID)
	if err != nil {
		return err
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Define phase durations
	const (
		NightDuration   = 30 * time.Second
		DiscussDuration = 120 * time.Second
		VoteDuration    = 30 * time.Second
	)

	// Transition to next phase
	switch game.Phase {
	case PhaseNight:
		// Process night actions before moving to discussion
		if err := m.ProcessNightActions(gameID); err != nil {
			return err
		}
		game.Phase = PhaseDiscuss
		game.PhaseEndTime = time.Now().Add(DiscussDuration)
		log.Printf("Game %s: Night phase ended, moving to Discussion phase", gameID)

	case PhaseDiscuss:
		game.Phase = PhaseVote
		game.PhaseEndTime = time.Now().Add(VoteDuration)
		log.Printf("Game %s: Discussion phase ended, moving to Voting phase", gameID)

	case PhaseVote:
		// Process votes and eliminate player
		eliminated, err := m.ProcessVotes(gameID)
		if err != nil {
			return err
		}

		if eliminated != "" {
			log.Printf("Game %s: Player %s was eliminated", gameID, game.Players[eliminated].Name)
		}

		// Check win conditions
		mafiaCount := 0
		villagerCount := 0
		for _, p := range game.Players {
			if !p.IsAlive {
				continue
			}
			if p.Role == RoleMafia {
				mafiaCount++
			} else {
				villagerCount++
			}
		}

		// Check if game is over
		if mafiaCount == 0 {
			game.Phase = PhaseGameOver
			log.Printf("Game %s: Villagers win!", gameID)
			return nil
		} else if mafiaCount >= villagerCount {
			game.Phase = PhaseGameOver
			log.Printf("Game %s: Mafia wins!", gameID)
			return nil
		}

		// If game isn't over, start next night phase
		game.Phase = PhaseNight
		game.Round++
		game.PhaseEndTime = time.Now().Add(NightDuration)
		log.Printf("Game %s: Starting night phase of round %d", gameID, game.Round)

	default:
		return ErrInvalidPhase
	}

	return nil
}
