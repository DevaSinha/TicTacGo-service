package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type MockPresence struct {
	UserID    string
	SessionID string
	Username  string
	Node      string
}

func (p *MockPresence) GetUserId() string    { return p.UserID }
func (p *MockPresence) GetSessionId() string { return p.SessionID }
func (p *MockPresence) GetNodeId() string    { return p.Node }
func (p *MockPresence) GetHidden() bool      { return false }
func (p *MockPresence) GetPersistence() bool { return true }
func (p *MockPresence) GetUsername() string   { return p.Username }
func (p *MockPresence) GetStatus() string    { return "" }
func (p *MockPresence) GetReason() runtime.PresenceReason {
	return runtime.PresenceReason(0)
}

type MockMatchData struct {
	UserIDVal    string
	SessionIDVal string
	UsernameVal  string
	NodeVal      string
	OpCodeVal    int64
	DataVal      []byte
	ReceiveTime  int64
}

func (m *MockMatchData) GetUserId() string    { return m.UserIDVal }
func (m *MockMatchData) GetSessionId() string { return m.SessionIDVal }
func (m *MockMatchData) GetNodeId() string    { return m.NodeVal }
func (m *MockMatchData) GetHidden() bool      { return false }
func (m *MockMatchData) GetPersistence() bool { return true }
func (m *MockMatchData) GetUsername() string   { return m.UsernameVal }
func (m *MockMatchData) GetStatus() string    { return "" }
func (m *MockMatchData) GetOpCode() int64     { return m.OpCodeVal }
func (m *MockMatchData) GetData() []byte      { return m.DataVal }
func (m *MockMatchData) GetReliable() bool      { return true }
func (m *MockMatchData) GetReceiveTime() int64 { return m.ReceiveTime }
func (m *MockMatchData) GetReason() runtime.PresenceReason {
	return runtime.PresenceReason(0)
}

type MockDispatcher struct {
	BroadcastedOpCode int64
	BroadcastedData   []byte
	BroadcastCount    int
}

func (d *MockDispatcher) BroadcastMessage(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	d.BroadcastedOpCode = opCode
	d.BroadcastedData = data
	d.BroadcastCount++
	return nil
}

func (d *MockDispatcher) BroadcastMessageDeferred(opCode int64, data []byte, presences []runtime.Presence, sender runtime.Presence, reliable bool) error {
	return nil
}

func (d *MockDispatcher) MatchKick(presences []runtime.Presence) error { return nil }

func (d *MockDispatcher) MatchLabelUpdate(label string) error { return nil }

func TestValidateMove(t *testing.T) {
	tests := []struct {
		name     string
		state    GameState
		userID   string
		position int
		wantErr  error
	}{
		{
			name: "valid move",
			state: GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: 0,
			wantErr:  nil,
		},
		{
			name: "not your turn",
			state: GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user2",
			position: 0,
			wantErr:  ErrNotYourTurn,
		},
		{
			name: "position out of range negative",
			state: GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: -1,
			wantErr:  ErrInvalidPosition,
		},
		{
			name: "position out of range high",
			state: GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: 9,
			wantErr:  ErrInvalidPosition,
		},
		{
			name: "position taken",
			state: GameState{
				Board:       [9]string{"X", "", "", "", "", "", "", "", ""},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: 0,
			wantErr:  ErrPositionTaken,
		},
		{
			name: "match not started",
			state: GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X"},
				CurrentTurn: "",
				Status:      "waiting",
			},
			userID:   "user1",
			position: 0,
			wantErr:  ErrMatchNotStarted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMove(tt.state, tt.userID, tt.position)
			if tt.wantErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCheckWinner(t *testing.T) {
	tests := []struct {
		name   string
		board  [9]string
		expect string
	}{
		{
			name:   "row 0 X wins",
			board:  [9]string{"X", "X", "X", "", "", "", "", "", ""},
			expect: "X",
		},
		{
			name:   "row 1 O wins",
			board:  [9]string{"", "", "", "O", "O", "O", "", "", ""},
			expect: "O",
		},
		{
			name:   "row 2 X wins",
			board:  [9]string{"", "", "", "", "", "", "X", "X", "X"},
			expect: "X",
		},
		{
			name:   "col 0 X wins",
			board:  [9]string{"X", "", "", "X", "", "", "X", "", ""},
			expect: "X",
		},
		{
			name:   "col 1 O wins",
			board:  [9]string{"", "O", "", "", "O", "", "", "O", ""},
			expect: "O",
		},
		{
			name:   "col 2 X wins",
			board:  [9]string{"", "", "X", "", "", "X", "", "", "X"},
			expect: "X",
		},
		{
			name:   "diagonal top-left X wins",
			board:  [9]string{"X", "", "", "", "X", "", "", "", "X"},
			expect: "X",
		},
		{
			name:   "diagonal top-right O wins",
			board:  [9]string{"", "", "O", "", "O", "", "O", "", ""},
			expect: "O",
		},
		{
			name:   "draw",
			board:  [9]string{"X", "O", "X", "X", "O", "O", "O", "X", "X"},
			expect: "draw",
		},
		{
			name:   "in progress",
			board:  [9]string{"X", "O", "", "", "", "", "", "", ""},
			expect: "",
		},
		{
			name:   "empty board",
			board:  [9]string{},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckWinner(tt.board)
			if result != tt.expect {
				t.Errorf("expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestMatchInit(t *testing.T) {
	handler := &MatchHandler{}

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

			gs, ok := state.(*GameState)
			if !ok {
				t.Fatal("state is not *GameState")
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
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}

	tests := []struct {
		name       string
		state      *GameState
		wantAccept bool
	}{
		{
			name: "valid join first player",
			state: &GameState{
				Players: map[string]string{},
				Status:  "waiting",
			},
			wantAccept: true,
		},
		{
			name: "valid join second player",
			state: &GameState{
				Players: map[string]string{"user1": "X"},
				Status:  "waiting",
			},
			wantAccept: true,
		},
		{
			name: "reject full match",
			state: &GameState{
				Players: map[string]string{"user1": "X", "user2": "O"},
				Status:  "playing",
			},
			wantAccept: false,
		},
		{
			name: "reject finished match",
			state: &GameState{
				Players: map[string]string{"user1": "X"},
				Status:  "finished",
			},
			wantAccept: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presence := &MockPresence{UserID: "newuser"}
			_, accepted, _ := handler.MatchJoinAttempt(context.Background(), nil, nil, nil, dispatcher, 0, tt.state, presence, nil)

			if accepted != tt.wantAccept {
				t.Errorf("expected accepted=%v, got %v", tt.wantAccept, accepted)
			}
		})
	}
}

func TestMatchJoin(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}

	t.Run("first player joins", func(t *testing.T) {
		state := &GameState{
			Players: make(map[string]string),
			Status:  "waiting",
			Mode:    "classic",
		}

		presences := []runtime.Presence{
			&MockPresence{UserID: "user1"},
		}

		result := handler.MatchJoin(context.Background(), nil, nil, nil, dispatcher, 0, state, presences)
		gs := result.(*GameState)

		if gs.Players["user1"] != "X" {
			t.Errorf("expected user1 to be X, got %q", gs.Players["user1"])
		}

		if gs.Status != "waiting" {
			t.Errorf("expected status 'waiting', got %q", gs.Status)
		}
	})

	t.Run("second player joins starts game", func(t *testing.T) {
		state := &GameState{
			Players: map[string]string{"user1": "X"},
			Status:  "waiting",
			Mode:    "classic",
		}

		presences := []runtime.Presence{
			&MockPresence{UserID: "user2"},
		}

		dispatcher.BroadcastCount = 0
		result := handler.MatchJoin(context.Background(), nil, nil, nil, dispatcher, 0, state, presences)
		gs := result.(*GameState)

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

		if dispatcher.BroadcastedOpCode != OpCodeStateUpdate {
			t.Errorf("expected OpCodeStateUpdate broadcast, got %d", dispatcher.BroadcastedOpCode)
		}
	})
}

func TestMatchLoop_ValidMove(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 4})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*GameState)

	if gs.Board[4] != "X" {
		t.Errorf("expected board[4] to be X, got %q", gs.Board[4])
	}

	if gs.CurrentTurn != "user2" {
		t.Errorf("expected current turn to switch to user2, got %q", gs.CurrentTurn)
	}
}

func TestMatchLoop_WrongTurn(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user2",
			OpCodeVal: OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*GameState)

	if gs.Board[0] != "" {
		t.Error("move should not have been applied for wrong turn")
	}
}

func TestMatchLoop_OutOfRange(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 10})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*GameState)

	if gs.CurrentTurn != "user1" {
		t.Error("turn should not have changed for invalid position")
	}
}

func TestMatchLoop_PositionTaken(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{"O", "", "", "", "", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*GameState)

	if gs.Board[0] != "O" {
		t.Error("position should remain unchanged when already taken")
	}
}

func TestMatchLoop_WinDetection(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{"X", "X", "", "O", "O", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 2})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
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

	if dispatcher.BroadcastedOpCode != OpCodeGameOver {
		t.Errorf("expected OpCodeGameOver broadcast, got %d", dispatcher.BroadcastedOpCode)
	}
}

func TestMatchLoop_DrawDetection(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{"X", "O", "X", "X", "O", "O", "O", "X", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 8})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
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
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
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
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}

	state := &GameState{
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
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}

	state := &GameState{
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

type MockNakamaModule struct {
	runtime.NakamaModule
	StorageData       map[string]string
	LeaderboardWrites []string
}

func (m *MockNakamaModule) StorageRead(ctx context.Context, reads []*runtime.StorageRead) ([]*api.StorageObject, error) {
	if m.StorageData == nil {
		return nil, nil
	}

	var results []*api.StorageObject
	for _, r := range reads {
		if val, ok := m.StorageData[r.Key]; ok {
			results = append(results, &api.StorageObject{
				Collection: r.Collection,
				Key:        r.Key,
				UserId:     r.UserID,
				Value:      val,
			})
		}
	}
	return results, nil
}

func (m *MockNakamaModule) StorageWrite(ctx context.Context, writes []*runtime.StorageWrite) ([]*api.StorageObjectAck, error) {
	if m.StorageData == nil {
		m.StorageData = make(map[string]string)
	}
	for _, w := range writes {
		m.StorageData[w.Key] = w.Value
	}
	return nil, nil
}

func (m *MockNakamaModule) LeaderboardRecordWrite(ctx context.Context, id string, ownerID string, username string, score int64, subscore int64, metadata map[string]interface{}, overrideOperator *int) (*api.LeaderboardRecord, error) {
	if m.LeaderboardWrites == nil {
		m.LeaderboardWrites = []string{}
	}
	m.LeaderboardWrites = append(m.LeaderboardWrites, ownerID)
	return nil, nil
}

func (m *MockNakamaModule) AccountGetId(ctx context.Context, userID string) (*api.Account, error) {
	return &api.Account{
		User: &api.User{
			Id:       userID,
			Username: "player_" + userID,
		},
	}, nil
}

func (m *MockNakamaModule) LeaderboardCreate(ctx context.Context, id string, authoritative bool, sortOrder string, operator string, resetSchedule string, metadata map[string]interface{}) error {
	return nil
}

func (m *MockNakamaModule) MatchCreate(ctx context.Context, module string, params map[string]interface{}) (string, error) {
	return "match-id-123", nil
}

func (m *MockNakamaModule) LeaderboardRecordsList(ctx context.Context, id string, ownerIDs []string, limit int, cursor string, expiry int64) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, string, error) {
	return []*api.LeaderboardRecord{}, nil, "", "", nil
}

func TestMatchLeave_Forfeit(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{"X", "O", "", "", "", "", "", "", ""},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "classic",
	}

	presences := []runtime.Presence{
		&MockPresence{UserID: "user1"},
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

	if dispatcher.BroadcastedOpCode != OpCodeOpponentLeft {
		t.Errorf("expected OpCodeOpponentLeft broadcast, got %d", dispatcher.BroadcastedOpCode)
	}
}

func TestMatchLeave_WaitingState(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:   [9]string{},
		Players: map[string]string{"user1": "X"},
		Status:  "waiting",
		Mode:    "classic",
	}

	presences := []runtime.Presence{
		&MockPresence{UserID: "user1"},
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
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
		Board:       [9]string{},
		Players:     map[string]string{"user1": "X", "user2": "O"},
		CurrentTurn: "user1",
		Status:      "playing",
		Mode:        "timed",
		Timer:       15,
	}

	moveData, _ := json.Marshal(MoveMessage{Position: 0})
	messages := []runtime.MatchData{
		&MockMatchData{
			UserIDVal: "user1",
			OpCodeVal: OpCodeMove,
			DataVal:   moveData,
		},
	}

	result := handler.MatchLoop(context.Background(), nil, nil, nk, dispatcher, 1, state, messages)
	gs := result.(*GameState)

	if gs.Timer != 30 {
		t.Errorf("expected timer to reset to 30 after move in timed mode, got %d", gs.Timer)
	}
}

func TestMatchLoop_NoMessages(t *testing.T) {
	handler := &MatchHandler{}
	dispatcher := &MockDispatcher{}
	nk := &MockNakamaModule{}

	state := &GameState{
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

var _ runtime.NakamaModule = (*MockNakamaModule)(nil)
var _ runtime.Presence = (*MockPresence)(nil)
var _ runtime.MatchData = (*MockMatchData)(nil)
var _ runtime.MatchDispatcher = (*MockDispatcher)(nil)
var _ runtime.Match = (*MatchHandler)(nil)
var _ = (*sql.DB)(nil)
