package stats

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	pbstats "github.com/darkphotonKN/cosmic-void-server/common/api/proto/stats"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	commoncache "github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

type Repository interface {
	// transaction methods
	UpsertPlayerMatchStatsTx(ctx context.Context, tx *sqlx.Tx, params *UpdateStatsParams) (*PlayerMatchStats, error)
	UpsertPlayerRankingStatsTx(ctx context.Context, tx *sqlx.Tx, params *UpdatePlayerRankingsParams) (*PlayerRankingStats, error)
	CreateMatchHistoryTx(ctx context.Context, tx *sqlx.Tx, history *MatchHistory) error
	GetPlayerMatchStatsTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID) (*PlayerMatchStats, error)
	GetPlayerRankingStatsTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID) (*PlayerRankingStats, error)

	// non transaction methods
	UpsertPlayerRankingStats(ctx context.Context, params *UpdatePlayerRankingsParams) (*PlayerRankingStats, error)
	GetPlayerRankings(ctx context.Context, params *GetPlayerRankings) ([]*PlayerRankingStats, error)
}

type service struct {
	repo      Repository
	publishCh commonbroker.Publisher
	db        *sqlx.DB
	cache     cache.Cache
}

func NewService(repo Repository, publishCh commonbroker.Publisher, db *sqlx.DB, cache cache.Cache) *service {
	return &service{
		repo:      repo,
		publishCh: publishCh,
		db:        db,
		cache:     cache,
	}
}

/**
* Runs all the relevant processes after a match is completed, updating the
* relavant tables.
**/
func (s *service) ProcessMatchCompleted(ctx context.Context, req *pb.MatchEndedEvent) (*ProcessMatchCompletedResponse, error) {
	slog.Info("Processing match completed",
		"session_id", req.SessionId,
		"match_started_at", req.MatchStartedAt.AsTime(),
		"match_ended_at", req.MatchEndedAt.AsTime(),
	)

	for _, playerResults := range req.Players {
		slog.Info("Player match outcome",
			"playerResults", playerResults,
		)

		// wrap transaction for business critical sync up between players match
		// history and their personal stats
		err := commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
			// -- player stats --
			err := s.updatePlayerStats(ctx, tx, playerResults)
			if err != nil {
				slog.Error("error when updating match stats", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			// -- leaderboards --
			err = s.updateDenormalizedLeaderboard(ctx, tx, playerResults)
			if err != nil {
				slog.Error("error when updating leaderboard stats", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			// -- match history --
			sessionId, err := uuid.Parse(req.SessionId)
			if err != nil {
				slog.Error("error parsing session id", "sessionId", req.SessionId, "error", err)
				return err
			}

			memberId, err := uuid.Parse(playerResults.MemberId)
			if err != nil {
				slog.Error("error parsing member id", "memberId", playerResults.MemberId, "error", err)
				return err
			}

			matchHistory := &MatchHistory{
				SessionID:      sessionId,
				MemberID:       memberId,
				Win:            playerResults.Win,
				FinalPosition:  int(playerResults.FinalPosition),
				Kills:          int(playerResults.Kills),
				Deaths:         int(playerResults.Deaths),
				MatchStartedAt: req.MatchStartedAt.AsTime(),
			}

			err = s.repo.CreateMatchHistoryTx(ctx, tx, matchHistory)
			if err != nil {
				slog.Error("error creating match history", "memberID", playerResults.MemberId, "error", err)
				return err
			}

			return nil
		})

		// transaction errored, continue to next player
		// (roll back automagically handled in helper already)
		if err != nil {
			slog.Error("error occured during match processing transaction, rolledback.",
				"error", err)
			continue
		}

	}

	// TODO: call auth service for player information

	// transaction was fine, invalidate cache results for denormalized leaderboard
	// version increment to invalidate, no deletion needed, short TTL is enough
	go func() {
		newVersion, err := s.cache.Incr(context.Background(), commoncache.StatsLeaderboardVersionKey())

		if err != nil {
			slog.Warn("Cache version could not be updated", "error", err)
			return
		}

		slog.Info("Cache version updated", "for key", commoncache.StatsLeaderboardVersionKey(), "new version", newVersion)
	}()

	return &ProcessMatchCompletedResponse{
		Success: true,
		Message: "Match data processed successfully",
	}, nil
}

func (s *service) updatePlayerStats(ctx context.Context, tx *sqlx.Tx, player *pb.PlayerMatchResult) error {
	memberId, err := uuid.Parse(player.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return err
	}

	// TODO: recalculate averages, averages migration and update WIP
	matchHistoryData, err := s.getMatchHistory(ctx, memberId)

	if err != nil {
		slog.Info("Errored when attempting to get match history", "err", err)
		return err
	}

	slog.Info("Match history data for player", "player", player, "data", matchHistoryData)

	playerStats, err := s.repo.GetPlayerMatchStatsTx(ctx, tx, memberId)

	if err != nil {
		return err
	}

	if playerStats == nil {
		// initialize struct, players first time
		playerStats = &PlayerMatchStats{
			MemberID:            memberId,
			GamesPlayed:         0,
			Wins:                0,
			Losses:              0,
			Kills:               0,
			Deaths:              0,
			TimesPlacedTopThree: 0,
		}
	}

	playerStats.GamesPlayed += 1
	playerStats.Kills += int(player.Kills)
	playerStats.Deaths += int(player.Deaths)

	// TODO: check if player won
	if player.Win {
		playerStats.Wins += 1
	} else if player.FinalPosition == 5 {
		playerStats.Losses += 1
	}

	// update aggregate stats
	_, err = s.repo.UpsertPlayerMatchStatsTx(ctx, tx, &UpdateStatsParams{
		MemberID:            playerStats.MemberID,
		GamesPlayed:         playerStats.GamesPlayed,
		Wins:                playerStats.Wins,
		Losses:              playerStats.Losses,
		Kills:               playerStats.Kills,
		Deaths:              playerStats.Deaths,
		TimesPlacedTopThree: playerStats.TimesPlacedTopThree,
	})

	if err != nil {
		return err
	}
	return nil
}

func (s *service) getMatchHistory(ctx context.Context, memberID uuid.UUID) ([]*MatchHistory, error) {

	return nil, nil
}

func (s *service) calculateMatchAverage(ctx context.Context, matchHistory []*MatchHistory) (*PlayerMatchStats, error) {
	// TODO: calculate the new averages
	return nil, nil
}

/**
* handles setting up and updating the denormalized leaderboard stats.
**/
func (s *service) updateDenormalizedLeaderboard(ctx context.Context, tx *sqlx.Tx, results *pb.PlayerMatchResult) error {
	memberId, err := uuid.Parse(results.MemberId)
	if err != nil {
		slog.Info("Errored when attempting to get parse member id into UUID", "err", err)
		return err
	}

	stats, err := s.repo.GetPlayerMatchStatsTx(ctx, tx, memberId)

	if err != nil {
		return err
	}

	wins := stats.Wins
	topThree := stats.TimesPlacedTopThree
	if results.Win {
		wins += 1
	}
	if results.FinalPosition >= 3 {
		topThree += 1
	}

	// recalculate rank position
	rankingStats, err := s.repo.GetPlayerRankingStatsTx(ctx, tx, memberId)

	slog.Debug("getting ranking stats from GetPlayerRankingStats", "rankingStats", rankingStats)

	var ranking *int
	if rankingStats != nil {
		ranking = rankingStats.RankPosition
	}

	statsParam := &UpdatePlayerRankingsParams{
		MemberID:         memberId,
		Username:         results.Username,
		Wins:             wins,
		TopThrees:        topThree,
		Rating:           0,
		RankPosition:     ranking,
		LastCalculatedAt: time.Now(),
	}

	playerRankingStats, err := s.repo.UpsertPlayerRankingStatsTx(ctx, tx, statsParam)

	if err != nil {
		return err
	}

	slog.Debug("after upsert for player ranking stats", "playerRankingStats", playerRankingStats)

	return nil
}

/**
* Updates the player rankings leaderboard.
**/
func (s *service) UpdatePlayerRankings(ctx context.Context, updateData *pb.MemberProfileUpdatedEvent) error {
	memberIdUUID, err := uuid.Parse(updateData.MemberId)
	if err != nil {
		slog.Error("error when parsing updateData.MemberId as uuid", "memberID", updateData.MemberId, "error", err)
	}
	_, err = s.repo.UpsertPlayerRankingStats(ctx, &UpdatePlayerRankingsParams{
		MemberID:  memberIdUUID,
		AvatarUrl: updateData.AvatarUrl,
	})

	if err != nil {
		slog.Error("error when updating player rankings", "memberID", updateData.MemberId, "error", err)
		return err
	}

	return nil
}

/**
* Grabs leaderboard related data from the player rankings table.
**/

func (s *service) GetLeaderboard(ctx context.Context, req *pbstats.GetLeaderboardRequest) (*pbstats.GetLeaderboardResponse, error) {
	limit := 50
	offset := 0

	if req.Limit != nil {
		limit = int(*req.Limit)
	}

	if req.Offset != nil {
		offset = int(*req.Offset)
	}

	version, err := s.cache.Get(ctx, commoncache.StatsLeaderboardVersionKey())
	var versionInt64 int64

	if err != nil {
		slog.Warn("Cached result for status leaderboard version doesn't exist or retreival failed.", "key", commoncache.StatsLeaderboardVersionKey(), "error", err)

		versionInt64 = int64(1) // default to 1
	} else {
		// cache exists, parse into int64
		versionInt64, _ = strconv.ParseInt(version, 10, 64)
	}

	key := commoncache.StatsLeaderboardKey(versionInt64, limit, offset)
	cachedResStr, err := s.cache.Get(ctx, key)

	if err == nil && cachedResStr != "" {
		// -- cache exists, return cached data --
		slog.Info("Getleaderboard cached result", "cachedRes", cachedResStr)

		var cachedResProto pbstats.GetLeaderboardResponse
		err := proto.Unmarshal([]byte(cachedResStr), &cachedResProto)

		if err == nil {
			slog.Info("cache hit", "cachedResProto", cachedResProto)
			return &cachedResProto, nil
		}

		slog.Warn("Error when attempting to unmarshal proto", "error", err)
		// back to db fetch after this point
	}

	// -- cache stale / invalid, pull from repo --
	slog.Warn("Cached result doesn't exist or retreival failed.", "key", key, "error", err)

	params := GetPlayerRankings{
		limit:  limit,
		offset: offset,
	}

	playerRankings, err := s.repo.GetPlayerRankings(ctx, &params)

	if err != nil {
		return nil, err
	}

	playerRankingsProto := make([]*pbstats.PlayerRankingStats, len(playerRankings))

	for index, playerRanking := range playerRankings {
		playerRankingProto := &pbstats.PlayerRankingStats{
			Id:        playerRanking.ID.String(),
			Wins:      int32(playerRanking.Wins),
			Username:  playerRanking.Username,
			TopThrees: int32(playerRanking.TopThrees),
			AvatarUrl: playerRanking.AvatarUrl,
			Rating:    int32(playerRanking.Rating),
		}
		playerRankingsProto[index] = playerRankingProto
	}

	res := pbstats.GetLeaderboardResponse{
		Players: playerRankingsProto,
	}

	// cache results
	go func() {
		protoRes, err := proto.Marshal(&res)
		if err != nil {
			slog.Warn("could not marshal result for caching, caching failed", "key", key, "error", err)
			return
		}

		s.cache.Set(context.Background(), key, protoRes, time.Hour*1)
	}()

	return &res, nil
}
