package dto

type GradingItem struct {
	ID        int32
	Name      string
	Score     float64
	Status    string
	Testcases []TestCase
}
