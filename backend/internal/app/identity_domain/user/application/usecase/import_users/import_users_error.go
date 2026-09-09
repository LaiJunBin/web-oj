package importusers

import (
	"encoding/json"
	"fmt"
)

type ImportUsersError struct {
	details []string
}

func (e *ImportUsersError) Error() string {
	return "importing users error"
}

func (e *ImportUsersError) Details() []string {
	return e.details
}

func (e *ImportUsersError) AddDetail(detail string) *ImportUsersError {
	return &ImportUsersError{
		details: append([]string{detail}, e.details...),
	}
}

func (e *ImportUsersError) WithFailedRow(row interface{}) *ImportUsersError {
	jsonRow, _ := json.Marshal(row)
	return e.AddDetail(fmt.Sprintf("failed in row: %v", string(jsonRow)))
}
