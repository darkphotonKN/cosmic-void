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

const DefaultSpeed float64 = 1
