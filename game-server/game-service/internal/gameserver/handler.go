package gameserver

import "github.com/gin-gonic/gin"

/**
* Handles all the management and maintenance of connections with client
**/

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleWebSocketConnection(c *gin.Context) {
	// TODO: call game server's channel
}
