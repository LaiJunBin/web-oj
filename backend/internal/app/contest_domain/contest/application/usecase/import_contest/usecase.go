package importcontest

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"

	"judge-backend/internal/app/contest_domain/contest/application/bundle"
	createcontest "judge-backend/internal/app/contest_domain/contest/application/usecase/create_contest"
	sharedservice "judge-backend/internal/app/shared/service"
)

var (
	ErrorInternalServer = errors.New("internal server error")
	ErrorInvalidBundle  = errors.New("這不是有效的競賽匯出檔")
)

type InvalidBundleError struct{ Details []string }

func (e *InvalidBundleError) Error() string { return ErrorInvalidBundle.Error() }
func (e *InvalidBundleError) Unwrap() error { return ErrorInvalidBundle }

type Input struct {
	Bundle *bytes.Buffer
}

type Output struct {
	ContestID int32
	Name      string
	// CreatedProjects counts the slots that resolved; UnresolvedProjects names
	// the ones that did not, so the operator knows exactly which problem
	// bundles to import before adding them by hand.
	CreatedProjects    int32
	UnresolvedProjects []string
}

type Usecase interface {
	Execute(input *Input) (*Output, error)
}

// ProjectLibrary resolves a problem name to its id.
type ProjectLibrary interface {
	// FindIDByName returns 0 when the name matches nothing, and also when it
	// matches more than one: names are not unique, and a slot pointed at the
	// wrong problem produces a contest that looks configured and grades
	// something else. Ambiguity is reported, never guessed.
	FindIDByName(name string) (int32, error)
}

// ContestNamer resolves a collision with an existing contest's name.
type ContestNamer interface {
	UniqueName(wanted string) (string, error)
}

type usecase struct {
	zip           sharedservice.ZipService
	library       ProjectLibrary
	namer         ContestNamer
	createContest createcontest.Usecase
}

func NewUsecase(
	zip sharedservice.ZipService,
	library ProjectLibrary,
	namer ContestNamer,
	createContest createcontest.Usecase,
) Usecase {
	return &usecase{zip: zip, library: library, namer: namer, createContest: createContest}
}

// Execute verifies the whole bundle before creating anything, so a rejected
// bundle leaves nothing behind.
//
// Unresolved slots are NOT a rejection. A venue importing a sitting before all
// its problems have arrived is an ordinary order to work in, and refusing
// would make the operator sequence their imports by trial and error. The
// contest is created with the slots that resolved, and the rest are named in
// the response.
func (u *usecase) Execute(input *Input) (*Output, error) {
	definition, err := u.verify(input)
	if err != nil {
		return nil, err
	}

	name, err := u.namer.UniqueName(definition.Name)
	if err != nil {
		return nil, ErrorInternalServer
	}

	configs := make([]createcontest.CreateProjectConfig, 0, len(definition.Projects))
	unresolved := make([]string, 0)

	for _, p := range definition.Projects {
		projectID, err := u.library.FindIDByName(p.ProjectRef)
		if err != nil {
			return nil, ErrorInternalServer
		}
		if projectID == 0 {
			unresolved = append(unresolved, p.ProjectRef)
			continue
		}

		configs = append(configs, createcontest.CreateProjectConfig{
			ProjectID:                 projectID,
			Name:                      p.Name,
			ActivationTime:            p.ActivateTime,
			DeactivationTime:          p.DeactivateTime,
			StartTime:                 p.StartTime,
			EndTime:                   p.EndTime,
			IsPublic:                  p.IsPublic,
			JudgePriorityStrategy:     p.JudgePriorityStrategy,
			MaxJudgeDurationMinutes:   p.Limits.MaxJudgeDurationMinutes,
			SubmissionCooldownSeconds: p.Limits.SubmissionCooldownSeconds,
			PidsLimit:                 p.Limits.PidsLimit,
			MemoryLimitMB:             p.Limits.MemoryLimitMB,
			ScoreboardCompact:         p.Scoreboard.Compact,
			ScoreboardMaskNames:       p.Scoreboard.MaskNames,
		})
	}

	out, err := u.createContest.Execute(&createcontest.Input{
		Name:           name,
		Description:    definition.Description,
		ActivateTime:   definition.ActivateTime,
		DeactivateTime: definition.DeactivateTime,
		IsPublic:       definition.IsPublic,
		ProjectConfigs: configs,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrorInternalServer, err.Error())
	}

	return &Output{
		ContestID:          out.ID,
		Name:               name,
		CreatedProjects:    int32(len(configs)),
		UnresolvedProjects: unresolved,
	}, nil
}

func (u *usecase) verify(input *Input) (*bundle.Contest, error) {
	if input.Bundle == nil || input.Bundle.Len() == 0 {
		return nil, &InvalidBundleError{Details: []string{"沒有收到檔案"}}
	}
	if err := u.zip.CheckZip(input.Bundle); err != nil {
		return nil, &InvalidBundleError{Details: []string{"上傳的檔案不是有效的 zip"}}
	}

	raw := input.Bundle.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return nil, &InvalidBundleError{Details: []string{"上傳的檔案不是有效的 zip"}}
	}

	var definitionJSON []byte
	for _, f := range reader.File {
		if f.Name != bundle.ContestFile {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, ErrorInternalServer
		}
		definitionJSON, err = io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, ErrorInternalServer
		}
		break
	}
	if definitionJSON == nil {
		return nil, &InvalidBundleError{Details: []string{"匯出檔裡沒有 " + bundle.ContestFile}}
	}

	definition, err := bundle.Unmarshal(definitionJSON)
	if err != nil {
		return nil, &InvalidBundleError{Details: []string{bundle.ContestFile + " 無法解析：" + err.Error()}}
	}

	if problems := definition.Validate(); len(problems) > 0 {
		details := make([]string, 0, len(problems))
		for _, p := range problems {
			details = append(details, p.Error())
		}
		return nil, &InvalidBundleError{Details: details}
	}

	return definition, nil
}
