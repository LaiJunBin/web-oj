package query

import (
	"errors"

	"judge-backend/internal/app/project_domain/project/application/query/dto"
)

type ProjectQuery interface {
	ListProject() ([]dto.Project, error)
	GetProjectByID(id int32) (*dto.ProjectDetail, error)
}

// ErrorProjectNotFound is returned instead of a nil result, following
// ErrorAssetNotFound in this package. The (nil, nil) spelling the older
// queries use puts the burden on every caller to remember that a nil result
// is not a bug.
var ErrorProjectNotFound = errors.New("project not found")

// ProjectExportQuery loads a project whole, for writing to a bundle.
//
// Separate from ProjectQuery because it answers a different question: not
// "what does the project screen need" but "what would be lost if this project
// were recreated elsewhere from this data alone".
type ProjectExportQuery interface {
	LoadProjectForExport(id int32) (*dto.ProjectExport, error)
}
