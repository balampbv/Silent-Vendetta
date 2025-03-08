package game

import "errors"

var (
	ErrGameFull            = errors.New("game is full")
	ErrGameNotFound        = errors.New("game not found")
	ErrPlayerNotFound      = errors.New("player not found")
	ErrInvalidPhase        = errors.New("invalid game phase")
	ErrNotEnoughPlayers    = errors.New("not enough players to start game")
	ErrGameAlreadyStarted  = errors.New("game has already started")
	ErrPlayerNotAlive      = errors.New("player is not alive")
	ErrInvalidVote         = errors.New("invalid vote")
	ErrPlayerNameTaken     = errors.New("player name already taken")
	ErrPlayerAlreadyExists = errors.New("player already exists")
	ErrNotMafia            = errors.New("player is not mafia")
)
