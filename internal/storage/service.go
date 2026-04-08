package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

const (
	statsCollection = "player_stats"
)

func ReadPlayerStats(ctx context.Context, nk runtime.NakamaModule, userID string) (types.PlayerStats, error) {
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{
		{
			Collection: statsCollection,
			Key:        userID,
			UserID:     userID,
		},
	})
	if err != nil {
		// Treat missing record or errors as a zero-value PlayerStats{}
		return types.PlayerStats{}, nil
	}

	if len(objects) == 0 {
		return types.PlayerStats{}, nil
	}

	var stats types.PlayerStats
	if err := json.Unmarshal([]byte(objects[0].Value), &stats); err != nil {
		return types.PlayerStats{}, nil
	}

	return stats, nil
}

func WritePlayerStats(ctx context.Context, nk runtime.NakamaModule, userID string, stats *types.PlayerStats) error {
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
