package storage

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

// mockNakamaModule is a local test double — only storage methods needed here.
type mockNakamaModule struct {
	runtime.NakamaModule
	storageData map[string]string
}

func (m *mockNakamaModule) StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error) {
	if m.storageData == nil {
		return nil, nil
	}
	var results []*api.StorageObject
	for _, r := range reads {
		if val, ok := m.storageData[r.Key]; ok {
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
	if m.storageData == nil {
		m.storageData = make(map[string]string)
	}
	for _, w := range writes {
		m.storageData[w.Key] = w.Value
	}
	return nil, nil
}

func TestReadPlayerStats_NoData(t *testing.T) {
	nk := &mockNakamaModule{}
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
	expected := types.PlayerStats{Wins: 10, Losses: 5, WinStreak: 3, BestStreak: 7}
	data, _ := json.Marshal(expected)

	nk := &mockNakamaModule{
		storageData: map[string]string{
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
	nk := &mockNakamaModule{}
	ctx := context.Background()

	stats := types.PlayerStats{Wins: 3, Losses: 1, WinStreak: 2, BestStreak: 5}

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
	nk := &mockNakamaModule{}
	ctx := context.Background()

	initial := types.PlayerStats{Wins: 1, Losses: 0, WinStreak: 1, BestStreak: 1}
	_ = WritePlayerStats(ctx, nk, "user1", &initial)

	updated := types.PlayerStats{Wins: 2, Losses: 1, WinStreak: 0, BestStreak: 1}
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
