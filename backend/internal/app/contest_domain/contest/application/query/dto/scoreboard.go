package dto

type Scoreboard struct {
	Heads []*ScoreboardHead `json:"heads"`
	Rows  []*ScoreboardRow  `json:"rows"`
	// Compact tells the frontend to render the compact style (one bar per
	// row with per-item gradient) instead of the per-item table.
	Compact bool `json:"compact"`
}

type ScoreboardHead struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type ScoreboardRow struct {
	Name     string   `json:"name"`
	Username string   `json:"username"`
	Statuses []string `json:"statuses"`
}
