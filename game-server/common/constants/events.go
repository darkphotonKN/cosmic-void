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

	// Stats events
	StatsUpdatedEvent     = "stats.updated"      // when player stats are updated
	MatchCompletedEvent   = "match.completed"    // when a game match is completed
	PlayerActionEvent     = "player.action"     // when player performs an action
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
* StatsUpdatedEventPayload
*
* Published by stats-service.
* Consumed by:
* - notification-service
* - leaderboard-service
**/
type StatsUpdatedPayload struct {
	PlayerID  string `json:"playerId"`
	Level     int32  `json:"level"`
	XP        int32  `json:"xp"`
	UpdatedAt string `json:"updatedAt"`
}

/**
* MatchCompletedEventPayload
*
* Published by game-service.
* Consumed by:
* - stats-service
**/
type MatchCompletedPayload struct {
	MatchID  string `json:"matchId"`
	Duration int32  `json:"duration"`
	Players  []struct {
		PlayerID       string  `json:"playerId"`
		Won            bool    `json:"won"`
		Kills          int32   `json:"kills"`
		Deaths         int32   `json:"deaths"`
		Assists        int32   `json:"assists"`
		DamageDealt    float32 `json:"damageDealt"`
		DamageTaken    float32 `json:"damageTaken"`
		ItemsCollected int32   `json:"itemsCollected"`
	} `json:"players"`
}

/**
* PlayerActionEventPayload
*
* Published by game-service.
* Consumed by:
* - stats-service
**/
type PlayerActionPayload struct {
	PlayerID    string  `json:"playerId"`
	Action      string  `json:"action"` // "kill", "death", "assist", "item_collected"
	DamageDealt float32 `json:"damageDealt,omitempty"`
	DamageTaken float32 `json:"damageTaken,omitempty"`
}
