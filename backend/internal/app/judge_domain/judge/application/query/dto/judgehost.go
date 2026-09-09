package dto

import "io"

type JudgehostInfo struct {
	ID     string
	Status string
}

type Judgehost struct {
	ID     string
	Status string
	Log    string
	Reader io.Reader
}
