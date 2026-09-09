package dto

type Submission struct {
	ID           string
	Status       string
	GradingItems []GradingItem
	CreatedAt    int64
}
