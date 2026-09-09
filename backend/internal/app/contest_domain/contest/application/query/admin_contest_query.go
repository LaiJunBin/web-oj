package query

import "judge-backend/internal/app/contest_domain/contest/application/query/dto"

type AdminContestQuery interface {
	ListContest() ([]*dto.AdminContest, error)
	GetContestByID(id int32) (*dto.AdminContestDetail, error)
}

// ContestExportQuery loads a contest whole, for writing to a bundle. Separate
// from the admin query because it answers a different question: not "what does
// the contest screen need" but "what would be lost if this sitting were
// recreated elsewhere from this data alone".
type ContestExportQuery interface {
	LoadContestForExport(id int32) (*dto.ContestExport, error)
}
