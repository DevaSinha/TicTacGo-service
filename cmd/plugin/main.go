package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/internal/leaderboard"
	internalmatch "github.com/yourusername/lila-tictactoe-server/internal/match"
	"github.com/yourusername/lila-tictactoe-server/internal/rpc"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("tictactoe", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &internalmatch.Handler{}, nil
	}); err != nil {
		return err
	}

	if err := initializer.RegisterMatchmakerMatched(MatchmakerMatched); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("create_match", rpc.RpcCreateMatch); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_player_stats", rpc.RpcGetPlayerStats); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_leaderboard", rpc.RpcGetLeaderboard); err != nil {
		return err
	}

	if err := leaderboard.CreateLeaderboard(ctx, nk); err != nil {
		return err
	}

	logger.Info("tictactoe module loaded")
	return nil
}

func MatchmakerMatched(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
	params := map[string]interface{}{
		"mode": "classic",
	}

	if len(entries) > 0 {
		props := entries[0].GetProperties()
		if modeStr, ok := props["mode"].(string); ok {
			params["mode"] = modeStr
		}
	}

	matchID, err := nk.MatchCreate(ctx, "tictactoe", params)
	if err != nil {
		return "", err
	}

	return matchID, nil
}
