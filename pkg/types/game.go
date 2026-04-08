package types

type GameState struct {
	Board       [9]string         `json:"board"`
	Players     map[string]string `json:"players"`
	CurrentTurn string            `json:"current_turn"`
	Status      string            `json:"status"`
	Winner      string            `json:"winner"`
	Timer       int               `json:"timer"`
	Mode        string            `json:"mode"`
}

type MoveMessage struct {
	Position int `json:"position"`
}

type LeaderboardRecord struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Streak   int    `json:"streak"`
	Score    int    `json:"score"`
}

type PlayerStats struct {
	Wins       int `json:"wins"`
	Losses     int `json:"losses"`
	WinStreak  int `json:"win_streak"`
	BestStreak int `json:"best_streak"`
}
