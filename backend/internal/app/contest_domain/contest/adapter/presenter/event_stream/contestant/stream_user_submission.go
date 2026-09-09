package espresenter

import (
	"context"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	streamusersubmission "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_submission"
	"judge-backend/internal/app/contest_domain/events/submissionupdated"
	"judge-backend/internal/app/shared/adapter/presenter/sse"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

type StreamUserSubmissionPresenter types.EsPresenter[streamusersubmission.Output, api.StreamUserSubmissionResponseObject]
type streamUserSubmissionPresenter struct {
	eventBus sharedtypes.EventBus
}

func NewStreamUserSubmissionPresenter(
	eventBus sharedtypes.EventBus,
) StreamUserSubmissionPresenter {
	return &streamUserSubmissionPresenter{
		eventBus: eventBus,
	}
}

type GradingItem struct {
	ID        int32       `json:"id"`
	Score     float64     `json:"score"`
	Status    string      `json:"status"`
	Testcases []*Testcase `json:"testcases"`
}

type Testcase struct {
	ID       int32  `json:"id"`
	ExecTime int64  `json:"execTime"`
	ExitCode int32  `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Status   string `json:"status"`
}

type Submission struct {
	Status string  `json:"status"`
	Score  float64 `json:"score"`
	// Reason rides along so an ERROR pushed live carries its explanation;
	// without it the contestant had to reload to read the REST detail.
	Reason       *string        `json:"reason,omitempty"`
	// StaffNote is the operator's note to the contestant. Unlike Reason it
	// can be retracted: a frame WITHOUT the field means the note is gone, so
	// the client must clear what it shows rather than keep the last value.
	StaffNote    *string        `json:"staffNote,omitempty"`
	GradingItems []*GradingItem `json:"gradingItems"`
}

func (p *streamUserSubmissionPresenter) Output(ctx context.Context, output *streamusersubmission.Output) api.StreamUserSubmissionResponseObject {
	// The comment is load-bearing: with no init write, the response headers
	// stay buffered until the first judge update, so a client (and the BDD
	// harness) waiting for the stream to open would hang on a submission
	// nothing is currently judging.
	writeInit := func(w *sse.Writer) error { return w.WriteComment("stream opened") }
	body := sse.Stream(ctx, p.eventBus, writeInit, func(w *sse.Writer, e *submissionupdated.Event) error {
		j := e.Judgment()
		if j.SubmissionID != output.SubmissionID || j.IsRejudge {
			return nil
		}

		// Resolved once per delivered frame, AFTER the filters above: only
		// this submission's own events pay for it, and the resolver's TTL
		// memo caps even that at one query per stream per few seconds.
		// Re-resolving at all is the point — testcase visibility can change
		// while the stream is open.
		sampleTestcaseIDs := output.ResolveSampleTestcaseIDs()

		gradingItems := make([]*GradingItem, 0, len(j.GradingItems))
		for _, gi := range j.GradingItems {
			testcases := make([]*Testcase, 0, len(gi.Testcases))
			for _, tc := range gi.Testcases {
				// Same visibility rule as the REST detail: a testcase that is
				// not one of the problem's samples is absent from the frame
				// entirely. The event payload carries every testcase because
				// the admin stream consumes the same events; the contestant
				// side is where the filtering belongs.
				if !sampleTestcaseIDs[tc.ID] {
					continue
				}
				testcases = append(testcases, &Testcase{
					ID:       tc.ID,
					Status:   tc.Status,
					ExecTime: tc.ExecTime,
					ExitCode: int32(tc.ExitCode),
					Stdout:   tc.Stdout,
					Stderr:   tc.Stderr,
				})
			}
			gradingItems = append(gradingItems, &GradingItem{
				ID:        gi.ID,
				Score:     gi.Score,
				Status:    gi.Status,
				Testcases: testcases,
			})
		}

		return w.WriteData(&Submission{
			Status:       j.Status,
			Score:        j.Score,
			Reason:       j.Reason,
			StaffNote:    j.StaffNote,
			GradingItems: gradingItems,
		})
	})

	return api.StreamUserSubmission200TexteventStreamResponse{
		Body: body,
		Headers: api.StreamUserSubmission200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamUserSubmissionPresenter) Error(err error) api.StreamUserSubmissionResponseObject {
	return httperror.From(err).Response()
}
