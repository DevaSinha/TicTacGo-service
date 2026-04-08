package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

type MatchHandler struct{}

func (m *MatchHandler) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	mode := "classic"
	if m, ok := params["mode"].(string); ok && (m == "classic" || m == "timed") {
		mode = m
	}

	timer := 0
	if mode == "timed" {
		timer = 30
	}

	state := &GameState{
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

func (m *MatchHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	gameState, ok := state.(*GameState)
	if !ok {
		return state, false, "invalid match state"
	}

	if gameState.Status == "finished" {
		return state, false, ErrMatchFull.Error()
	}

	if len(gameState.Players) >= 2 {
		return state, false, ErrMatchFull.Error()
	}

	return state, true, ""
}

func (m *MatchHandler) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	gameState, ok := state.(*GameState)
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
			dispatcher.BroadcastMessage(OpCodeStateUpdate, stateData, nil, nil, true)
		}
	}

	return gameState
}

func (m *MatchHandler) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	gameState, ok := state.(*GameState)
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
					_ = RecordWin(ctx, nk, winnerID, account.User.Username)
				}

				updatePlayerStatsOnGameEnd(ctx, nk, winnerID, leavingUserID)

				overData, err := json.Marshal(gameState)
				if err == nil {
					dispatcher.BroadcastMessage(OpCodeOpponentLeft, overData, nil, nil, true)
				}
			}
		}
	}

	return nil
}

func (m *MatchHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	gameState, ok := state.(*GameState)
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
			dispatcher.BroadcastMessage(OpCodeTimerUpdate, timerData, nil, nil, true)
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
				_ = RecordWin(ctx, nk, winnerID, account.User.Username)
			}

			updatePlayerStatsOnGameEnd(ctx, nk, winnerID, loserID)

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(OpCodeGameOver, overData, nil, nil, true)
			}

			return nil
		}
	}

	for _, msg := range messages {
		if msg.GetOpCode() != OpCodeMove {
			continue
		}

		var move MoveMessage
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
			for uid, m := range gameState.Players {
				if m == result {
					winnerID = uid
				} else {
					loserID = uid
				}
			}

			gameState.Status = "finished"
			gameState.Winner = winnerID

			account, err := nk.AccountGetId(ctx, winnerID)
			if err == nil && account != nil {
				_ = RecordWin(ctx, nk, winnerID, account.User.Username)
			}

			updatePlayerStatsOnGameEnd(ctx, nk, winnerID, loserID)

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(OpCodeGameOver, overData, nil, nil, true)
			}

			return nil
		}

		if result == "draw" {
			gameState.Status = "finished"
			gameState.Winner = "draw"

			for uid := range gameState.Players {
				stats, _ := ReadPlayerStats(ctx, nk, uid)
				stats.WinStreak = 0
				_ = WritePlayerStats(ctx, nk, uid, &stats)
			}

			overData, err := json.Marshal(gameState)
			if err == nil {
				dispatcher.BroadcastMessage(OpCodeGameOver, overData, nil, nil, true)
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
			dispatcher.BroadcastMessage(OpCodeStateUpdate, stateData, nil, nil, true)
		}
	}

	return gameState
}

func (m *MatchHandler) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	return nil
}

func (m *MatchHandler) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	return state, ""
}

func ValidateMove(state GameState, userID string, position int) error {
	if state.Status != "playing" {
		return fmt.Errorf("match not in playing state: %w", ErrMatchNotStarted)
	}

	if state.CurrentTurn != userID {
		return fmt.Errorf("user %s attempted move out of turn: %w", userID, ErrNotYourTurn)
	}

	if position < 0 || position > 8 {
		return fmt.Errorf("position %d out of range: %w", position, ErrInvalidPosition)
	}

	if state.Board[position] != "" {
		return fmt.Errorf("position %d already occupied: %w", position, ErrPositionTaken)
	}

	return nil
}

func CheckWinner(board [9]string) string {
	winPatterns := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
	}

	for _, p := range winPatterns {
		if board[p[0]] != "" && board[p[0]] == board[p[1]] && board[p[1]] == board[p[2]] {
			return board[p[0]]
		}
	}

	for _, cell := range board {
		if cell == "" {
			return ""
		}
	}

	return "draw"
}

func updatePlayerStatsOnGameEnd(ctx context.Context, nk runtime.NakamaModule, winnerID string, loserID string) {
	winnerStats, _ := ReadPlayerStats(ctx, nk, winnerID)
	winnerStats.Wins++
	winnerStats.WinStreak++
	if winnerStats.WinStreak > winnerStats.BestStreak {
		winnerStats.BestStreak = winnerStats.WinStreak
	}
	_ = WritePlayerStats(ctx, nk, winnerID, &winnerStats)

	loserStats, _ := ReadPlayerStats(ctx, nk, loserID)
	loserStats.Losses++
	loserStats.WinStreak = 0
	_ = WritePlayerStats(ctx, nk, loserID, &loserStats)
}
