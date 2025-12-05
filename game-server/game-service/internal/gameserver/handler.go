package gameserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

/**
* Handles all the management and maintenance of connections with client
**/

func (s *Server) HandleWebSocketConnection(c *gin.Context) {
	// get token
	token := c.Query("token")
	name := c.Query("name")
	// if token == "" {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
	// }
	// DOTO: get user from db
	// 暫時帶入token當user id
	uuid := uuid.MustParse(token)
	player := &types.Player{
		ID:       uuid,
		Username: name,
	}
	// TODO verify JWT，get player
	// player, err := s.validateJWT(token)
	// if err != nil {
	// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	// 		return
	// }

	// TODO: call game server's channel
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Error establishing websocket connection.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade connection"})
		return
	}

	s.MapConnToPlayer(conn, *player)
	// handle each connected client's messages concurrently
	go s.ServeConnectedPlayer(conn)

}

/**
* Serves each individual connected player.
**/
func (s *Server) ServeConnectedPlayer(conn *websocket.Conn) {
	// removes client and closes connection
	defer func() {
		fmt.Println("Connection closed due to end of function.")
		// TODO: clean up player / client
		// s.cleanUpClient(conn)
	}()

	for {
		fmt.Println("Listening for user messages...")
		_, message, err := conn.ReadMessage()

		fmt.Printf("\nMessage received from connected user: %s\n\n", string(message))

		// --- clean up connection ---
		if err != nil {
			// Unexpected Error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// TODO: Add player name here
				fmt.Printf("Abormal error occured with player %v. Closing connection.\n", "Add player name here")
				break
			}

			// Close Error
			if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				fmt.Printf("Error on close, close going away, error: %s\n", err)
				break
			}

			// General Error
			fmt.Printf("General error occured during connection: %s\n", err)
			break
		}

		fmt.Println("before decoding received message")

		// --- Client Connection Handling ---
		// Decodes Incoming client message and serves their unique connection its own goroutine

		// decode message to pre-defined json structure "GameMessage"
		var decodedMsg types.Message

		err = json.Unmarshal(message, &decodedMsg)

		if err != nil {
			fmt.Println("Error when decoding payload.")
			conn.WriteJSON(types.Message{Action: "Error", Payload: "Your message to server was the incorrect format and could not be decoded as JSON."})
			continue
		}

		// handle concurrent writes back to clients
		s.setupClientWriter(conn)

		clientPackage := types.ClientPackage{Message: decodedMsg, Conn: conn}

		fmt.Println("Sending clientPackage to message hub.")

		// send message to MessageHub via an *unbuffered channel* for handling based on the type field.
		s.serverChan <- clientPackage
	}
}

/**
* Handles adding clients and creating gameMsgChans for handling connection writes
* back to the connected client.
*
* NOTE: Gorilla Websocket package only allows ONE CONCURRENT WRITER
* at a time, meaning its best to utilize *unbuffered* channels to prevent
* a single client from locking the entire server, and prevent race conditions
* where multiple writes to the same connection.
**/
func (s *Server) setupClientWriter(conn *websocket.Conn) {
	// sets up this connection's personal game message channel
	s.createMsgChan(conn)

	// in the case the channel exists
	msgChan, err := s.getGameMsgChan(conn)

	if err != nil {
		fmt.Println(err)

		// TODO: CLEAN UP CLIENT
		// s.cleanUpClient(conn)
		return
	}

	// concurrently listen to all incoming messages over the channel to write game actions
	// back to the client
	go func() {
		// reading from unbuffered channel to prevent more than one write
		// a time from ANY single connection
		for msg := range msgChan {
			err := conn.WriteJSON(msg)

			if err != nil {
				// TODO: remove connection from channel and close, clean up client
				// s.cleanUpClient(conn)
				break
			}
		}
	}()

}

/**
* Creates the unique game message channel for a specific connection for writing back
* from server to client.
**/
func (s *Server) createMsgChan(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.msgChan[conn] = make(chan types.Message)
}

/**
* Gets the unique game message channel for a specific connection for writing back
* from server to client, validating that it exists.
**/
func (s *Server) getGameMsgChan(conn *websocket.Conn) (chan types.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.msgChan[conn]

	if !exists {
		return nil, fmt.Errorf("Game message channel for this connection does not exist.")
	}

	return channel, nil
}
