package factory

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/repository"
)

type JudgmentFactory struct {
	projectRepo     repository.ProjectRepository
	submissionRepo  repository.SubmissionRepository
	gradingItemRepo repository.GradingItemRepository
	lifecycleRepo   repository.LifecycleRepository
}

func NewJudgmentFactory(
	projectRepo repository.ProjectRepository,
	submissionRepo repository.SubmissionRepository,
	gradingItemRepo repository.GradingItemRepository,
	lifecycleRepo repository.LifecycleRepository,
) *JudgmentFactory {
	return &JudgmentFactory{
		projectRepo:     projectRepo,
		submissionRepo:  submissionRepo,
		gradingItemRepo: gradingItemRepo,
		lifecycleRepo:   lifecycleRepo,
	}
}

func (f *JudgmentFactory) CreateJudgment(submissionID string, judgeType judgetype.JudgeType) (*judgment.Judgment, error) {
	submission, err := f.submissionRepo.FindByID(submissionID)
	if err != nil {
		return nil, err
	}

	project, err := f.projectRepo.FindBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	lifecycles, err := f.lifecycleRepo.FindBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	gradingItems, err := f.gradingItemRepo.FindBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	for _, gradingItem := range gradingItems {
		for _, testcase := range gradingItem.Testcases() {
			testcase.SetLifecycle(lifecycle.NewTestcaseLifecycle(lifecycles))
		}
	}

	return judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: submission.ContestProjectID(),
		Project:          project,
		SubmissionID:     submissionID,
		GradingItems:     gradingItems,
		Lifecycle:        lifecycle.NewJudgmentLifecycle(lifecycles),
		JudgeType:        judgeType,
	}), nil
}

// GradingItemMismatchReason is the reason the translation sites attach when a
// submission's grading-item set cannot be resolved for its own problem. It is
// shown to the contestant, so it names the observable condition and nothing
// about how such a submission comes to exist.
const GradingItemMismatchReason = "無法載入評分項設定"

// CreateFailedJudgment builds a judgment already terminal in ERROR for a
// submission whose grading-item set can no longer be resolved (see
// repository.ErrGradingItemsNotOwned). It deliberately skips the grading-item
// load — that is the step that failed — so the judgment carries no items and
// no score, only the reason. Without it such a submission simply never got a
// judgment: invisible to the contestant (every list and detail query requires
// one) and re-attempted by the orphan repair sweep forever.
func (f *JudgmentFactory) CreateFailedJudgment(submissionID string, judgeType judgetype.JudgeType, reason string) (*judgment.Judgment, error) {
	submission, err := f.submissionRepo.FindByID(submissionID)
	if err != nil {
		return nil, err
	}

	project, err := f.projectRepo.FindBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	lifecycles, err := f.lifecycleRepo.FindBySubmissionID(submissionID)
	if err != nil {
		return nil, err
	}

	j := judgment.New(&judgment.JudgmentConfig{
		ContestProjectID: submission.ContestProjectID(),
		Project:          project,
		SubmissionID:     submissionID,
		GradingItems:     []*gradingitem.GradingItem{},
		Lifecycle:        lifecycle.NewJudgmentLifecycle(lifecycles),
		JudgeType:        judgeType,
	})
	if err := j.Fail(); err != nil {
		return nil, err
	}
	j.SetReason(reason)
	j.SetErrorMessage(reason)
	return j, nil
}
