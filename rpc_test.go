package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func TestRpcCreateMatch(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "empty payload",
			payload: "",
		},
		{
			name:    "classic mode",
			payload: `{"mode":"classic"}`,
		},
		{
			name:    "timed mode",
			payload: `{"mode":"timed"}`,
		},
		{
			name:    "invalid mode defaults",
			payload: `{"mode":"invalid"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RpcCreateMatch(ctx, nil, nil, nk, tt.payload)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var response map[string]string
			if err := json.Unmarshal([]byte(result), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response["match_id"] == "" {
				t.Error("expected non-empty match_id")
			}
		})
	}
}

func TestRpcGetPlayerStats(t *testing.T) {
	nk := &MockNakamaModule{
		StorageData: map[string]string{
			"user1": `{"wins":5,"losses":3,"win_streak":2,"best_streak":4}`,
		},
	}

	t.Run("existing player", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "user1")

		result, err := RpcGetPlayerStats(ctx, nil, nil, nk, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var stats PlayerStats
		if err := json.Unmarshal([]byte(result), &stats); err != nil {
			t.Fatalf("failed to unmarshal stats: %v", err)
		}

		if stats.Wins != 5 {
			t.Errorf("expected 5 wins, got %d", stats.Wins)
		}

		if stats.Losses != 3 {
			t.Errorf("expected 3 losses, got %d", stats.Losses)
		}
	})

	t.Run("new player", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "newuser")

		result, err := RpcGetPlayerStats(ctx, nil, nil, nk, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var stats PlayerStats
		if err := json.Unmarshal([]byte(result), &stats); err != nil {
			t.Fatalf("failed to unmarshal stats: %v", err)
		}

		if stats.Wins != 0 || stats.Losses != 0 {
			t.Error("expected zeroed stats for new player")
		}
	})

	t.Run("no user id", func(t *testing.T) {
		ctx := context.Background()

		_, err := RpcGetPlayerStats(ctx, nil, nil, nk, "")
		if err == nil {
			t.Error("expected error when no user ID in context")
		}
	})
}

func TestRpcGetLeaderboard(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	result, err := RpcGetLeaderboard(ctx, nil, nil, nk, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var records []LeaderboardRecord
	if err := json.Unmarshal([]byte(result), &records); err != nil {
		t.Fatalf("failed to unmarshal leaderboard: %v", err)
	}

	if records == nil {
		t.Error("expected non-nil records slice")
	}
}
