package main

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama-common/runtime"
)

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
