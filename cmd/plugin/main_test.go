package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type mockLogger struct {
	runtime.Logger
}

func (l *mockLogger) Info(format string, v ...interface{})  {}
func (l *mockLogger) Warn(format string, v ...interface{})  {}
func (l *mockLogger) Error(format string, v ...interface{}) {}
func (l *mockLogger) Debug(format string, v ...interface{}) {}

type mockInitializer struct {
	runtime.Initializer
	MatchRegistered      bool
	MatchmakerRegistered bool
	RegisteredRPCs       map[string]bool
}

func (i *mockInitializer) RegisterMatch(name string, fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error)) error {
	i.MatchRegistered = true
	return nil
}

func (i *mockInitializer) RegisterMatchmakerMatched(fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error)) error {
	i.MatchmakerRegistered = true
	return nil
}

func (i *mockInitializer) RegisterRpc(id string, fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error)) error {
	if i.RegisteredRPCs == nil {
		i.RegisteredRPCs = make(map[string]bool)
	}
	i.RegisteredRPCs[id] = true
	return nil
}

type mockNakamaModule struct {
	runtime.NakamaModule
}

func (m *mockNakamaModule) LeaderboardCreate(ctx context.Context, id string, authoritative bool, sortOrder string, operator string, resetSchedule string, metadata map[string]interface{}) error {
	return nil
}

func (m *mockNakamaModule) MatchCreate(ctx context.Context, module string, params map[string]interface{}) (string, error) {
	return "match-id-123", nil
}

func (m *mockNakamaModule) StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error) {
	return nil, nil
}

func TestInitModule(t *testing.T) {
	initializer := &mockInitializer{}
	nk := &mockNakamaModule{}
	ctx := context.Background()

	err := InitModule(ctx, &mockLogger{}, nil, nk, initializer)
	if err != nil {
		t.Fatalf("InitModule failed: %v", err)
	}

	if !initializer.MatchRegistered {
		t.Error("expected match to be registered")
	}

	if !initializer.MatchmakerRegistered {
		t.Error("expected matchmaker to be registered")
	}

	expectedRPCs := []string{"create_match", "get_player_stats", "get_leaderboard"}
	for _, rpcName := range expectedRPCs {
		if _, ok := initializer.RegisteredRPCs[rpcName]; !ok {
			t.Errorf("expected RPC %q to be registered", rpcName)
		}
	}
}
