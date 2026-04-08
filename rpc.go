package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

func RpcCreateMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	params := map[string]interface{}{
		"mode": "classic",
	}

	if payload != "" {
		var input map[string]string
		if err := json.Unmarshal([]byte(payload), &input); err == nil {
			if mode, ok := input["mode"]; ok && (mode == "classic" || mode == "timed") {
				params["mode"] = mode
			}
		}
	}

	matchID, err := nk.MatchCreate(ctx, "tictactoe", params)
	if err != nil {
		return "", fmt.Errorf("failed to create match: %w", err)
	}

	response, err := json.Marshal(map[string]string{"match_id": matchID})
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(response), nil
}

func RpcGetPlayerStats(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context: %w", runtime.NewError("unauthenticated", 16))
	}

	stats, err := ReadPlayerStats(ctx, nk, userID)
	if err != nil {
		return "", fmt.Errorf("failed to read player stats: %w", err)
	}

	response, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal player stats: %w", err)
	}

	return string(response), nil
}

func RpcGetLeaderboard(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	records, err := GetTopLeaderboard(ctx, nk, 10)
	if err != nil {
		return "", fmt.Errorf("failed to get leaderboard: %w", err)
	}

	var results []LeaderboardRecord
	for _, r := range records {
		stats, _ := ReadPlayerStats(ctx, nk, r.OwnerId)
		results = append(results, LeaderboardRecord{
			UserID:   r.OwnerId,
			Username: r.Username.Value,
			Score:    int(r.Score),
			Wins:     stats.Wins,
			Losses:   stats.Losses,
			Streak:   stats.WinStreak,
		})
	}

	if results == nil {
		results = []LeaderboardRecord{}
	}

	response, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal leaderboard: %w", err)
	}

	return string(response), nil
}

func RpcForfeitMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user ID not found in context")
	}

	var input map[string]string
	if err := json.Unmarshal([]byte(payload), &input); err != nil {
		return "", err
	}
	matchID := input["match_id"]
	if matchID == "" {
		return "", fmt.Errorf("match_id required")
	}

	response, err := nk.MatchSignal(ctx, matchID, fmt.Sprintf(`{"action":"forfeit","userId":"%s"}`, userID))
	if err != nil {
		return "", fmt.Errorf("failed to signal match: %w", err)
	}

	if response == "error" || response == "invalid" {
		return "", fmt.Errorf("cannot forfeit match")
	}

	return "{}", nil
}
