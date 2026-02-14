package cache

import "fmt"

var leaderboardKeyTemplate = "stats:leaderboard:%d:%d"

func GetLeaderboardKey(limit, offset int) string {
	leaderboardKey := fmt.Sprintf(leaderboardKeyTemplate, limit, offset)
	return leaderboardKey
}
