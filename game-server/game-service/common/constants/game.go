package constants

type Action string
type ErrorCode string

const (
	// menu actions
	ActionQueue      Action = "queue"
	ActionFindGame   Action = "find_game"
	ActionLeaveQueue Action = "leave_queue"

	// active game actions
	ActionMove     Action = "move"
	ActionInteract Action = "interact"
	ActionAttack   Action = "attack"
	ActionPickup   Action = "pickup"
	ActionUseItem  Action = "use_item"
	ActionDropItem Action = "drop_item"
	ActionChat     Action = "chat"
	ActionLoot     Action = "loot"

	// system actions
	ActionError   Action = "error"
	ActionSuccess Action = "success"
)

const (
	ErrorSessionNotFound     ErrorCode = "session_not_found"
	ErrorInvalidSessionID    ErrorCode = "invalid_session_id"
	ErrorPlayerNotFound      ErrorCode = "player_not_found"
	ErrorInvalidPayload      ErrorCode = "invalid_payload"
	ErrorInternalServerError ErrorCode = "internal_server_error"
)

// Server Memory Constants
const MaxMsgChanBuffer int = 30

// Game Defaults
const DefaultSpeed float64 = 200
const DefaultInteractableRange float64 = 60
const DefautMaxSessionPlayers = 2

// map setting
const MapWidth float64 = 1200
const MapHeight float64 = 800
const PlayerRadius float64 = 20
const ContainerWidthRadius float64 = 20
const ContainerHeightRadius float64 = 16
const InitialPlayerX float64 = 600
const InitialPlayerY float64 = 400

// game loop
const GameFrameRate int = 30

// Item Pool Configuration
const ItemPoolSize int = 40 // Total number of item slots in the pool

// Item type ratios (must sum to 100)
const WeaponRatio int = 40     // 40% weapons
const ArmorRatio int = 35      // 35% armors
const ConsumableRatio int = 25 // 25% consumables

type ConnectState string

const (
	Connected    ConnectState = "connected"
	Disconnected ConnectState = "disconnected"
	Reconnecting ConnectState = "reconnecting"
)
