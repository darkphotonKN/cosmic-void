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

func TestSession_InitializeItems_CreateItemEntities(t *testing.T) {
	expectedWeaponCount := 2

	// table tests
	tests := []struct {
		name            string
		mockReturn      *pb.ListItemTemplatesResponse
		mockErr         error
		wantErr         bool
		wantItemCount   int
		wantWeaponCount *int
	}{
		{
			name: "creates item entities properly",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{
					// Weapons
					{
						Id:            "aa000000-0000-0000-0000-000000000001",
						ItemName:      "Vibro-blade",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/vibro_blade.png",
						RequiredLevel: 1,
						BaseSellPrice: 2,
						BaseBuyPrice:  5,
						AttackPower:   6,
						CriticalRate:  0.08,
						WeaponType:    "sword",
						Description:   "Mono-molecular vibrating combat blade",
					},
					{
						Id:            "aa000000-0000-0000-0000-000000000002",
						ItemName:      "Pulse Dagger",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/pulse_dagger.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  3,
						AttackPower:   3,
						CriticalRate:  0.12,
						WeaponType:    "knife",
						Description:   "Compact energy-pulse combat knife",
					},
					// Armors
					{
						Id:              "aa000000-0000-0000-0000-000000000007",
						ItemName:        "Titanium Helmet",
						Rarity:          "Common",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/titanium_helmet.png",
						RequiredLevel:   1,
						BaseSellPrice:   1,
						BaseBuyPrice:    4,
						DefenseRating:   3,
						MagicResistance: 1,
						ArmorSlot:       "head",
						Description:     "Standard-issue titanium alloy combat helmet",
					},
					{
						Id:              "aa000000-0000-0000-0000-000000000008",
						ItemName:        "Titanium Chest Plate",
						Rarity:          "Common",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/titanium_chest_plate.png",
						RequiredLevel:   1,
						BaseSellPrice:   3,
						BaseBuyPrice:    6,
						DefenseRating:   6,
						MagicResistance: 2,
						ArmorSlot:       "chest",
						Description:     "Titanium alloy chest plate with ballistic lining",
					},
					// Consumables
					{
						Id:            "aa000000-0000-0000-0000-000000000023",
						ItemName:      "Minor Stim Pack",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/minor_stim_pack.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  2,
						HealingAmount: 10,
						ManaAmount:    0,
						BuffDuration:  0,
						MaxStackSize:  20,
						Description:   "Basic nano-med stim injection",
					},
					{
						Id:            "aa000000-0000-0000-0000-000000000024",
						ItemName:      "Greater Stim Pack",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/greater_stim_pack.png",
						RequiredLevel: 1,
						BaseSellPrice: 2,
						BaseBuyPrice:  5,
						HealingAmount: 25,
						ManaAmount:    0,
						BuffDuration:  0,
						MaxStackSize:  10,
						Description:   "Advanced regenerative stim pack",
					},
				},
			},
			wantErr:       false,
			wantItemCount: 6,
		},
		{
			name: "weapon counts match not affect by consumables and armor",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{
					// Weapons
					{
						Id:            "aa000000-0000-0000-0000-000000000001",
						ItemName:      "Vibro-blade",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/vibro_blade.png",
						RequiredLevel: 1,
						BaseSellPrice: 2,
						BaseBuyPrice:  5,
						AttackPower:   6,
						CriticalRate:  0.08,
						WeaponType:    "sword",
						Description:   "Mono-molecular vibrating combat blade",
					},
					{
						Id:            "aa000000-0000-0000-0000-000000000002",
						ItemName:      "Pulse Dagger",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/pulse_dagger.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  3,
						AttackPower:   3,
						CriticalRate:  0.12,
						WeaponType:    "knife",
						Description:   "Compact energy-pulse combat knife",
					},
					// armor
					{
						Id:              "aa000000-0000-0000-0000-000000000008",
						ItemName:        "Titanium Chest Plate",
						Rarity:          "Common",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/titanium_chest_plate.png",
						RequiredLevel:   1,
						BaseSellPrice:   3,
						BaseBuyPrice:    6,
						DefenseRating:   6,
						MagicResistance: 2,
						ArmorSlot:       "chest",
						Description:     "Titanium alloy chest plate with ballistic lining",
					},
					// consumables
					{
						Id:            "aa000000-0000-0000-0000-000000000023",
						ItemName:      "Minor Stim Pack",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/minor_stim_pack.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  2,
						HealingAmount: 10,
						ManaAmount:    0,
						BuffDuration:  0,
						MaxStackSize:  20,
						Description:   "Basic nano-med stim injection",
					},
				},
			},
			mockErr:         nil,
			wantErr:         false,
			wantItemCount:   4,
			wantWeaponCount: &expectedWeaponCount,
		},
		{
			name: "handles empty template list",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{},
			},
			mockErr:       nil,
			wantErr:       false,
			wantItemCount: 0,
		},
	}

	// --- testing ---
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// -- setup --
			sender := createMockSender()
			em := ecs.NewEntityManager()
			mockEmitter := &mockEventEmitter{}
			mockClient := mockItemsClient{}

			session := NewSession(sender, &mockStateSerializer{}, em, mockEmitter, &mockClient)

			session.TestMessageSpy = make(chan types.Message, 1)

			defer session.Shutdown()

			mockClient.On("ListItemTemplates", mock.Anything).Return(
				tt.mockReturn,
				tt.mockErr,
			)

			// -- test --

			err := session.InitializeItems(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// get all entites and check creations
			allWeaponCounts := 0

			// slog.Info("session.itemPool during test",
			// 	"test name", tt.name,
			// 	"session.itemPool", session.itemPool,
			// )

			assert.Equal(t, tt.wantItemCount, session.itemPool.Count)

			// testing weapon item type specific counts check out
			if tt.wantWeaponCount != nil {

				for _, item := range session.itemPool.Weapons {
					if item.ItemType == types.ItemTypeWeapon {
						allWeaponCounts++
					}
				}

				assert.Equal(t, *tt.wantWeaponCount, allWeaponCounts)
			}

			mockClient.AssertExpectations(t)
		},
		)
	}
}

func TestSession_GenerateItems_CreateItemEntities(t *testing.T) {
	// table test data
	tests := []struct {
		name       string
		mockReturn *pb.ListItemTemplatesResponse
		mockErr    error
		wantErr    bool
	}{
		{
			name: "Creates correct number of item entities",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{
					// Weapons
					{
						Id:            "aa000000-0000-0000-0000-000000000001",
						ItemName:      "Vibro-blade",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/vibro_blade.png",
						RequiredLevel: 1,
						BaseSellPrice: 2,
						BaseBuyPrice:  5,
						AttackPower:   6,
						CriticalRate:  0.08,
						WeaponType:    "sword",
						Description:   "Mono-molecular vibrating combat blade",
					},
					{
						Id:            "aa000000-0000-0000-0000-000000000002",
						ItemName:      "Pulse Dagger",
						Rarity:        "Common",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/pulse_dagger.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  3,
						AttackPower:   3,
						CriticalRate:  0.12,
						WeaponType:    "knife",
						Description:   "Compact energy-pulse combat knife",
					},
					// Armors
					{
						Id:              "aa000000-0000-0000-0000-000000000007",
						ItemName:        "Titanium Helmet",
						Rarity:          "Common",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/titanium_helmet.png",
						RequiredLevel:   1,
						BaseSellPrice:   1,
						BaseBuyPrice:    4,
						DefenseRating:   3,
						MagicResistance: 1,
						ArmorSlot:       "head",
						Description:     "Standard-issue titanium alloy combat helmet",
					},
					{
						Id:              "aa000000-0000-0000-0000-000000000008",
						ItemName:        "Titanium Chest Plate",
						Rarity:          "Common",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/titanium_chest_plate.png",
						RequiredLevel:   1,
						BaseSellPrice:   3,
						BaseBuyPrice:    6,
						DefenseRating:   6,
						MagicResistance: 2,
						ArmorSlot:       "chest",
						Description:     "Titanium alloy chest plate with ballistic lining",
					},
					// Consumables
					{
						Id:            "aa000000-0000-0000-0000-000000000023",
						ItemName:      "Minor Stim Pack",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/minor_stim_pack.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  2,
						HealingAmount: 10,
						ManaAmount:    0,
						BuffDuration:  0,
						MaxStackSize:  20,
						Description:   "Basic nano-med stim injection",
					},
					{
						Id:            "aa000000-0000-0000-0000-000000000024",
						ItemName:      "Greater Stim Pack",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/greater_stim_pack.png",
						RequiredLevel: 1,
						BaseSellPrice: 2,
						BaseBuyPrice:  5,
						HealingAmount: 25,
						ManaAmount:    0,
						BuffDuration:  0,
						MaxStackSize:  10,
						Description:   "Advanced regenerative stim pack",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "returns error when itemPool empty",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{},
			},
			mockErr: nil,
			wantErr: true,
		},
		{
			name: "handles mixed rarity consumable-heavy loadout",
			mockReturn: &pb.ListItemTemplatesResponse{
				Items: []*pb.ItemTemplate{
					// Weapon
					{
						Id:            "bb000000-0000-0000-0000-000000000001",
						ItemName:      "Plasma Rifle",
						Rarity:        "Rare",
						ItemType:      "weapon",
						IconUrl:       "/icons/weapon/plasma_rifle.png",
						RequiredLevel: 5,
						BaseSellPrice: 15,
						BaseBuyPrice:  30,
						AttackPower:   18,
						CriticalRate:  0.05,
						WeaponType:    "rifle",
						Description:   "High-energy plasma discharge rifle",
					},
					// Armor
					{
						Id:              "bb000000-0000-0000-0000-000000000002",
						ItemName:        "Nano-Weave Boots",
						Rarity:          "Uncommon",
						ItemType:        "armor",
						IconUrl:         "/icons/armor/nano_weave_boots.png",
						RequiredLevel:   3,
						BaseSellPrice:   4,
						BaseBuyPrice:    10,
						DefenseRating:   4,
						MagicResistance: 3,
						ArmorSlot:       "feet",
						Description:     "Lightweight boots with nano-fiber reinforcement",
					},
					// Consumables
					{
						Id:            "bb000000-0000-0000-0000-000000000003",
						ItemName:      "Mana Siphon",
						Rarity:        "Uncommon",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/mana_siphon.png",
						RequiredLevel: 2,
						BaseSellPrice: 3,
						BaseBuyPrice:  7,
						HealingAmount: 0,
						ManaAmount:    20,
						BuffDuration:  0,
						MaxStackSize:  15,
						Description:   "Extracts ambient energy to restore mana",
					},
					{
						Id:            "bb000000-0000-0000-0000-000000000004",
						ItemName:      "Shield Booster",
						Rarity:        "Rare",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/shield_booster.png",
						RequiredLevel: 4,
						BaseSellPrice: 5,
						BaseBuyPrice:  12,
						HealingAmount: 0,
						ManaAmount:    0,
						BuffDuration:  30,
						MaxStackSize:  5,
						Description:   "Temporarily reinforces personal energy shield",
					},
					{
						Id:            "bb000000-0000-0000-0000-000000000005",
						ItemName:      "Adrenaline Shot",
						Rarity:        "Common",
						ItemType:      "consumable",
						IconUrl:       "/icons/consumable/adrenaline_shot.png",
						RequiredLevel: 1,
						BaseSellPrice: 1,
						BaseBuyPrice:  3,
						HealingAmount: 5,
						ManaAmount:    5,
						BuffDuration:  10,
						MaxStackSize:  20,
						Description:   "Quick-inject combat stimulant",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		// setup
		t.Run(tt.name, func(t *testing.T) {
			// -- setup --
			sender := createMockSender()
			em := ecs.NewEntityManager()
			mockEmitter := &mockEventEmitter{}
			mockClient := mockItemsClient{}

			session := NewSession(sender, &mockStateSerializer{}, em, mockEmitter, &mockClient)

			session.TestMessageSpy = make(chan types.Message, 1)

			defer session.Shutdown()

			mockClient.On("ListItemTemplates", mock.Anything).Return(
				tt.mockReturn,
				tt.mockErr,
			)

			// initialize seed items
			err := session.InitializeItems(context.Background())

			if err != nil {
				t.Fatal("Unexpected error when testing.")
			}

			// -- test --
			ids, err := session.generateItems()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// check ids exist in entity pool
			itemsHashMap := make(map[uuid.UUID]*components.ItemComponent)
			entities := session.EntityManager.GetAllEntities()

			for _, entity := range entities {
				itemComp, isItem := entity.GetComponent(ecs.ComponentTypeItem)
				if !isItem {
					continue
				}
				item, ok := itemComp.(*components.ItemComponent)

				if !ok {
					t.Fatal("unexpected error when asserting for item component")
				}

				itemsHashMap[entity.ID] = item
			}

			// loop through generated items
			generatedItems := 0
			for _, id := range ids {
				_, exists := itemsHashMap[id]
				assert.True(t, exists)
				if exists {
					generatedItems++
				}
			}

			t.Log("itemHashMap", itemsHashMap, "generatedItems", generatedItems, "ids", ids)

			assert.Equal(t, len(ids), generatedItems)
		})
	}
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

func (c *mockItemsClient) ListItemTemplates(ctx context.Context) (*pb.ListItemTemplatesResponse, error) {
	args := c.Called(ctx)

	return args.Get(0).(*pb.ListItemTemplatesResponse), args.Error(1)
}

func (c *mockItemsClient) ListArmorsWithTemplate(ctx context.Context) (*pb.ListArmorsResponse, error) {
	return nil, nil
}

func (c *mockItemsClient) ListConsumablesWithTemplate(ctx context.Context) (*pb.ListConsumablesResponse, error) {
	return nil, nil
}

type mockStateSerializer struct {
}

func (s *mockStateSerializer) PutBackendState(state *types.BackendGameState) {
}

func (s *mockStateSerializer) SerializeBackendState(ctx context.Context, sessionID uuid.UUID, entities []*ecs.Entity) (*types.BackendGameState, error) {
	return nil, nil
}

func (s *mockStateSerializer) FormatStateToClientState(backendState *types.BackendGameState, playerID uuid.UUID) *types.ClientGameState {
	return nil
}
