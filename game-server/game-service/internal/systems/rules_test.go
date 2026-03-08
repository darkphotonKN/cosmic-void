package systems_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/game"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/systems"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestRulesSystem_Update_EndsGameWhenOnePlayerLeft tests that the game ends
// when only one player is left alive
func TestRulesSystem_Update_EndsGameWhenOnePlayerLeft(t *testing.T) {
	rulesSystem := systems.NewRulesSystem()
	endSessionCh := make(chan bool, 1)
	deltaTime := 0.016 // 60 FPS frame time

	// Create entity manager and entities
	em := ecs.NewEntityManager()

	players := []struct {
		MemberID uuid.UUID
		Username string
	}{
		{
			MemberID: uuid.New(),
			Username: "test_player1",
		},
		{
			MemberID: uuid.New(),
			Username: "test_player2",
		},
	}

	for _, player := range players {
		game.CreatePlayerEntity(em, game.PlayerConfig{
			MemberID: player.MemberID,
			Username: player.Username,
		})
	}

	entities := em.GetAllEntities()

	// game not ended after this call, players are still alive
	rulesSystem.Update(deltaTime, entities, endSessionCh)

	select {
	case endSession := <-endSessionCh:
		t.Fatalf("should not have gotten end session, but got %v", endSession)
	default:
	}

	for idx, entity := range entities {
		if idx == 0 {
			// grab first player

			// validate in case test entity changes
			// we want health component specifically
			isPlayer := entity.HasComponent(ecs.ComponentTypePlayer)
			if !isPlayer {
				t.Fatal("Could not get player entity from pool of entities to test on.")
				return
			}

			healthComp, exists := entity.GetComponent(ecs.ComponentTypeHealth)

			if !exists {
				t.Fatal("Could not get player entity's health component to test on.")
				return
			}

			// eliminate player
			healthComp.(*components.HealthComponent).IsEliminated = true
		}
	}

	// game ended after this, only one player left
	rulesSystem.Update(deltaTime, entities, endSessionCh)

	select {
	case endSession := <-endSessionCh:
		assert.Equal(t, true, endSession)
	case <-time.After(time.Second):
		t.Fatal("Didnt receive end game session before timeout.")
	}

}

// TestRulesSystem_Update_ContinuesWhenMultipleAlive tests that the game continues
// when multiple players are still alive
func TestRulesSystem_Update_ContinuesWhenMultipleAlive(t *testing.T) {
	// Setup
	rulesSystem := systems.NewRulesSystem()
	endSessionCh := make(chan bool, 1)
	deltaTime := 0.016 // 60 FPS frame time

	// Create entity manager and entities
	em := ecs.NewEntityManager()

	// TODO: Create test entities with multiple alive players
	// TODO: Set up match progress component
	// TODO: Ensure at least 2 players are alive

	entities := em.GetAllEntities()

	// Act
	rulesSystem.Update(deltaTime, entities, endSessionCh)

	// Assert
	// TODO: Verify endSessionCh does NOT receive any signal
	// TODO: Verify match progress correctly tracks alive players
}
