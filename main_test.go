package main

import (
	"context"
	"database/sql"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func TestInitModule(t *testing.T) {
	initializer := &MockInitializer{}
	nk := &MockNakamaModule{}
	ctx := context.Background()

	err := InitModule(ctx, &MockLogger{}, nil, nk, initializer)
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
	for _, rpc := range expectedRPCs {
		if _, ok := initializer.RegisteredRPCs[rpc]; !ok {
			t.Errorf("expected RPC %q to be registered", rpc)
		}
	}
}

type MockLogger struct {
	runtime.Logger
}

func (l *MockLogger) Info(format string, v ...interface{})  {}
func (l *MockLogger) Warn(format string, v ...interface{})  {}
func (l *MockLogger) Error(format string, v ...interface{}) {}
func (l *MockLogger) Debug(format string, v ...interface{}) {}

type MockInitializer struct {
	runtime.Initializer
	MatchRegistered      bool
	MatchmakerRegistered bool
	RegisteredRPCs       map[string]bool
}

func (i *MockInitializer) RegisterMatch(name string, fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error)) error {
	i.MatchRegistered = true
	return nil
}

func (i *MockInitializer) RegisterMatchmakerMatched(fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error)) error {
	i.MatchmakerRegistered = true
	return nil
}

func (i *MockInitializer) RegisterRpc(id string, fn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error)) error {
	if i.RegisteredRPCs == nil {
		i.RegisteredRPCs = make(map[string]bool)
	}
	i.RegisteredRPCs[id] = true
	return nil
}
