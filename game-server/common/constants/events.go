package commonconstants

/**
NOTE: for those joining my project for the first time, follow this rule:

Simple rule:

Constant      |    Pattern                             |    Example
Exchange         {domain}.events                            game.events
Routing Key      {resource}.{action}                        match.ended
Exchange         {service}.{domain}.{resource}.{action}     stats.game.match.ended

**/

/**
* Exchange
* NOTE: try to keep these one per domain
* {domain}.events
**/
const (
	GameEventsExchange = "game.events"
	AuthEventsExchange = "auth.events"
	ItemEventsExchange = "item.events"
)

/**
* Message Broker Events
* NOTE: also acting as Routing Keys
* {resource}.{action}
**/
const (
	// example
	ExampleCreatedEvent = "example.created"

	// Member Events
	MemberSignedUpEvent = "member.signedup"       // when user creates account
	MemberSignedInEvent = "member.signedin"       // when user signs into their account
	PasswordResetEvent  = "member.password.reset" // when password reset is requested

	MemberProfileUpdated = "profile.updated"

	// Game Events
	GameMatchEnded = "match.ended" // match ended

	// Item Events
	ItemCreated = "item.created"
)

/**
* Queue Names
* NOTE: {service}.{domain}.{resource}.{action}
**/
const (
	StatsGameMatchEndedQueue        = "stats.game.match.ended"
	StatsAuthProfileUpdatedQueue    = "stats.auth.profile.updated"
	NotificationMemberSignedUpQueue = "notification.auth.member.signedup"
	NotificationItemCreatedQueue = "notification.item.item.created"
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
