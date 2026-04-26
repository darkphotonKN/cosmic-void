package gameserver

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/queue"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Core concurrent message orchestrator.
**/

type messageHub struct {
	sessionManager SessionManager
	gameSessionCh  chan types.Message
	sessions       map[string]*game.Session
	mu             sync.RWMutex
	sender         *messaging.MessageSender
}

type SessionManager interface {
	CreateGameSession(players []*types.Player) *game.Session
	GetGameSession(id uuid.UUID) (*game.Session, bool)
	GetServerChan() chan types.ClientPackage
	AddPlayer(*types.Player) error
	GetPlayerFromConn(conn *websocket.Conn) (*types.Player, bool)
	GetMatchedChan() chan []*types.Player
	GetQueueStatusChan() chan queue.QueueStatus
}

func NewMessageHub(sessionManager SessionManager, sender *messaging.MessageSender) *messageHub {
	return &messageHub{
		sessionManager: sessionManager,
		sessions:       make(map[string]*game.Session),
		sender:         sender,
	}
}

/**
* Core goroutine hub to handle all incoming messages and orchestrate them
* to other parts of game.
**/
func (h *messageHub) Run() {
	slog.Info("Initializing message hub.")

	for {
		select {
		case clientPackage := <-h.sessionManager.GetServerChan():
			slog.Info("incoming message.", "message", clientPackage.Message)

			// handle message based on action
			var gameActions map[constants.Action]bool = map[constants.Action]bool{
				constants.ActionMove:     true,
				constants.ActionAttack:   true,
				constants.ActionInteract: true,
				constants.ActionEquip:    true,
				constants.ActionUnequip:  true,
			}

			messageAction := constants.Action(clientPackage.Message.Action)

			// --- GAME RELATED ACTIONS ---
			// any message sent from the client after a game session is initialized
			// will be propogated from the message hub to corresponding server.

			if gameActions[messageAction] {
				sessionID, err := clientPackage.Message.GetSessionID()

				slog.Debug("debug sessionID clientPackage GetSessionID", "sessionID", sessionID)

				if err != nil {
					err := "invalid or missing session ID in payload"
					h.sender.SendMessageToConn(clientPackage.Conn, types.Message{
						Action: clientPackage.Message.Action,
						Payload: map[string]interface{}{
							"message": "Invalid or missing session ID in payload",
						},
						Error: &err,
					})
					continue
				}

				session, exists := h.sessionManager.GetGameSession(sessionID)

				if !exists {
					err := "Game session not found"
					h.sender.SendMessageToConn(clientPackage.Conn, types.Message{
						Action: clientPackage.Message.Action,
						Payload: map[string]interface{}{
							"message": fmt.Sprintf("Game session not found for session ID: %s", sessionID),
						},
						Error: &err,
					})
					slog.Error("Game doesn't exist for this player", "message", clientPackage.Message)
					continue
				}

				// propogate message to corresponding game
				session.MessageCh <- clientPackage
				continue
			}

			// --- MENU RELATED ACTIONS ---
			// These actions will be actions for before game initialization happens.
			switch messageAction {

			// NOTE: queues a player for a game
			case constants.ActionFindGame:
				slog.Debug("ActionFindGame")
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)

				// player doesn't exist at all in the server, skip them
				if !exists {
					slog.Debug("Player doesn't exist in session.")
					continue
				}

				// -- player already exists in an old game --
				err := h.handlePlayerExistingGame(player, clientPackage)

				// no error, so player exists, skip queue
				if err == nil {
					slog.Debug("player exists alreayd, skipping queue",
						"player_id", player.ID,
						"player_username", player.Username,
					)
					continue
				}

				// -- queue up player --
				err = h.sessionManager.AddPlayer(player)
				if err != nil {
					queueErr := err.Error()
					message := "Error occured when attempting to queue player"

					// broadcast error back to client attempting to queue

					if errors.Is(err, game.ErrPlayerAlreadyInQueue) {
						message = "Player attempted to queue twice."
					}

					h.sender.SendMessageToConn(clientPackage.Conn, types.Message{
						Action: clientPackage.Message.Action,
						Payload: map[string]interface{}{
							"message":   message,
							"player_id": player.ID.String(),
							"username":  player.Username,
						},
						Error: &queueErr,
					})
					continue
				}

				slog.Info("Player added to matchmaking queue", "player username", player.Username)

				h.sender.SendMessageToConn(clientPackage.Conn, types.Message{
					Action: clientPackage.Message.Action,
					Payload: map[string]interface{}{
						"message":   "Successfully joined matchmaking queue",
						"player_id": player.ID.String(),
						"username":  player.Username,
					},
				})

			case constants.ActionLeaveQueue:
				player, exists := h.sessionManager.GetPlayerFromConn(clientPackage.Conn)
				if !exists {
					h.sender.SendMessageToPlayer(player.ID, types.Message{
						Action: clientPackage.Message.Action,
						Payload: map[string]interface{}{
							"message":   "Player not found for connection",
							"player_id": player.ID.String(),
						},
					})
					slog.Error("Player not found for connection",
						"player_id", player.ID,
					)
					continue
				}

				// h.sessionManager.RemovePlayerFromQueue(player)
				slog.Debug("Player leaving game",
					"player_id", player.ID,
				)

				h.sender.SendMessageToPlayer(player.ID, types.Message{
					Action: clientPackage.Message.Action,
					Payload: map[string]interface{}{
						"message":   "Successfully left the queue",
						"player_id": player.ID.String(),
					},
				})

			default:
				err := "Unknown action"
				h.sender.SendMessageToConn(
					clientPackage.Conn, types.Message{
						Action: clientPackage.Message.Action,
						Payload: map[string]interface{}{
							"message": err,
						},
						Error: &err,
					},
				)
			}

		case matchedPlayers := <-h.sessionManager.GetMatchedChan():
			fmt.Printf("Received matched players, creating game session...\n")
			fmt.Println(matchedPlayers)
			session := h.sessionManager.CreateGameSession(matchedPlayers)
			playerIDs := make([]uuid.UUID, len(matchedPlayers))
			for i, player := range matchedPlayers {
				playerIDs[i] = player.ID
			}
			h.sender.BroadcastToPlayerList(playerIDs,
				types.Message{
					Action: "game_found",
					Payload: map[string]any{
						"session_id": session.ID.String(),
					},
				})

		case status := <-h.sessionManager.GetQueueStatusChan():
			fmt.Printf("Queue status update: %d/%d\n", status.Current, status.Total)
			playerIDs := make([]uuid.UUID, len(status.Players))
			for i, player := range status.Players {
				playerIDs[i] = player.ID
			}
			h.sender.BroadcastToPlayerList(playerIDs,
				types.Message{
					Action: "queue_status",
					Payload: map[string]any{
						"current": status.Current,
						"total":   status.Total,
					},
				})
		}
	}
}

/**
* Checks if a player exists in a game and handles the responses to the client if they
* exist, or throws an error if they dont.
**/
func (h *messageHub) handlePlayerExistingGame(player *types.Player, clientPackage types.ClientPackage) error {
	if player.CurrentGameSessionId != uuid.Nil {
		slog.Warn("Attempting to find a game when player already in an old session. Attempting to resume.")

		// find the session with their current game sessionId
		session, exists := h.sessionManager.GetGameSession(player.CurrentGameSessionId)

		if !exists {
			slog.Error("When attempting to resume game for player detected that game session doesn't exist anymore", "playerId", player.ID, "sessionId", player.CurrentGameSessionId)
			// clear the non-existing session
			player.CurrentGameSessionId = uuid.Nil
			return commonconstants.ErrGameDoesntExist
		}

		// game found, tells frontend to resume, player should be already receiving game state at this point
		slog.Debug("Resuamble session found", "sessionId", session.ID)

		slog.Info("Player already in session, sending game_found",
			"player_id", player.ID,
			"current_game_session_id", player.CurrentGameSessionId)

		h.sender.SendMessageToConn(clientPackage.Conn, types.Message{
			Action: "game_found",
			Payload: map[string]any{
				"session_id": player.CurrentGameSessionId.String(),
			},
		})

		// return no error if player exists in a game
		return nil
	}

	slog.Debug("Player doesn't exist in any game.")
	return commonconstants.ErrGameDoesntExist
}
