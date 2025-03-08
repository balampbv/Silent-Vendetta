package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberWs "github.com/gofiber/websocket/v2"
	"github.com/silent-vendetta/pkg/game"
	"github.com/silent-vendetta/pkg/websocket"
)

type CreateGameRequest struct {
	PlayerName string `json:"playerName"`
}

type JoinGameRequest struct {
	PlayerName string `json:"playerName"`
}

func main() {
	app := fiber.New()

	// Enable CORS
	app.Use(cors.New())

	// Initialize game manager and websocket manager
	gameManager := game.NewGameManager()
	wsManager := websocket.NewManager()
	go wsManager.Start()

	// API routes
	app.Post("/api/games", func(c *fiber.Ctx) error {
		var req CreateGameRequest
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		game, err := gameManager.CreateGame()
		if err != nil {
			return err
		}

		// Add the host player
		if err := game.AddPlayer(req.PlayerName, req.PlayerName); err != nil {
			return err
		}

		// Send initial player count
		wsManager.SendToGame(game.ID, websocket.Message{
			Type: "playerCount",
			Data: len(game.Players),
		})

		return c.JSON(fiber.Map{
			"gameId": game.ID,
		})
	})

	app.Post("/api/games/:id/join", func(c *fiber.Ctx) error {
		gameID := c.Params("id")
		log.Printf("Join game request received for game ID: %s", gameID)

		var req JoinGameRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("Error parsing join request: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}
		log.Printf("Player %s attempting to join game", req.PlayerName)

		game, err := gameManager.GetGame(gameID)
		if err != nil {
			log.Printf("Error getting game: %v", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Game not found",
			})
		}
		log.Printf("Found game with %d players", len(game.Players))

		if err := game.AddPlayer(req.PlayerName, req.PlayerName); err != nil {
			log.Printf("Error adding player: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		log.Printf("Player %s successfully joined game. Total players: %d", req.PlayerName, len(game.Players))

		// Broadcast updated game state and player count
		wsManager.SendToGame(gameID, websocket.Message{
			Type: "gameState",
			Data: game,
		})
		wsManager.SendToGame(gameID, websocket.Message{
			Type: "playerCount",
			Data: len(game.Players),
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Successfully joined game",
		})
	})

	app.Post("/api/games/:id/start", func(c *fiber.Ctx) error {
		gameID := c.Params("id")
		log.Printf("Starting game %s", gameID)
		if err := gameManager.StartGame(gameID); err != nil {
			log.Printf("Error starting game: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Get and log game state after starting
		game, _ := gameManager.GetGame(gameID)
		log.Printf("Game started successfully. Players and roles:")
		for id, player := range game.Players {
			log.Printf("Player %s (%s) - Role: %s", player.Name, id, player.Role)
		}

		// Broadcast game state to all players
		wsManager.SendToGame(gameID, websocket.Message{
			Type: "gameState",
			Data: game,
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Game started successfully",
		})
	})

	// Add new endpoint for advancing phase
	app.Post("/api/games/:id/next-phase", func(c *fiber.Ctx) error {
		gameID := c.Params("id")
		log.Printf("Advancing phase for game %s", gameID)
		if err := gameManager.AdvancePhase(gameID); err != nil {
			log.Printf("Error advancing phase: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Get updated game state to broadcast to all players
		game, err := gameManager.GetGame(gameID)
		if err != nil {
			log.Printf("Error getting game: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error getting game state",
			})
		}

		// Broadcast updated game state to all players
		wsManager.SendToGame(gameID, websocket.Message{
			Type: "gameState",
			Data: game,
		})

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Game phase advanced successfully",
		})
	})

	// WebSocket setup
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberWs.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:gameId", fiberWs.New(func(c *fiberWs.Conn) {
		gameID := c.Params("gameId")
		log.Printf("WebSocket connection established for game ID: %s", gameID)

		// Create new client
		client := &websocket.Client{
			Conn:   c,
			GameID: gameID,
		}

		// Register client
		wsManager.Register <- client

		// Send initial game state and player count
		game, err := gameManager.GetGame(gameID)
		if err == nil {
			log.Printf("Sending initial game state. Players count: %d", len(game.Players))
			client.Conn.WriteJSON(websocket.Message{
				Type: "gameState",
				Data: game,
			})
			client.Conn.WriteJSON(websocket.Message{
				Type: "playerCount",
				Data: len(game.Players),
			})
		}

		defer func() {
			wsManager.Unregister <- client
			// When a client disconnects, update player count
			if game, err := gameManager.GetGame(gameID); err == nil {
				wsManager.SendToGame(gameID, websocket.Message{
					Type: "playerCount",
					Data: len(game.Players),
				})
			}
			c.Close()
		}()

		for {
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				if fiberWs.IsUnexpectedCloseError(err, fiberWs.CloseGoingAway, fiberWs.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}

			var message websocket.Message
			if err := json.Unmarshal(msg, &message); err != nil {
				log.Printf("error unmarshaling message: %v", err)
				continue
			}

			log.Printf("Received message type: %s", message.Type)

			switch message.Type {
			case "join":
				if joinData, ok := message.Data.(map[string]interface{}); ok {
					playerName := joinData["playerName"].(string)
					client.PlayerID = playerName
					log.Printf("Player %s joined game %s", playerName, gameID)
				}
			case "mafiaAction":
				if err := gameManager.HandleMafiaAction(gameID, client.PlayerID, message.Data.(string)); err != nil {
					client.Conn.WriteJSON(websocket.Message{
						Type: "error",
						Data: err.Error(),
					})
					continue
				}

				// Get current game state
				game, _ := gameManager.GetGame(gameID)
				if game != nil {
					// Notify other mafia members about the vote
					for _, c := range wsManager.GetGameClients(gameID) {
						if p, exists := game.Players[c.PlayerID]; exists && p.Role == "mafia" {
							c.Conn.WriteJSON(websocket.Message{
								Type: "mafiaVote",
								Data: map[string]string{
									"voter":  client.PlayerID,
									"target": message.Data.(string),
								},
							})
						}
					}

					// Check if all mafia members have voted
					mafiaVotes := 0
					mafiaCount := 0
					for _, player := range game.Players {
						if player.IsAlive && player.Role == "mafia" {
							mafiaCount++
							if player.VotedFor != "" {
								mafiaVotes++
							}
						}
					}

					// If all mafia members have voted, automatically advance to next phase
					if mafiaVotes >= (mafiaCount+1)/2 {
						log.Printf("All mafia members have voted, advancing phase")
						if err := gameManager.AdvancePhase(gameID); err != nil {
							log.Printf("Error advancing phase: %v", err)
							continue
						}

						// Get updated game state
						updatedGame, _ := gameManager.GetGame(gameID)
						if updatedGame != nil {
							// Broadcast updated game state to all players
							wsManager.SendToGame(gameID, websocket.Message{
								Type: "gameState",
								Data: updatedGame,
							})
						}
					}
				}
			case "vote":
				if err := gameManager.HandleVote(gameID, client.PlayerID, message.Data.(string)); err != nil {
					client.Conn.WriteJSON(websocket.Message{
						Type: "error",
						Data: err.Error(),
					})
					continue
				}
			case "chat":
				wsManager.SendToGame(gameID, message)
			}

			if err := c.WriteMessage(messageType, msg); err != nil {
				log.Printf("write error: %v", err)
				return
			}
		}
	}))

	log.Fatal(app.Listen(":3001"))
}
