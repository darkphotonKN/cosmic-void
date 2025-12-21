package game

import (
	"errors"
	"fmt"
	"testing"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
* testing all session related business logic like creation and
* session manipulation.
**/

// mock sender for testing
func createMockSender() *types.MessageSender {
	return types.NewMessageSender(func(playerID uuid.UUID, msg types.Message) error {
		// mock: do nothing
		return nil
	})
}

// TestSessionCreation tests that a session initializes correctly with players
// white box test, we need to verify internal state like playerEntities
func TestSessionCreation(t *testing.T) {
	sender := createMockSender()
	session := NewSession(sender)

	// verify session initialized
	require.NotNil(t, session, "Session should not be nil")
	require.NotEqual(t, uuid.Nil, session.ID, "Session should have valid ID")
	require.NotNil(t, session.EntityManager, "EntityManager should be initialized")
	require.NotNil(t, session.MessageCh, "MessageCh should be initialized")
	require.NotNil(t, session.sender, "Sender should be initialized")

	// initial state checks
	assert.Equal(t, 0, len(session.playerEntities), "Should have no players initially")

	// clean up goroutines
	defer session.Shutdown()
}

// test adding a single player to an existing session
func TestSessionAddPlayer(t *testing.T) {
	sender := createMockSender()
	session := NewSession(sender)
	defer session.Shutdown()

	playerID := uuid.New()
	username := "TestPlayer"

	entityID := session.AddPlayer(playerID, username)

	assert.NotEqual(t, uuid.Nil, entityID, "Should return valid entity ID")

	assert.Equal(t, 1, len(session.playerEntities), "Should have 1 player")
	storedEntityID, exists := session.playerEntities[playerID]
	assert.True(t, exists, "Player should be in playerEntities map")
	assert.Equal(t, entityID, storedEntityID, "Entity IDs should match")

	entity, exists := session.EntityManager.GetEntity(entityID)
	require.True(t, exists, "Entity should exist in EntityManager")

	assert.True(t, entity.HasComponent(ecs.ComponentTypePlayer), "Should have Player component")
	assert.True(t, entity.HasComponent(ecs.ComponentTypeTransform), "Should have Transform component")
	assert.True(t, entity.HasComponent(ecs.ComponentTypeVelocity), "Should have Velocity component")

	// TODO: temporarily removed for simpler version of the game
	// assert.True(t, entity.HasComponent(ecs.ComponentTypeHealth), "Should have Health component")
	// assert.True(t, entity.HasComponent(ecs.ComponentTypeInventory), "Should have Inventory component")
}

// test focused on validating multiplayer players can be added to an
// existing session
func TestSessionAddMultiplePlayers(t *testing.T) {
	sender := createMockSender()
	session := NewSession(sender)
	defer session.Shutdown()

	player1ID := uuid.New()
	player2ID := uuid.New()

	entity1ID := session.AddPlayer(player1ID, "Player1")
	entity2ID := session.AddPlayer(player2ID, "Player2")

	assert.NotEqual(t, entity1ID, entity2ID, "Entity IDs should be unique")
	assert.Equal(t, 2, len(session.playerEntities), "Should have 2 players")

	_, exists1 := session.EntityManager.GetEntity(entity1ID)
	_, exists2 := session.EntityManager.GetEntity(entity2ID)
	assert.True(t, exists1, "Player 1 entity should exist")
	assert.True(t, exists2, "Player 2 entity should exist")
}

// NOTE: note to team, also white box test here, testing internals
// test initial coordinates are correctly set by addPlayer
func TestAddPlayerSetsInitialPosition(t *testing.T) {
	sender := createMockSender()
	session := NewSession(sender)

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
}

// ----- Testing Session Handles -----

func TestHandleInteract(t *testing.T) {
	sender := createMockSender()
	session := NewSession(sender)

	player1ID := uuid.New()
	username := "Player1"

	// default location 0, 0
	playerEntityID := session.AddPlayer(player1ID, username)

	// door one, door thats out of range
	doorOneEntityID := session.AddDoor(1.1, 1.1)

	err := session.handleInteract(playerEntityID, doorOneEntityID)
	isOutOfRange := errors.Is(err, ErrOutOfRange)
	assert.Equal(t, true, isOutOfRange)

	// door two, door thats within range
	doorTwoEntityID := session.AddDoor(0.1, 0.1)
	err = session.handleInteract(playerEntityID, doorTwoEntityID)
	assert.Nil(t, err)
}
