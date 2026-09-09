package exportcontest

import (
	"errors"
	"fmt"
	"strings"

	"judge-backend/internal/app/contest_domain/contest/application/bundle"
	"judge-backend/internal/app/contest_domain/contest/application/query"
	sharedservice "judge-backend/internal/app/shared/service"
	"judge-backend/internal/app/shared/types/zipfile"
)

var (
	ErrorInvalidContestID = errors.New("invalid contest id, must be greater than 0")
	ErrorContestNotFound  = errors.New("contest not found")
	ErrorInternalServer   = errors.New("internal server error")
)

type Input struct {
	ContestID int32
}

type Output struct {
	Filename string
	Content  []byte
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

type usecase struct {
	query query.ContestExportQuery
	zip   sharedservice.ZipService
}

func NewUsecase(q query.ContestExportQuery, zip sharedservice.ZipService) Usecase {
	return &usecase{query: q, zip: zip}
}

func (u *usecase) Execute(input *Input) (*Output, error) {
	if input.ContestID <= 0 {
		return nil, ErrorInvalidContestID
	}

	loaded, err := u.query.LoadContestForExport(input.ContestID)
	if errors.Is(err, query.ErrorContestNotFound) {
		return nil, ErrorContestNotFound
	}
	if err != nil {
		return nil, ErrorInternalServer
	}

	// Times go out in UTC, not the server's local zone.
	//
	// A bundle is meant to be diffed, and two installations in different zones
	// would otherwise write the same instant as different bytes — the sitting
	// would look edited when nothing had changed. UTC is also the only reading
	// that survives being carried to a venue in another zone without somebody
	// having to know where it was exported from.
	definition := &bundle.Contest{
		FormatVersion:  bundle.FormatVersion,
		Kind:           bundle.KindContest,
		Name:           loaded.Name,
		Description:    loaded.Description,
		IsPublic:       loaded.IsPublic,
		ActivateTime:   loaded.ActivateTime.UTC(),
		DeactivateTime: loaded.DeactivateTime.UTC(),
		Projects:       make([]bundle.Project, 0, len(loaded.Projects)),
	}

	for _, p := range loaded.Projects {
		definition.Projects = append(definition.Projects, bundle.Project{
			ProjectRef:            p.ProjectName,
			Name:                  p.Name,
			IsPublic:              p.IsPublic,
			JudgePriorityStrategy: p.JudgePriorityStrategy,
			ActivateTime:          p.ActivateTime.UTC(),
			DeactivateTime:        p.DeactivateTime.UTC(),
			StartTime:             p.StartTime.UTC(),
			EndTime:               p.EndTime.UTC(),
			Limits: bundle.Limits{
				MaxJudgeDurationMinutes:   p.MaxJudgeDurationMinutes,
				SubmissionCooldownSeconds: p.SubmissionCooldownSeconds,
				PidsLimit:                 p.PidsLimit,
				MemoryLimitMB:             p.MemoryLimitMB,
			},
			Scoreboard: bundle.Scoreboard{
				Compact:   p.ScoreboardCompact,
				MaskNames: p.ScoreboardMaskNames,
			},
		})
	}

	definitionJSON, err := bundle.Marshal(definition)
	if err != nil {
		return nil, ErrorInternalServer
	}

	files := []zipfile.File{{Path: bundle.ContestFile, Body: string(definitionJSON)}}
	archive, err := u.zip.ZipFromFiles(&files)
	if err != nil {
		return nil, ErrorInternalServer
	}

	return &Output{Filename: filenameFor(loaded.Name, loaded.ID), Content: archive.Bytes()}, nil
}

// filenameFor is what the browser saves. The name is only a hint — it can be
// in any script — so the id is what actually distinguishes two downloads.
func filenameFor(name string, id int32) string {
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		case r == ' ':
			return '-'
		default:
			return -1
		}
	}, name)
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "contest"
	}
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	return fmt.Sprintf("contest-%s-%d.zip", slug, id)
}
