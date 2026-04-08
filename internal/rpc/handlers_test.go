package rpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

type mockNakamaModule struct {
	runtime.NakamaModule
	StorageData       map[string]string
	LeaderboardWrites []string
}

func (m *mockNakamaModule) StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error) {
	if m.StorageData == nil {
		return nil, nil
	}
	var results []*api.StorageObject
	for _, r := range reads {
		if val, ok := m.StorageData[r.Key]; ok {
			results = append(results, &api.StorageObject{
				Collection: r.Collection,
				Key:        r.Key,
				UserId:     r.UserID,
				Value:      val,
			})
		}
	}
	return results, nil
}

func (m *mockNakamaModule) StorageWrite(ctx context.Context, writes []*runtime.StorageWrite) ([]*api.StorageObjectAck, error) {
	if m.StorageData == nil {
		m.StorageData = make(map[string]string)
	}
	for _, w := range writes {
		m.StorageData[w.Key] = w.Value
	}
	return nil, nil
}

func (m *mockNakamaModule) LeaderboardRecordsList(ctx context.Context, id string, ownerIDs []string, limit int, cursor string, expiry int64) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, string, error) {
	return []*api.LeaderboardRecord{}, nil, "", "", nil
}

func (m *mockNakamaModule) MatchCreate(ctx context.Context, module string, params map[string]interface{}) (string, error) {
	return "match-id-123", nil
}

func (m *mockNakamaModule) MatchSignal(ctx context.Context, matchID string, data string) (string, error) {
	if matchID == "valid-match" {
		return "success", nil
	}
	return "error", nil
}

func TestRpcCreateMatch(t *testing.T) {
	nk := &mockNakamaModule{}
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
	nk := &mockNakamaModule{
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

		var stats types.PlayerStats
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

		var stats types.PlayerStats
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
	nk := &mockNakamaModule{}
	ctx := context.Background()

	result, err := RpcGetLeaderboard(ctx, nil, nil, nk, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var records []types.LeaderboardRecord
	if err := json.Unmarshal([]byte(result), &records); err != nil {
		t.Fatalf("failed to unmarshal leaderboard: %v", err)
	}

	if records == nil {
		t.Error("expected non-nil records slice")
	}
}
