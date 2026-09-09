package imageinfo

import "time"

type ImageInfo struct {
	Tag         string
	Description string
	IsReady     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
