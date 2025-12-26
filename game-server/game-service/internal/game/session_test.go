package game

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/**
* testing all session related business logic like creation and
* session manipulation.
**/

type mockMessageSender struct{}

func (m *mockMessageSender) SendMessageInternal(
	playerID uuid.UUID,
	msg types.Message,
) error {
	return nil
}

// mock sender for testing
func createMockSender() *types.MessageSender {
	mockMessageSender := &mockMessageSender{}
	return types.NewMessageSender(mockMessageSender)
}

// TestSessionCreation tests that a session initializes correctly with players
// white box test, we need to verify internal state like playerEntities
func TestSessionCreation(t *testing.T) {
	sender := createMockSender()
	stateSerializer := serializer.NewStateSerializer()
	session := NewSession(sender, stateSerializer)

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
	stateSerializer := serializer.NewStateSerializer()
	session := NewSession(sender, stateSerializer)
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
	stateSerializer := serializer.NewStateSerializer()
	session := NewSession(sender, stateSerializer)
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
	stateSerializer := serializer.NewStateSerializer()
	session := NewSession(sender, stateSerializer)

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

type handleInteractTable []struct {
	doorX              float64
	doorY              float64
	expectedOutOfRange bool
}

func TestHandleInteract(t *testing.T) {

	tableTests := handleInteractTable{
		{
			doorX:              0.1,
			doorY:              0.1,
			expectedOutOfRange: false,
		},
		{
			doorX:              1.5,
			doorY:              1.5,
			expectedOutOfRange: true,
		},
		{
			doorX:              100.0,
			doorY:              100.0,
			expectedOutOfRange: true,
		},
		{
			doorX:              0.2,
			doorY:              0.1,
			expectedOutOfRange: false,
		},
	}

	sender := createMockSender()
	stateSerializer := serializer.NewStateSerializer()
	session := NewSession(sender, stateSerializer)

	player1ID := uuid.New()
	username := "Player1"

	// default location 0, 0
	session.AddPlayer(player1ID, username)

	for _, tableTest := range tableTests {
		// door one, door thats out of range
		doorOneEntityID := session.AddDoor(tableTest.doorX, tableTest.doorY)
		doorEntity, _ := session.EntityManager.GetEntity(doorOneEntityID)
		doorEntity.GetComponent(ecs.ComponentTypeOpenable)

		time.Sleep(time.Millisecond * 150) // delay to account for rate limiting

		err := session.handleInteract(player1ID, doorOneEntityID)

		// expect out of range
		if tableTest.expectedOutOfRange {
			isOutOfRange := errors.Is(err, ErrOutOfRange)
			assert.Equal(t, true, isOutOfRange)
			continue
		}

		assert.Nil(t, err)

		// check its opposite
	}
}
