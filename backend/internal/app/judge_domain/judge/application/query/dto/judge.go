package dto

type Judge struct {
	ID         string
	Type       string
	Status     string
	Submission Submission
	CreatedAt  int64
}
