package service

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/gradingitem"
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/testcase"
)

// ExecutionState tracks scheduling progress of a single judgment run: which
// grading items and testcases have been dispatched, and which testcases passed.
type ExecutionState struct {
	passedTestcaseIDs      map[int32]struct{}
	executedTestcaseIDs    map[int32]struct{}
	executedGradingItemIDs map[int32]struct{}
}

func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		passedTestcaseIDs:      map[int32]struct{}{},
		executedTestcaseIDs:    map[int32]struct{}{},
		executedGradingItemIDs: map[int32]struct{}{},
	}
}

func (s *ExecutionState) MarkPassed(testcaseID int32) {
	s.passedTestcaseIDs[testcaseID] = struct{}{}
}

// Batch is one wave of executable testcases resolved from the dependency graph.
type Batch struct {
	Testcases []*testcase.Testcase

	gradingItemByTestcaseID map[int32]*gradingitem.GradingItem
}

// GradingItemFor returns the grading item a testcase in this batch belongs to.
func (b *Batch) GradingItemFor(testcaseID int32) *gradingitem.GradingItem {
	return b.gradingItemByTestcaseID[testcaseID]
}

// NextBatch resolves the next wave of executable testcases for a judgment:
// grading items whose depend-testcases have all passed are opened, and within
// the opened items, testcases whose own dependencies are met are selected.
// Selected grading items and testcases are recorded in state so they are not
// dispatched twice. Returns nil when nothing further is executable.
//
// NOTE(behavior-preserving): once a grading item is opened, its testcases that
// were not executable in that same wave are never revisited in later waves.
// This mirrors the scheduling loop of the old infrastructure judge service
// (since replaced by the execute_judgment usecase);
// changing it is a behavior change and needs its own discussion.
func NextBatch(j *judgment.Judgment, state *ExecutionState) *Batch {
	gradingItems := executableGradingItems(j, state)
	if len(gradingItems) == 0 {
		return nil
	}

	testcases := executableTestcases(gradingItems, state)
	if len(testcases) == 0 {
		return nil
	}

	gradingItemByTestcaseID := map[int32]*gradingitem.GradingItem{}
	for _, gi := range gradingItems {
		for _, tc := range gi.Testcases() {
			gradingItemByTestcaseID[tc.ID()] = gi
		}
	}

	return &Batch{
		Testcases:               testcases,
		gradingItemByTestcaseID: gradingItemByTestcaseID,
	}
}

func executableGradingItems(j *judgment.Judgment, state *ExecutionState) []*gradingitem.GradingItem {
	gradingItems := []*gradingitem.GradingItem{}
	for _, gradingItem := range j.GradingItems() {
		if _, ok := state.executedGradingItemIDs[gradingItem.ID()]; ok {
			continue
		}
		if gradingItem.DependencyMet(state.passedTestcaseIDs) {
			gradingItems = append(gradingItems, gradingItem)
			state.executedGradingItemIDs[gradingItem.ID()] = struct{}{}
		}
	}
	return gradingItems
}

func executableTestcases(gradingItems []*gradingitem.GradingItem, state *ExecutionState) []*testcase.Testcase {
	testcases := []*testcase.Testcase{}
	for _, gradingItem := range gradingItems {
		for _, tc := range gradingItem.Testcases() {
			if _, ok := state.executedTestcaseIDs[tc.ID()]; ok {
				continue
			}
			if tc.DependencyMet(state.passedTestcaseIDs) {
				testcases = append(testcases, tc)
				state.executedTestcaseIDs[tc.ID()] = struct{}{}
			}
		}
	}
	return testcases
}
