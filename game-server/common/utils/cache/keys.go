package cache

import "fmt"

const (
	leaderboardKeyTemplate = "stats:leaderboard:version:%d:%d"
	leaderboardVersionKey  = "stats:leaderboard:version"
)

func StatsLeaderboardKey(limit, offset int) string {
	leaderboardKey := fmt.Sprintf(leaderboardKeyTemplate, limit, offset)
	return leaderboardKey
}

func StatsLeaderboardVersionKey() string {
	return leaderboardVersionKey
}
