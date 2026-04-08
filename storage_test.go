package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestReadPlayerStats_NoData(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	stats, err := ReadPlayerStats(ctx, nk, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Wins != 0 || stats.Losses != 0 || stats.WinStreak != 0 || stats.BestStreak != 0 {
		t.Error("expected zeroed stats for nonexistent player")
	}
}

func TestReadPlayerStats_ExistingData(t *testing.T) {
	expected := PlayerStats{Wins: 10, Losses: 5, WinStreak: 3, BestStreak: 7}
	data, _ := json.Marshal(expected)

	nk := &MockNakamaModule{
		StorageData: map[string]string{
			"user1": string(data),
		},
	}
	ctx := context.Background()

	stats, err := ReadPlayerStats(ctx, nk, "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Wins != expected.Wins {
		t.Errorf("expected %d wins, got %d", expected.Wins, stats.Wins)
	}

	if stats.BestStreak != expected.BestStreak {
		t.Errorf("expected %d best streak, got %d", expected.BestStreak, stats.BestStreak)
	}
}

func TestWritePlayerStats(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	stats := PlayerStats{Wins: 3, Losses: 1, WinStreak: 2, BestStreak: 5}

	err := WritePlayerStats(ctx, nk, "user1", &stats)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	readBack, err := ReadPlayerStats(ctx, nk, "user1")
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}

	if readBack.Wins != stats.Wins {
		t.Errorf("expected %d wins, got %d", stats.Wins, readBack.Wins)
	}

	if readBack.Losses != stats.Losses {
		t.Errorf("expected %d losses, got %d", stats.Losses, readBack.Losses)
	}

	if readBack.WinStreak != stats.WinStreak {
		t.Errorf("expected %d win streak, got %d", stats.WinStreak, readBack.WinStreak)
	}

	if readBack.BestStreak != stats.BestStreak {
		t.Errorf("expected %d best streak, got %d", stats.BestStreak, readBack.BestStreak)
	}
}

func TestWritePlayerStats_Overwrite(t *testing.T) {
	nk := &MockNakamaModule{}
	ctx := context.Background()

	initial := PlayerStats{Wins: 1, Losses: 0, WinStreak: 1, BestStreak: 1}
	_ = WritePlayerStats(ctx, nk, "user1", &initial)

	updated := PlayerStats{Wins: 2, Losses: 1, WinStreak: 0, BestStreak: 1}
	err := WritePlayerStats(ctx, nk, "user1", &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	readBack, err := ReadPlayerStats(ctx, nk, "user1")
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}

	if readBack.Wins != 2 {
		t.Errorf("expected 2 wins after overwrite, got %d", readBack.Wins)
	}

	if readBack.WinStreak != 0 {
		t.Errorf("expected 0 win streak after loss, got %d", readBack.WinStreak)
	}
}
