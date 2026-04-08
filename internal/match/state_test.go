package match

import (
	"errors"
	"testing"

	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

func TestValidateMove(t *testing.T) {
	tests := []struct {
		name     string
		state    types.GameState
		userID   string
		position int
		wantErr  error
	}{
		{
			name: "valid move",
			state: types.GameState{
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
			state: types.GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user2",
			position: 0,
			wantErr:  types.ErrNotYourTurn,
		},
		{
			name: "position out of range negative",
			state: types.GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: -1,
			wantErr:  types.ErrInvalidPosition,
		},
		{
			name: "position out of range high",
			state: types.GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: 9,
			wantErr:  types.ErrInvalidPosition,
		},
		{
			name: "position taken",
			state: types.GameState{
				Board:       [9]string{"X", "", "", "", "", "", "", "", ""},
				Players:     map[string]string{"user1": "X", "user2": "O"},
				CurrentTurn: "user1",
				Status:      "playing",
			},
			userID:   "user1",
			position: 0,
			wantErr:  types.ErrPositionTaken,
		},
		{
			name: "match not started",
			state: types.GameState{
				Board:       [9]string{},
				Players:     map[string]string{"user1": "X"},
				CurrentTurn: "",
				Status:      "waiting",
			},
			userID:   "user1",
			position: 0,
			wantErr:  types.ErrMatchNotStarted,
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
