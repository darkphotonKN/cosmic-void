package constants

type Action string

const (
	// menu actions
	ActionQueue      Action = "queue"
	ActionFindGame   Action = "find_game"
	ActionLeaveQueue Action = "leave_queue"

	// active game actions
	ActionMove     Action = "move"
	ActionAttack   Action = "attack"
	ActionPickup   Action = "pickup"
	ActionUseItem  Action = "use_item"
	ActionDropItem Action = "drop_item"
	ActionChat     Action = "chat"
)
