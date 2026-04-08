package match

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

func TestMatchInit(t *testing.T) {
	handler := &Handler{}

	tests := []struct {
		name       string
		params     map[string]interface{}
		expectMode string
	}{
		{
			name:       "classic mode default",
			params:     map[string]interface{}{},
			expectMode: "classic",
		},
		{
			name:       "classic mode explicit",
			params:     map[string]interface{}{"mode": "classic"},
			expectMode: "classic",
		},
		{
			name:       "timed mode",
			params:     map[string]interface{}{"mode": "timed"},
			expectMode: "timed",
		},
		{
			name:       "invalid mode defaults to classic",
			params:     map[string]interface{}{"mode": "invalid"},
			expectMode: "classic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, tickRate, label := handler.MatchInit(context.Background(), nil, nil, nil, tt.params)

			gs, ok := state.(*types.GameState)
			if !ok {
				t.Fatal("state is not *types.GameState")
			}

			if gs.Mode != tt.expectMode {
				t.Errorf("expected mode %q, got %q", tt.expectMode, gs.Mode)
			}

			if gs.Status != "waiting" {
				t.Errorf("expected status 'waiting', got %q", gs.Status)
			}

			if tickRate != 1 {
				t.Errorf("expected tick rate 1, got %d", tickRate)
			}

			if label == "" {
				t.Error("expected non-empty label")
			}

			if tt.expectMode == "timed" && gs.Timer != 30 {
				t.Errorf("expected timer 30 for timed mode, got %d", gs.Timer)
			}

			if tt.expectMode == "classic" && gs.Timer != 0 {
				t.Errorf("expected timer 0 for classic mode, got %d", gs.Timer)
			}
		})
	}
}

func TestMatchJoinAttempt(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}

	tests := []struct {
		name       string
		state      *types.GameState
		wantAccept bool
	}{
		{
			name: "valid join first player",
			state: &types.GameState{
				Players: map[string]string{},
				Status:  "waiting",
			},
			wantAccept: true,
		},
		{
			name: "valid join second player",
			state: &types.GameState{
				Players: map[string]string{"user1": "X"},
				Status:  "waiting",
			},
			wantAccept: true,
		},
		{
			name: "reject full match",
			state: &types.GameState{
				Players: map[string]string{"user1": "X", "user2": "O"},
				Status:  "playing",
			},
			wantAccept: false,
		},
		{
			name: "reject finished match",
			state: &types.GameState{
				Players: map[string]string{"user1": "X"},
				Status:  "finished",
			},
			wantAccept: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presence := &mockPresence{UserID: "newuser"}
			_, accepted, _ := handler.MatchJoinAttempt(context.Background(), nil, nil, nil, dispatcher, 0, tt.state, presence, nil)

			if accepted != tt.wantAccept {
				t.Errorf("expected accepted=%v, got %v", tt.wantAccept, accepted)
			}
		})
	}
}

func TestMatchJoin(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}

	t.Run("first player joins", func(t *testing.T) {
		state := &types.GameState{
			Players: make(map[string]string),
			Status:  "waiting",
			Mode:    "classic",
		}

		presences := []runtime.Presence{
			&mockPresence{UserID: "user1"},
		}

		result := handler.MatchJoin(context.Background(), nil, nil, nil, dispatcher, 0, state, presences)
		gs := result.(*types.GameState)

		if gs.Players["user1"] != "X" {
			t.Errorf("expected user1 to be X, got %q", gs.Players["user1"])
		}

		if gs.Status != "waiting" {
			t.Errorf("expected status 'waiting', got %q", gs.Status)
		}
	})

	t.Run("second player joins starts game", func(t *testing.T) {
		state := &types.GameState{
			Players: map[string]string{"user1": "X"},
			Status:  "waiting",
			Mode:    "classic",
		}

		presences := []runtime.Presence{
			&mockPresence{UserID: "user2"},
		}

		dispatcher.BroadcastCount = 0
		result := handler.MatchJoin(context.Background(), nil, nil, nil, dispatcher, 0, state, presences)
		gs := result.(*types.GameState)

		if gs.Players["user2"] != "O" {
			t.Errorf("expected user2 to be O, got %q", gs.Players["user2"])
		}

		if gs.Status != "playing" {
			t.Errorf("expected status 'playing', got %q", gs.Status)
		}

		if gs.CurrentTurn != "user1" {
			t.Errorf("expected current turn to be user1, got %q", gs.CurrentTurn)
		}

		if dispatcher.BroadcastCount == 0 {
			t.Error("expected state broadcast on game start")
		}

		if dispatcher.BroadcastedOpCode != types.OpCodeStateUpdate {
			t.Errorf("expected OpCodeStateUpdate broadcast, got %d", dispatcher.BroadcastedOpCode)
		}
	})
}

func TestMatchLoop_ValidMove(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 4})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*types.GameState)

	if gs.Board[4] != "X" {
		t.Errorf("expected board[4] to be X, got %q", gs.Board[4])
	}

	if gs.CurrentTurn != "user2" {
		t.Errorf("expected current turn to switch to user2, got %q", gs.CurrentTurn)
	}
}

func TestMatchLoop_WrongTurn(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user2",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*types.GameState)

	if gs.Board[0] != "" {
		t.Error("move should not have been applied for wrong turn")
	}
}

func TestMatchLoop_OutOfRange(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 10})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*types.GameState)

	if gs.CurrentTurn != "user1" {
		t.Error("turn should not have changed for invalid position")
	}
}

func TestMatchLoop_PositionTaken(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{"O", "", "", "", "", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*types.GameState)

	if gs.Board[0] != "O" {
		t.Error("position should remain unchanged when already taken")
	}
}

func TestMatchLoop_WinDetection(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{"X", "X", "", "O", "O", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 2})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)

	if result != nil {
		t.Error("expected nil return on game over (match termination)")
	}

	if state.Status != "finished" {
		t.Errorf("expected status 'finished', got %q", state.Status)
	}

	if state.Winner != "user1" {
		t.Errorf("expected winner 'user1', got %q", state.Winner)
	}

	if dispatcher.BroadcastedOpCode != types.OpCodeGameOver {
		t.Errorf("expected OpCodeGameOver broadcast, got %d", dispatcher.BroadcastedOpCode)
	}
}

func TestMatchLoop_DrawDetection(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{"X", "O", "X", "X", "O", "O", "O", "X", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 8})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)

	if result != nil {
		t.Error("expected nil return on draw (match termination)")
	}

	if state.Winner != "draw" {
		t.Errorf("expected winner 'draw', got %q", state.Winner)
	}
}

func TestMatchLoop_TimedModeTimeout(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "timed",
		Timer:       1,
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, nil)

	if result != nil {
		t.Error("expected nil return on timeout (match termination)")
	}

	if state.Status != "finished" {
		t.Errorf("expected status 'finished', got %q", state.Status)
	}

	if state.Winner == "user1" {
		t.Error("current turn player should not win on timeout")
	}

	if state.Winner != "user2" {
		t.Errorf("expected user2 to win on timeout, got %q", state.Winner)
	}
}

func TestMatchTerminate(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}

	state := &types.GameState{
		Board:   [9]string{},
		Players: map[string]string{"user1": "X", "user2": "O"},
		Status:  "playing",
		Mode:    "classic",
	}

	result := handler.MatchTerminate(context.Background(), nil, nil, nil, dispatcher, 0, state, 10)

	if result != nil {
		t.Error("expected nil return from MatchTerminate")
	}
}

func TestMatchSignal(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}

	state := &types.GameState{
		Board:   [9]string{},
		Players: map[string]string{},
		Status:  "waiting",
	}

	resultState, resultData := handler.MatchSignal(context.Background(), nil, nil, nil, dispatcher, 0, state, "test")

	if resultState == nil {
		t.Error("expected non-nil state from MatchSignal")
	}

	if resultData != "" {
		t.Errorf("expected empty data from MatchSignal, got %q", resultData)
	}
}

func TestMatchLeave_Forfeit(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{"X", "O", "", "", "", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	presences := []runtime.Presence{
		&mockPresence{UserID: "user1"},
	}

	result := handler.MatchLeave(context.Background(), nil, nil, nk, dispatcher, 0, state, presences)

	if result != nil {
		t.Error("expected nil return on match leave (termination)")
	}

	if state.Status != "finished" {
		t.Errorf("expected status 'finished', got %q", state.Status)
	}

	if state.Winner != "user2" {
		t.Errorf("expected winner 'user2', got %q", state.Winner)
	}

	if dispatcher.BroadcastedOpCode != types.OpCodeOpponentLeft {
		t.Errorf("expected OpCodeOpponentLeft broadcast, got %d", dispatcher.BroadcastedOpCode)
	}
}

func TestMatchLeave_WaitingState(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:   [9]string{},
		Players: map[string]string{"user1": "X"},
		Status:  "waiting",
		Mode:    "classic",
	}

	presences := []runtime.Presence{
		&mockPresence{UserID: "user1"},
	}

	result := handler.MatchLeave(context.Background(), nil, nil, nk, dispatcher, 0, state, presences)

	if result != nil {
		t.Error("expected nil return on match leave")
	}

	if state.Status != "waiting" {
		t.Errorf("expected status to remain 'waiting', got %q", state.Status)
	}
}

func TestMatchLoop_TimedModeTimerReset(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "timed",
		Timer:       15,
	}

	moveData, _ := json.Marshal(types.MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&mockMatchData{
			UserIDVal: "user1",
			OpCodeVal: types.OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*types.GameState)

	if gs.Timer != 30 {
		t.Errorf("expected timer to reset to 30 after move in timed mode, got %d", gs.Timer)
	}
}

func TestMatchLoop_NoMessages(t *testing.T) {
	handler := &Handler{}
	dispatcher := &mockDispatcher{}
	nk := &mockNakamaModule{}

	state := &types.GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, nil)

	if result == nil {
		t.Error("expected non-nil state when no messages")
	}
}
