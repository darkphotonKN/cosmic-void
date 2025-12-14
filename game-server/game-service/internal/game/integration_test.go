package game

import (
	"fmt"
	"testing"

	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// test velocity updates transform of player entity after handle move and system update cycle
func TestHandleMoveUpdatesPositionIntegration(t *testing.T) {
	session := NewSession()

	player1ID := uuid.New()
	username := "Player1"
	playerEntityID := session.AddPlayer(player1ID, username)

	// check player initial position
	playerEntity, ok := session.EntityManager.GetEntity(playerEntityID)

	if !ok {
		fmt.Printf("\nPlayerEntity doesn't exist for player playerEntityID %s\n\n", playerEntityID)
	}

	playerTransformComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !ok {
		fmt.Printf("\nPlayers Velocity Component doesn't exist for enntity ID: %s\n\n", playerEntity.ID)
	}

	component := playerTransformComponent.(*components.TransformComponent)
	fmt.Printf("\nplayerTransformCoords Initial: %+v\n\n", component)

	assert.Equal(t, float64(0), component.X)
	assert.Equal(t, float64(0), component.Y)

	// update player transform
	session.handleMove(player1ID, 0.81, 0.81)

	// account for system game loop refresh rate, but only time for 1 move
	time.Sleep(time.Millisecond * 1200)

	fmt.Printf("\nplayerTransformCoords after update: %+v\n\n", component)
	assert.Equal(t, float64(0.81), component.X)
	assert.Equal(t, float64(0.81), component.Y)
}
