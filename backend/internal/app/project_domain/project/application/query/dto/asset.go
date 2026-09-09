package dto

import "time"

type Asset struct {
	ID           string
	Name         string
	Size         int64
	LastModified time.Time
	Enabled      bool
}
