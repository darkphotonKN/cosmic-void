package commonconstants

/**
* Message Broker Events
**/
const (
	// example
	ExampleCreatedEvent = "example.created"

	// Member Events
	MemberSignedUpEvent = "member.signedup"       // when user creates account
	MemberSignedInEvent = "member.signedin"       // when user signs into their account
	PasswordResetEvent  = "member.password_reset" // when password reset is requested

	// Build events
	BuildCreatedEvent   = "build.created"   // when build is first created (draft)
	BuildPublishedEvent = "build.published" // when build is made public
	BuildUpdatedEvent   = "build.updated"   // when published build is edited
	BuildDeletedEvent   = "build.deleted"   // when build is deleted
	BuildRatedEvent     = "build.rated"     // when someone rates a build)

	// Item events
	ItemCreatedItemEvent = "item.created" // when item is created

	// Game events
	RoomCreatedEvent = "room.created" // when game room is created
	GameStartedEvent = "game.started" // when game starts
	GameEndedEvent   = "game.ended"   // when game ends
)

/**
* Message Broker Event Payloads
**/

/**
* MemberSignedUpEventPayload
*
* Published by auth-service.
* Consumed by:
* - notification-service
* - analytics-service
**/
type MemberSignedUpEventPayload struct {
	UserID     string `json:"userId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	SignedUpAt string `json:"signedUpAt"`
}

/**
* MemberSignedInEventPayload
*
* Published by auth-service.
* Consumed by:
* - notification-service
* - analytics-service
**/
type MemberSignedInEventPayload struct {
	UserID string `json:"userId"`
}

/*
*
* type ItemCreatedItemEventPayload struct {

*
* Published by item-service.
* Consumed by:
* - notification-service
*
 */
type ItemCreatedItemEventPayload struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	// Email      string `json:"email"`
	SignedUpAt string `json:"signedUpAt"`
}

/**
* Game Event Payloads
**/

/**
* RoomCreatedEventPayload
*
* Published by game-service.
* Consumed by:
* - notification-service
* - analytics-service
**/
type RoomCreatedEventPayload struct {
	RoomID    string `json:"roomId"`
	Name      string `json:"name"`
	CreatorID string `json:"creatorId"`
	GameMode  string `json:"gameMode"`
	CreatedAt string `json:"createdAt"`
}

/**
* GameStartedEventPayload
*
* Published by game-service.
* Consumed by:
* - analytics-service
* - notification-service
**/
type GameStartedEventPayload struct {
	RoomID    string   `json:"roomId"`
	GameMode  string   `json:"gameMode"`
	PlayerIDs []string `json:"playerIds"`
	StartedAt string   `json:"startedAt"`
}

/**
* GameEndedEventPayload
*
* Published by game-service.
* Consumed by:
* - analytics-service
* - notification-service
**/
type GameEndedEventPayload struct {
	RoomID   string `json:"roomId"`
	GameMode string `json:"gameMode"`
	WinnerID string `json:"winnerId"`
	EndedAt  string `json:"endedAt"`
}
