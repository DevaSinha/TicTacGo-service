package leaderboard

import (
	"context"
	"fmt"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	leaderboardID = "global_wins"
)

func CreateLeaderboard(ctx context.Context, nk runtime.NakamaModule) error {
	authoritative := false
	sortOrder := "descending"
	operator := "increment"
	resetSchedule := ""

	err := nk.LeaderboardCreate(ctx, leaderboardID, authoritative, sortOrder, operator, resetSchedule, nil)
	if err != nil {
		return fmt.Errorf("failed to create leaderboard: %w", err)
	}

	return nil
}

func RecordWin(ctx context.Context, nk runtime.NakamaModule, userID string, username string) error {
	score := int64(1)
	subscore := int64(0)

	_, err := nk.LeaderboardRecordWrite(ctx, leaderboardID, userID, username, score, subscore, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to write leaderboard record: %w", err)
	}

	return nil
}

func GetTopLeaderboard(ctx context.Context, nk runtime.NakamaModule, limit int) ([]*api.LeaderboardRecord, error) {
	records, _, _, _, err := nk.LeaderboardRecordsList(ctx, leaderboardID, []string{}, limit, "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list leaderboard records: %w", err)
	}

	return records, nil
}
