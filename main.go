package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("tictactoe", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		return &MatchHandler{}, nil
	}); err != nil {
		return err
	}

	if err := initializer.RegisterMatchmakerMatched(MatchmakerMatched); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("create_match", RpcCreateMatch); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_player_stats", RpcGetPlayerStats); err != nil {
		return err
	}

	if err := initializer.RegisterRpc("get_leaderboard", RpcGetLeaderboard); err != nil {
		return err
	}

	if err := CreateLeaderboard(ctx, nk); err != nil {
		return err
	}

	logger.Info("tictactoe module loaded")
	return nil
}
