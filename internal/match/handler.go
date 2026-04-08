package match

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/internal/leaderboard"
	"github.com/yourusername/lila-tictactoe-server/internal/storage"
	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

type Handler struct{}

func (m *Handler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	mode := "classic"
	if mv, ok := params["mode"].(string); ok && (mv == "classic" || mv == "timed") {
		mode = mv
	}

	timer := 0
	if mode == "timed" {
		timer = 30
	}

	state := &types.GameState{
		Board:   [9]string{},
		Players: make(map[string]string),
		Status:  "waiting",
		Winner:  "",
		Timer:   timer,
		Mode:    mode,
	}

	tickRate := 1
	label := fmt.Sprintf(`{"mode":"%s"}`, mode)

	return state, tickRate, label
}

func (m *Handler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	gameState, ok := state.(*types.GameState)
	if !ok {
		return state, false, "invalid match state"
	}

	if gameState.Status == "finished" {
		return state, false, types.ErrMatchFull.Error()
	}

	if len(gameState.Players) >= 2 {
		return state, false, types.ErrMatchFull.Error()
	}

	return state, true, ""
}

func (m *Handler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	gameState, ok := state.(*types.GameState)
	if !ok {
		return state
	}

	for _, p := range presences {
		userID := p.GetUserId()
		if len(gameState.Players) == 0 {
			gameState.Players[userID] = "X"
		} else if len(gameState.Players) == 1 {
			if _, exists := gameState.Players[userID]; !exists {
				gameState.Players[userID] = "O"
			}
		}
	}

	if len(gameState.Players) == 2 {
		gameState.Status = "playing"
		for userID, mark := range gameState.Players {
			if mark == "X" {
				gameState.CurrentTurn = userID
				break
			}
		}

		stateData, err := json.Marshal(gameState)
		if err == nil {
			dispatcher.BroadcastMessage(types.OpCodeStateUpdate, stateData, nil, nil, true)
		}
	}

	return gameState
}

func (m *Handler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	gameState, ok := state.(*types.GameState)
	if !ok {
		return nil
	}

	for _, p := range presences {
		leavingUserID := p.GetUserId()

		if gameState.Status == "playing" {
			var winnerID string
			for userID := range gameState.Players {
				if userID != leavingUserID {
					winnerID = userID
					break
				}
			}

			if winnerID != "" {
				gameState.Status = "finished"
				gameState.Winner = winnerID

				account, err := nk.AccountGetId(ctx, winnerID)
				if err == nil && account != nil {
					_ = leaderboard.RecordWin(ctx, nk, winnerID, account.User.Username)
				}

				updatePlayerStatsOnGameEnd(ctx, nk, winnerID, leavingUserID)

				overData, err := json.Marshal(gameState)
				if err == nil {
					dispatcher.BroadcastMessage(types.OpCodeOpponentLeft, overData, nil, nil, true)
				}
			}
		}
	}

	return nil
}

func (m *Handler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	gameState, ok := state.(*types.GameState)
	if !ok {
		return nil
	}

	if gameState.Status != "playing" {
		return gameState
	}

	if gameState.Mode == "timed" {
		gameState.Timer--

		timerData, err := json.Marshal(map[string]int{"timer": gameState.Timer})
		if err == nil {
			dispatcher.BroadcastMessage(types.OpCodeTimerUpdate, timerData, nil, nil, true)
		}

		if gameState.Timer <= 0 {
			var winnerID string
			var loserID string
			for userID := range gameState.Players {
				if userID != gameState.CurrentTurn {
					winnerID = userID
				} else {
					loserID = userID
				}
			}

			gameState.Status = "finished"
			gameState.Winner = winnerID

			account, err := nk.AccountGetId(ctx, winnerID)
			if err == nil && account != nil {
				_ = leaderboard.RecordWin(ctx, nk, winnerID, account.User.Username)
			}

			updatePlayerStatsOnGameEnd(ctx, nk, winnerID, loserID)

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(types.OpCodeGameOver, overData, nil, nil, true)
			}

			return nil
		}
	}

	for _, msg := range messages {
		if msg.GetOpCode() != types.OpCodeMove {
			continue
		}

		var move types.MoveMessage
		if err := json.Unmarshal(msg.GetData(), &move); err != nil {
			continue
		}

		userID := msg.GetUserId()

		if err := ValidateMove(*gameState, userID, move.Position); err != nil {
			continue
		}

		mark := gameState.Players[userID]
		gameState.Board[move.Position] = mark

		result := CheckWinner(gameState.Board)

		if result == "X" || result == "O" {
			var winnerID, loserID string
			for uid, mv := range gameState.Players {
				if mv == result {
					winnerID = uid
				} else {
					loserID = uid
				}
			}

			gameState.Status = "finished"
			gameState.Winner = winnerID

			account, err := nk.AccountGetId(ctx, winnerID)
			if err == nil && account != nil {
				_ = leaderboard.RecordWin(ctx, nk, winnerID, account.User.Username)
			}

			updatePlayerStatsOnGameEnd(ctx, nk, winnerID, loserID)

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(types.OpCodeGameOver, overData, nil, nil, true)
			}

			return nil
		}

		if result == "draw" {
			gameState.Status = "finished"
			gameState.Winner = "draw"

			for uid := range gameState.Players {
				stats, _ := storage.ReadPlayerStats(ctx, nk, uid)
				stats.WinStreak = 0
				_ = storage.WritePlayerStats(ctx, nk, uid, &stats)
			}

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(types.OpCodeGameOver, overData, nil, nil, true)
			}

			return nil
		}

		for uid := range gameState.Players {
			if uid != userID {
				gameState.CurrentTurn = uid
				break
			}
		}

		if gameState.Mode == "timed" {
			gameState.Timer = 30
		}

		stateData, err := json.Marshal(gameState)
		if err == nil {
			dispatcher.BroadcastMessage(types.OpCodeStateUpdate, stateData, nil, nil, true)
		}
	}

	return gameState
}

func (m *Handler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return nil
}

func (m *Handler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

func updatePlayerStatsOnGameEnd(ctx context.Context, nk runtime.NakamaModule, winnerID string, loserID string) {
	winnerStats, _ := storage.ReadPlayerStats(ctx, nk, winnerID)
	winnerStats.Wins++
	winnerStats.WinStreak++
	if winnerStats.WinStreak > winnerStats.BestStreak {
		winnerStats.BestStreak = winnerStats.WinStreak
	}
	_ = storage.WritePlayerStats(ctx, nk, winnerID, &winnerStats)

	loserStats, _ := storage.ReadPlayerStats(ctx, nk, loserID)
	loserStats.Losses++
	loserStats.WinStreak = 0
	_ = storage.WritePlayerStats(ctx, nk, loserID, &loserStats)
}
