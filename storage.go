package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	statsCollection = "player_stats"
)

func ReadPlayerStats(ctx context.Context, nk runtime.NakamaModule, userID string) (PlayerStats, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: statsCollection,
			Key:        userID,
			UserID:     userID,
		},
	})
	if err != nil {
		// Treat missing record or errors as a zero-value PlayerStats{}
		return PlayerStats{}, nil
	}

	if len(objects) == 0 {
		return PlayerStats{}, nil
	}

	var stats PlayerStats
	if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
		return PlayerStats{}, nil
	}

	return stats, nil
}

func WritePlayerStats(ctx context.Context, nk runtime.NakamaModule, userID string, stats *PlayerStats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("failed to marshal player stats: %w", err)
	}

	_, err = nk.StorageWrite(ctx, []*runtime.StorageWrite{
		{
			Collection:      statsCollection,
			Key:             userID,
			UserID:          userID,
			Value:           string(data),
			PermissionRead:  2,
			PermissionWrite: 0,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to write player stats: %w", err)
	}

	return nil
}
