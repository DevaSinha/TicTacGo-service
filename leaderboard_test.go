package main

import (
	"context"
	"testing"
)

func TestCreateLeaderboard(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	err := CreateLeaderboard(ctx, nk)
	if err != nil {
		t.Fatalf("unexpected error creating leaderboard: %v", err)
	}
}

func TestRecordWin(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	err := RecordWin(ctx, nk, "user1", "player1")
	if err != nil {
		t.Fatalf("unexpected error recording win: %v", err)
	}

	if len(nk.LeaderboardWrites) != 1 {
		t.Errorf("expected 1 leaderboard write, got %d", len(nk.LeaderboardWrites))
	}

	if nk.LeaderboardWrites[0] != "user1" {
		t.Errorf("expected leaderboard write for user1, got %q", nk.LeaderboardWrites[0])
	}
}

func TestRecordWin_Multiple(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	_ = RecordWin(ctx, nk, "user1", "player1")
	_ = RecordWin(ctx, nk, "user1", "player1")
	_ = RecordWin(ctx, nk, "user2", "player2")

	if len(nk.LeaderboardWrites) != 3 {
		t.Errorf("expected 3 leaderboard writes, got %d", len(nk.LeaderboardWrites))
	}
}

func TestGetTopLeaderboard(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	records, err := GetTopLeaderboard(ctx, nk, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if records == nil {
		t.Error("expected non-nil records")
	}
}
