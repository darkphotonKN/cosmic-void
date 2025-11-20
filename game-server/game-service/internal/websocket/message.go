package websocket

type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	SenderID  string      `json:"sender_id,omitempty"`
	RoomID    string      `json:"room_id,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

type PlayerMovePayload struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
}

type PlayerActionPayload struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data,omitempty"`
}

type ChatMessagePayload struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	RoomID  string `json:"room_id"`
}

type GameStatePayload struct {
	Players   []PlayerState `json:"players"`
	GameTick  int64         `json:"game_tick"`
	Timestamp string        `json:"timestamp"`
}

type PlayerState struct {
	UserID    string  `json:"user_id"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
	Health    int     `json:"health"`
	Score     int     `json:"score"`
	IsAlive   bool    `json:"is_alive"`
}

type RoomUpdatePayload struct {
	RoomID         string `json:"room_id"`
	Status         string `json:"status"`
	CurrentPlayers int    `json:"current_players"`
	MaxPlayers     int    `json:"max_players"`
}