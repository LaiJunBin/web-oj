package service

type ScoreReportItem struct {
	Label       string
	Name        string
	Description string
	Score       float64
	TotalScore  float64
	Status      string
}

type ContestantScoreReport struct {
	// UserID makes the contestant's sheet name deterministic and unique:
	// names and usernames survive neither Excel's sheet-name character rules
	// nor its 31-rune cap intact, so the id is the only part guaranteed to
	// tell two contestants apart after sanitising.
	UserID   int32
	Name     string
	Username string
	Items    []*ScoreReportItem
}

type ScoreDetailReportGenerator interface {
	Generate(reports []*ContestantScoreReport) ([]byte, error)
}
