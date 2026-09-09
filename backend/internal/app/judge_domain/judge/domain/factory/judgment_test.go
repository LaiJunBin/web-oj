package factory_test

import (
	"errors"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem/scoringstrategy"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/judgetype"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/lifecycle"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/submission"
	"judge-backend/internal/app/judge_domain/judge/domain/factory"
	"judge-backend/internal/app/judge_domain/judge/mocks"
	"testing"
)

func newTestSubmission(t *testing.T, submissionID string, contestProjectID int32) *submission.Submission {
	t.Helper()
	sub, err := submission.New(&submission.SubmissionConfig{
		UserID:           1,
		ContestProjectID: contestProjectID,
	})
	if err != nil {
		t.Fatalf("failed to create test submission: %v", err)
	}
	sub.SetID(submissionID)
	return sub
}

func TestJudgmentFactory(t *testing.T) {
	t.Run("CreateJudgment", func(t *testing.T) {
		t.Run("should create judgment with assembled data from repositories", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			submissionID := "sub-1"
			contestProjectID := int32(42)

			expectedSubmission := newTestSubmission(t, submissionID, contestProjectID)

			expectedProject := project.NewProject(&project.ProjectConfig{
				ID:           1,
				ImageTag:     "gcc:12",
				TestImageTag: "test:1.0",
				MountPath:    "/workspace",
			})

			expectedLifecycles := []*lifecycle.Lifecycle{
				lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
					ID: 1, ExecutionEnvironment: "project", Type: "before_all", Command: "make setup",
				}),
				lifecycle.NewLifecycle(&lifecycle.LifecycleConfig{
					ID: 2, ExecutionEnvironment: "test", Type: "before_each", Command: "init.sh",
				}),
			}

			tc1 := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID: 1, Command: "test1", Score: 50,
			})
			tc2 := testcase.NewTestcase(&testcase.TestcaseConfig{
				ID: 2, Command: "test2", Score: 50,
			})
			expectedGradingItems := []*gradingitem.GradingItem{
				gradingitem.NewGradingItem(&gradingitem.GradingItemConfig{
					ID:              1,
					Score:           100,
					ScoringStrategy: scoringstrategy.IncreaseScoreStrategy,
					Testcases:       []*testcase.Testcase{tc1, tc2},
				}),
			}

			submissionRepo.On("FindByID", submissionID).Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", submissionID).Return(expectedProject, nil)
			lifecycleRepo.On("FindBySubmissionID", submissionID).Return(expectedLifecycles, nil)
			gradingItemRepo.On("FindBySubmissionID", submissionID).Return(expectedGradingItems, nil)

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment(submissionID, judgetype.JUDGE)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if judgment == nil {
				t.Fatal("expected judgment to be non-nil")
			}
			if judgment.SubmissionID() != submissionID {
				t.Errorf("expected submissionID to be %s, got %s", submissionID, judgment.SubmissionID())
			}
			if judgment.ContestProjectID() != contestProjectID {
				t.Errorf("expected contestProjectID to be %d, got %d", contestProjectID, judgment.ContestProjectID())
			}
			if judgment.Project().ID() != 1 {
				t.Errorf("expected project ID to be 1, got %d", judgment.Project().ID())
			}
			if !judgment.JudgeType().IsJudge() {
				t.Errorf("expected judgeType to be JUDGE, got %s", judgment.JudgeType().String())
			}
			if len(judgment.GradingItems()) != 1 {
				t.Errorf("expected 1 grading item, got %d", len(judgment.GradingItems()))
			}
			if judgment.Lifecycle() == nil {
				t.Fatal("expected lifecycle to be non-nil")
			}
			if len(judgment.Lifecycle().BeforeAll()) != 1 {
				t.Errorf("expected 1 BeforeAll lifecycle, got %d", len(judgment.Lifecycle().BeforeAll()))
			}
			for _, gi := range judgment.GradingItems() {
				for _, tc := range gi.Testcases() {
					if tc.Lifecycle() == nil {
						t.Errorf("expected testcase %d lifecycle to be set", tc.ID())
					}
				}
			}
		})

		t.Run("should return error when submission repository fails", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			submissionRepo.On("FindByID", "sub-1").Return(nil, errors.New("submission not found"))

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.JUDGE)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if judgment != nil {
				t.Errorf("expected judgment to be nil on error")
			}
		})

		t.Run("should return error when project repository fails", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			expectedSubmission := newTestSubmission(t, "sub-1", 10)
			submissionRepo.On("FindByID", "sub-1").Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", "sub-1").Return(nil, errors.New("project not found"))

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.JUDGE)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if judgment != nil {
				t.Errorf("expected judgment to be nil on error")
			}
		})

		t.Run("should return error when lifecycle repository fails", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			expectedSubmission := newTestSubmission(t, "sub-1", 10)
			expectedProject := project.NewProject(&project.ProjectConfig{
				ID: 1, ImageTag: "gcc:12", TestImageTag: "test:1.0", MountPath: "/workspace",
			})

			submissionRepo.On("FindByID", "sub-1").Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", "sub-1").Return(expectedProject, nil)
			lifecycleRepo.On("FindBySubmissionID", "sub-1").Return(nil, errors.New("lifecycle query failed"))

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.JUDGE)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if judgment != nil {
				t.Errorf("expected judgment to be nil on error")
			}
		})

		t.Run("should return error when grading item repository fails", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			expectedSubmission := newTestSubmission(t, "sub-1", 10)
			expectedProject := project.NewProject(&project.ProjectConfig{
				ID: 1, ImageTag: "gcc:12", TestImageTag: "test:1.0", MountPath: "/workspace",
			})

			submissionRepo.On("FindByID", "sub-1").Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", "sub-1").Return(expectedProject, nil)
			lifecycleRepo.On("FindBySubmissionID", "sub-1").Return([]*lifecycle.Lifecycle{}, nil)
			gradingItemRepo.On("FindBySubmissionID", "sub-1").Return(nil, errors.New("grading items not found"))

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.JUDGE)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if judgment != nil {
				t.Errorf("expected judgment to be nil on error")
			}
		})

		t.Run("should handle empty grading items and lifecycles", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			expectedSubmission := newTestSubmission(t, "sub-1", 10)
			expectedProject := project.NewProject(&project.ProjectConfig{
				ID: 1, ImageTag: "gcc:12", TestImageTag: "test:1.0", MountPath: "/workspace",
			})

			submissionRepo.On("FindByID", "sub-1").Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", "sub-1").Return(expectedProject, nil)
			lifecycleRepo.On("FindBySubmissionID", "sub-1").Return([]*lifecycle.Lifecycle{}, nil)
			gradingItemRepo.On("FindBySubmissionID", "sub-1").Return([]*gradingitem.GradingItem{}, nil)

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.JUDGE)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if judgment == nil {
				t.Fatal("expected judgment to be non-nil")
			}
			if len(judgment.GradingItems()) != 0 {
				t.Errorf("expected 0 grading items, got %d", len(judgment.GradingItems()))
			}
		})

		t.Run("should create judgment with REJUDGE type", func(t *testing.T) {
			projectRepo := mocks.NewProjectRepository(t)
			submissionRepo := mocks.NewSubmissionRepository(t)
			gradingItemRepo := mocks.NewGradingItemRepository(t)
			lifecycleRepo := mocks.NewLifecycleRepository(t)

			expectedSubmission := newTestSubmission(t, "sub-1", 10)
			expectedProject := project.NewProject(&project.ProjectConfig{
				ID: 1, ImageTag: "gcc:12", TestImageTag: "test:1.0", MountPath: "/workspace",
			})

			submissionRepo.On("FindByID", "sub-1").Return(expectedSubmission, nil)
			projectRepo.On("FindBySubmissionID", "sub-1").Return(expectedProject, nil)
			lifecycleRepo.On("FindBySubmissionID", "sub-1").Return([]*lifecycle.Lifecycle{}, nil)
			gradingItemRepo.On("FindBySubmissionID", "sub-1").Return([]*gradingitem.GradingItem{}, nil)

			f := factory.NewJudgmentFactory(projectRepo, submissionRepo, gradingItemRepo, lifecycleRepo)
			judgment, err := f.CreateJudgment("sub-1", judgetype.REJUDGE)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !judgment.JudgeType().IsRejudge() {
				t.Errorf("expected judgeType to be REJUDGE, got %s", judgment.JudgeType().String())
			}
		})
	})
}
