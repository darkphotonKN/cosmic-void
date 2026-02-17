package cache

import "fmt"

const (
	leaderboardKeyTemplate = "stats:leaderboard:v%d:%d:%d"
	leaderboardVersionKey  = "stats:leaderboard:version"
)

func StatsLeaderboardKey(version int64, limit, offset int) string {
	return fmt.Sprintf(leaderboardKeyTemplate, version, limit, offset)
}

func StatsLeaderboardVersionKey() string {
	return leaderboardVersionKey
}
