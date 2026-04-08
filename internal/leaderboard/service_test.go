package leaderboard

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type mockNakamaModule struct {
	runtime.NakamaModule
	leaderboardWrites []string
}

func (m *mockNakamaModule) LeaderboardCreate(ctx context.Context, id string, authoritative bool, sortOrder string, operator string, resetSchedule string, metadata map[string]interface{}) error {
	return nil
}

func (m *mockNakamaModule) LeaderboardRecordWrite(ctx context.Context, id string, ownerID string, username string, score int64, subscore int64, metadata map[string]interface{}, overrideOperator *int) (*api.LeaderboardRecord, error) {
	if m.leaderboardWrites == nil {
		m.leaderboardWrites = []string{}
	}
	m.leaderboardWrites = append(m.leaderboardWrites, ownerID)
	return nil, nil
}

func (m *mockNakamaModule) LeaderboardRecordsList(ctx context.Context, id string, ownerIDs []string, limit int, cursor string, expiry int64) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, string, error) {
	return []*api.LeaderboardRecord{}, nil, "", "", nil
}

func TestCreateLeaderboard(t *testing.T) {
	nk := &mockNakamaModule{}
	ctx := context.Background()

	err := CreateLeaderboard(ctx, nk)
	if err != nil {
		t.Fatalf("unexpected error creating leaderboard: %v", err)
	}
}

func TestRecordWin(t *testing.T) {
	nk := &mockNakamaModule{}
	ctx := context.Background()

	err := RecordWin(ctx, nk, "user1", "player1")
	if err != nil {
		t.Fatalf("unexpected error recording win: %v", err)
	}

	if len(nk.leaderboardWrites) != 1 {
		t.Errorf("expected 1 leaderboard write, got %d", len(nk.leaderboardWrites))
	}

	if nk.leaderboardWrites[0] != "user1" {
		t.Errorf("expected leaderboard write for user1, got %q", nk.leaderboardWrites[0])
	}
}

func TestRecordWin_Multiple(t *testing.T) {
	nk := &mockNakamaModule{}
	ctx := context.Background()

	_ = RecordWin(ctx, nk, "user1", "player1")
	_ = RecordWin(ctx, nk, "user1", "player1")
	_ = RecordWin(ctx, nk, "user2", "player2")

	if len(nk.leaderboardWrites) != 3 {
		t.Errorf("expected 3 leaderboard writes, got %d", len(nk.leaderboardWrites))
	}
}

func TestGetTopLeaderboard(t *testing.T) {
	nk := &mockNakamaModule{}
	ctx := context.Background()

	records, err := GetTopLeaderboard(ctx, nk, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if records == nil {
		t.Error("expected non-nil records")
	}
}
