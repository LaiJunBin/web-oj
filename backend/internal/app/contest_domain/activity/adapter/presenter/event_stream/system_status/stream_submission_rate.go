package espresenter

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"time"

	"judge-backend/api"
	jsonpresenter "judge-backend/internal/app/contest_domain/activity/adapter/presenter/json/system_status"
	"judge-backend/internal/app/contest_domain/activity/adapter/presenter/types"
	getsubmissionrate "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/get_submission_rate"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

// submissionRateInterval is how often a connected client is re-sent the series.
//
// Deliberately the slowest of the status streams. The buckets have minute
// resolution, so a faster tick re-sends numbers that cannot have changed —
// pushing every two seconds would repeat each bucket thirty times before it
// could move — while paying for a scan of the submissions collection each time.
const submissionRateInterval = 10 * time.Second

// StreamSubmissionRateInput carries what the stream needs to keep going, plus
// the snapshot the controller already fetched. See StreamJudgmentErrorsInput
// for why the first result is handed over rather than re-queried.
type StreamSubmissionRateInput struct {
	Input *getsubmissionrate.Input
	First *getsubmissionrate.Output
}

type StreamSubmissionRatePresenter types.EsPresenter[StreamSubmissionRateInput, api.StreamSubmissionRateResponseObject]

type streamSubmissionRatePresenter struct {
	usecase getsubmissionrate.Usecase
	logger  sharedtypes.Logger
}

func NewStreamSubmissionRatePresenter(
	usecase getsubmissionrate.Usecase,
	logger sharedtypes.Logger,
) StreamSubmissionRatePresenter {
	return &streamSubmissionRatePresenter{
		usecase: usecase,
		logger:  logger,
	}
}

func (p *streamSubmissionRatePresenter) Output(ctx context.Context, input *StreamSubmissionRateInput) api.StreamSubmissionRateResponseObject {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		writer := bufio.NewWriter(pw)

		if !p.write(writer, input.First) {
			return
		}

		ticker := time.NewTicker(submissionRateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			// Re-run rather than shifting the previous series along: the window
			// is anchored to the current minute, so each tick drops the oldest
			// bucket and adds a new one, and the newest bucket is still filling
			// while it is being reported.
			output, err := p.usecase.Execute(input.Input)
			if err != nil {
				p.logger.Errorf("[StreamSubmissionRate] Failed to read the submission rate: %v", err)
				return
			}
			if !p.write(writer, output) {
				return
			}
		}
	}()

	return api.StreamSubmissionRate200TexteventStreamResponse{
		Body: pr,
		Headers: api.StreamSubmissionRate200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

// write emits one frame, reporting whether the stream should continue.
func (p *streamSubmissionRatePresenter) write(writer *bufio.Writer, output *getsubmissionrate.Output) bool {
	payload := api.GetSubmissionRate200JSONResponse{
		Success: true,
		Data:    jsonpresenter.SubmissionRateSeries(output.Points),
	}
	jsonText, err := json.Marshal(payload)
	if err != nil {
		p.logger.Errorf("[StreamSubmissionRate] Failed to encode the submission rate: %v", err)
		return false
	}

	if _, err := writer.WriteString("data: " + string(jsonText) + "\n\n"); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamSubmissionRate] Failed to write a snapshot: %v", err)
		}
		return false
	}
	if err := writer.Flush(); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamSubmissionRate] Failed to flush a snapshot: %v", err)
		}
		return false
	}
	return true
}

func (p *streamSubmissionRatePresenter) Error(err error) api.StreamSubmissionRateResponseObject {
	return httperror.From(err).Response()
}
