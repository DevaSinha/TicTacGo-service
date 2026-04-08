package match

import (
	"fmt"

	"github.com/yourusername/lila-tictactoe-server/pkg/types"
)

func ValidateMove(state types.GameState, userID string, position int) error {
	if state.Status != "playing" {
		return fmt.Errorf("match not in playing state: %w", types.ErrMatchNotStarted)
	}

	if state.CurrentTurn != userID {
		return fmt.Errorf("user %s attempted move out of turn: %w", userID, types.ErrNotYourTurn)
	}

	if position < 0 || position > 8 {
		return fmt.Errorf("position %d out of range: %w", position, types.ErrInvalidPosition)
	}

	if state.Board[position] != "" {
		return fmt.Errorf("position %d already occupied: %w", position, types.ErrPositionTaken)
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
