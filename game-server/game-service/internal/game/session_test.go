package game

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/items"

	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/components"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/ecs"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/messaging"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/serializer"
	"github.com/darkphotonKN/cosmic-void-server/game-service/internal/types"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

/**
* testing all session related business logic like creation and
* session manipulation.
**/

type mockMessageSender struct{}

func (m *mockMessageSender) PushMessageToChannelQueue(
	playerID uuid.UUID,
	msg interface{},
) error {
	return nil
}

func (m *mockMessageSender) PushMessageToConn(
	conn *websocket.Conn,
	msg interface{},
) error {
	return nil
}

// mock sender for testing
func createMockSender() *messaging.MessageSender {
	mockMessageSender := &mockMessageSender{}
	return messaging.NewMessageSender(mockMessageSender)
}

// TestSessionCreation tests that a session initializes correctly with players
// white box test, we need to verify internal state like playerEntities
func TestSessionCreation(t *testing.T) {
	sender := createMockSender()
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)

	// verify session initialized
	require.NotNil(t, session, "Session should not be nil")
	require.NotEqual(t, uuid.Nil, session.ID, "Session should have valid ID")
	require.NotNil(t, session.EntityManager, "EntityManager should be initialized")
	require.NotNil(t, session.MessageCh, "MessageCh should be initialized")
	require.NotNil(t, session.sender, "Sender should be initialized")

	// initial state checks
	assert.Equal(t, 0, len(session.playerIDToEntitiesID), "Should have no players initially")

	// clean up goroutines
	defer session.Shutdown()
}

// test adding a single player to an existing session
func TestSessionAddPlayer(t *testing.T) {
	sender := createMockSender()
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)
	defer session.Shutdown()

	playerID := uuid.New()
	username := "TestPlayer"

	entityID := session.AddPlayer(playerID, username)

	assert.NotEqual(t, uuid.Nil, entityID, "Should return valid entity ID")

	assert.Equal(t, 1, len(session.playerIDToEntitiesID), "Should have 1 player")
	storedEntityID, exists := session.playerIDToEntitiesID[playerID]
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
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)
	defer session.Shutdown()

	player1ID := uuid.New()
	player2ID := uuid.New()

	entity1ID := session.AddPlayer(player1ID, "Player1")
	entity2ID := session.AddPlayer(player2ID, "Player2")

	assert.NotEqual(t, entity1ID, entity2ID, "Entity IDs should be unique")
	assert.Equal(t, 2, len(session.playerIDToEntitiesID), "Should have 2 players")

	_, exists1 := session.EntityManager.GetEntity(entity1ID)
	_, exists2 := session.EntityManager.GetEntity(entity2ID)
	assert.True(t, exists1, "Player 1 entity should exist")
	assert.True(t, exists2, "Player 2 entity should exist")
}

// NOTE: note to team, also white box test here, testing internals
// test initial coordinates are correctly set by addPlayer
func TestAddPlayerSetsInitialPosition(t *testing.T) {
	sender := createMockSender()
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)

	player1ID := uuid.New()
	username := "Player1"
	playerEntityID := session.AddPlayer(player1ID, username)

	// check player initial position
	playerEntity, ok := session.EntityManager.GetEntity(playerEntityID)

	if !ok {
		slog.Error("PlayerEntity doesn't exist", "playerEntityID", playerEntityID)
	}

	playerTransformComponent, ok := playerEntity.GetComponent(ecs.ComponentTypeTransform)

	if !ok {
		slog.Error("Player's Velocity Component doesn't exist", "entityID", playerEntity.ID)
	}

	component := playerTransformComponent.(*components.TransformComponent)
	slog.Debug("Player transform coordinates initial", "coordinates", component)

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
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)

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

type handleInteractContainerTable []struct {
	containerX         float64
	containerY         float64
	expectedOutOfRange bool
}

func TestHandleInteractContainer(t *testing.T) {

	tableTests := handleInteractContainerTable{
		{
			containerX:         0.1,
			containerY:         0.1,
			expectedOutOfRange: false,
		},
		{
			containerX:         1.5,
			containerY:         1.5,
			expectedOutOfRange: true,
		},
		{
			containerX:         100.0,
			containerY:         100.0,
			expectedOutOfRange: true,
		},
		{
			containerX:         0.2,
			containerY:         0.1,
			expectedOutOfRange: false,
		},
	}

	sender := createMockSender()
	em := ecs.NewEntityManager()
	stateSerializer := serializer.NewStateSerializer(em)
	mockEmitter := &mockEventEmitter{}
	session := NewSession(sender, stateSerializer, em, mockEmitter, nil)

	player1ID := uuid.New()
	username := "Player1"

	// default location 0, 0
	session.AddPlayer(player1ID, username)

	for _, tableTest := range tableTests {
		// container one, container thats out of range
		containerOneEntityID := session.AddContainer(tableTest.containerX, tableTest.containerY)
		containerEntity, _ := session.EntityManager.GetEntity(containerOneEntityID)

		// first open
		containerOpenableComponent, _ := containerEntity.GetComponent(ecs.ComponentTypeOpenable)

		containerOpenable := containerOpenableComponent.(*components.OpenableComponent)

		containerItemIDListComponent, _ := containerEntity.GetComponent(ecs.ComponentTypeItemIDList)
		containerItemIDList := containerItemIDListComponent.(*components.ItemIDListComponent)

		assert.False(t, containerOpenable.HasBeenOpened)
		assert.Equal(t, 0, len(containerItemIDList.ItemIDs))

		time.Sleep(time.Millisecond * 150) // delay to account for rate limiting
		// first time open
		err := session.handleInteract(player1ID, containerOneEntityID)

		// expect out of range
		if tableTest.expectedOutOfRange {
			isOutOfRange := errors.Is(err, ErrOutOfRange)
			assert.Equal(t, true, isOutOfRange)
			continue
		}

		assert.Nil(t, err)
		// verify 1-4 items
		assert.Equal(t, true, containerOpenable.HasBeenOpened)
		assert.GreaterOrEqual(t, len(containerItemIDList.ItemIDs), 1, "At least 1 item")
		assert.LessOrEqual(t, len(containerItemIDList.ItemIDs), 4, "At most 4 item")

		firstOpenItemIDs := make([]uuid.UUID, len(containerItemIDList.ItemIDs))
		copy(firstOpenItemIDs, containerItemIDList.ItemIDs)

		// second time open
		// verify items are same
		time.Sleep(time.Millisecond * 150)
		err = session.handleInteract(player1ID, containerOneEntityID)
		assert.Nil(t, err)

		containerItemIDListComponent2, _ := containerEntity.GetComponent(ecs.ComponentTypeItemIDList)
		containerItemIDList2 := containerItemIDListComponent2.(*components.ItemIDListComponent)
		assert.Equal(t, firstOpenItemIDs, containerItemIDList2.ItemIDs)
	}
}

func TestSession_InitializeBaseArmors_CreateItemEntities(t *testing.T) {
	sender := createMockSender()
	em := ecs.NewEntityManager()
	mockEmitter := &mockEventEmitter{}

	mockClient := mockItemsClient{}

	mockClient.On("ListArmorsWithTemplate", mock.Anything, mock.Anything).Return(&pb.ListArmorsResponse{
		Armors: []*pb.ArmorDetail{
			{
				Id:            "550e8400-e29b-41d4-a716-446655440001", // fake armor 1
				ItemName:      "Iron Armor",
				DefenseRating: 50,
			},
			{
				Id:            "550e8400-e29b-41d4-a716-446655440002", // fake armor 2
				ItemName:      "Steel Armor",
				DefenseRating: 100,
			},
		},
	},
		nil, // [D] No error
	)

	session := NewSession(sender, &mockStateSerializer{}, em, mockEmitter, &mockClient)
	defer session.Shutdown()

	err := session.InitializeBaseArmors(context.Background())

	assert.NoError(t, err)

	slog.Info("Testerson")
}

type mockItemsClient struct {
	mock.Mock
}

func (c *mockItemsClient) CreateWeapon(ctx context.Context, req *pb.CreateWeaponRequest) (*pb.Weapon, error) {
	return nil, nil
}

func (c *mockItemsClient) GetWeaponWithTemplateByID(ctx context.Context, req *pb.GetWeaponRequest) (*pb.WeaponDetail, error) {
	return nil, nil
}

func (c *mockItemsClient) ListWeaponsWithTemplate(ctx context.Context) (*pb.ListWeaponsResponse, error) {

	return nil, nil
}

func (c *mockItemsClient) ListArmorsWithTemplate(ctx context.Context) (*pb.ListArmorsResponse, error) {
	return &pb.ListArmorsResponse{
		Armors: []*pb.ArmorDetail{
			{
				Id:              "88000000-0000-0000-0000-000000000001",
				RarityId:        "660e8400-e29b-41d4-a716-446655440001",
				DefenseRating:   3,
				MagicResistance: 1,
				ArmorSlot:       "head",
				Description:     "Standard-issue titanium alloy combat helmet",
				ItemTemplateId:  "aa000000-0000-0000-0000-000000000007",
				ItemName:        "Titanium Helmet",
				IconUrl:         "/icons/armor/titanium_helmet.png",
				RequiredLevel:   1,
				BaseSellPrice:   1,
				BaseBuyPrice:    4,
			},
			{
				Id:              "88000000-0000-0000-0000-000000000002",
				RarityId:        "660e8400-e29b-41d4-a716-446655440001",
				DefenseRating:   6,
				MagicResistance: 2,
				ArmorSlot:       "chest",
				Description:     "Titanium alloy chest plate with ballistic lining",
				ItemTemplateId:  "aa000000-0000-0000-0000-000000000008",
				ItemName:        "Titanium Chest Plate",
				IconUrl:         "/icons/armor/titanium_chest_plate.png",
				RequiredLevel:   1,
				BaseSellPrice:   3,
				BaseBuyPrice:    6,
			},
			{
				Id:              "88000000-0000-0000-0000-000000000003",
				RarityId:        "660e8400-e29b-41d4-a716-446655440001",
				DefenseRating:   4,
				MagicResistance: 1,
				ArmorSlot:       "legs",
				Description:     "Reinforced titanium alloy leg guards",
				ItemTemplateId:  "aa000000-0000-0000-0000-000000000009",
				ItemName:        "Titanium Leg Guards",
				IconUrl:         "/icons/armor/titanium_leg_guards.png",
				RequiredLevel:   1,
				BaseSellPrice:   2,
				BaseBuyPrice:    5,
			},
			{
				Id:              "88000000-0000-0000-0000-000000000004",
				RarityId:        "660e8400-e29b-41d4-a716-446655440001",
				DefenseRating:   2,
				MagicResistance: 1,
				ArmorSlot:       "gloves",
				Description:     "Articulated titanium alloy combat gauntlets",
				ItemTemplateId:  "aa000000-0000-0000-0000-000000000010",
				ItemName:        "Titanium Gauntlets",
				IconUrl:         "/icons/armor/titanium_gauntlets.png",
				RequiredLevel:   1,
				BaseSellPrice:   1,
				BaseBuyPrice:    3,
			},
		},
	}, nil
}

func (c *mockItemsClient) ListConsumablesWithTemplate(ctx context.Context) (*pb.ListConsumablesResponse, error) {
	return nil, nil
}

type mockStateSerializer struct {
}

func (s *mockStateSerializer) SerializeBackendState(ctx context.Context, sessionID uuid.UUID, entities []*ecs.Entity) (*types.BackendGameState, error) {
	return nil, nil
}

func (s *mockStateSerializer) Serialize(ctx context.Context, sessionID uuid.UUID, recipientPlayerID uuid.UUID, entities []*ecs.Entity) (*types.ClientGameState, error) {
	return nil, nil
}

func (s *mockStateSerializer) FormatStateToClientState(backendState *types.BackendGameState, playerID uuid.UUID) *types.ClientGameState {
	return nil
}
