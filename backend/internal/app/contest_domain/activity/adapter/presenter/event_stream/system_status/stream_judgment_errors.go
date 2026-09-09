// Package espresenter streams the whole-instance status views on a timer.
//
// Ticker-driven rather than event-driven, like the judge-queue stream this is
// modelled on: both sources here are queries over stored documents with no
// event to subscribe to. A timer also bounds the write rate, which an
// event-per-change stream would not — a burst of a thousand submissions would
// otherwise be a thousand frames.
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
	listjudgmenterrors "judge-backend/internal/app/contest_domain/activity/application/usecase/system_status/list_judgment_errors"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

// judgmentErrorsInterval is how often a connected client is re-sent the list.
//
// Faster than the submission-rate stream beside it and slower than the
// judge-queue one: a new failure is something an admin wants to see promptly,
// but the payload only changes when something actually breaks, and the query
// behind it is a scan of the judgments collection rather than a read of
// in-memory counters.
const judgmentErrorsInterval = 5 * time.Second

// StreamJudgmentErrorsInput carries what the stream needs to keep going, plus
// the snapshot the controller already fetched.
//
// The controller runs the usecase once before the stream opens so a rejected
// parameter answers 400 instead of opening a stream that closes again without
// explanation. Handing that first result over rather than discarding it is what
// lets the connection emit immediately without repeating the query.
type StreamJudgmentErrorsInput struct {
	Input *listjudgmenterrors.Input
	First *listjudgmenterrors.Output
}

type StreamJudgmentErrorsPresenter types.EsPresenter[StreamJudgmentErrorsInput, api.StreamJudgmentErrorsResponseObject]

type streamJudgmentErrorsPresenter struct {
	usecase listjudgmenterrors.Usecase
	logger  sharedtypes.Logger
}

func NewStreamJudgmentErrorsPresenter(
	usecase listjudgmenterrors.Usecase,
	logger sharedtypes.Logger,
) StreamJudgmentErrorsPresenter {
	return &streamJudgmentErrorsPresenter{
		usecase: usecase,
		logger:  logger,
	}
}

func (p *streamJudgmentErrorsPresenter) Output(ctx context.Context, input *StreamJudgmentErrorsInput) api.StreamJudgmentErrorsResponseObject {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		writer := bufio.NewWriter(pw)

		// The connect-time snapshot goes out before the first tick: a client
		// that had to wait a whole interval would render an empty panel on
		// every page load.
		if !p.write(writer, input.First) {
			return
		}

		ticker := time.NewTicker(judgmentErrorsInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			output, err := p.usecase.Execute(input.Input)
			if err != nil {
				p.logger.Errorf("[StreamJudgmentErrors] Failed to read the judgment errors: %v", err)
				return
			}
			if !p.write(writer, output) {
				return
			}
		}
	}()

	return api.StreamJudgmentErrors200TexteventStreamResponse{
		Body: pr,
		Headers: api.StreamJudgmentErrors200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

// write emits one frame, reporting whether the stream should continue.
func (p *streamJudgmentErrorsPresenter) write(writer *bufio.Writer, output *listjudgmenterrors.Output) bool {
	// Deliberately the polled endpoint's own response type: a client reading
	// this stream and a client polling GET /judgment-errors must be able to use
	// one parser, and reusing the type is what keeps that true when the payload
	// changes.
	payload := api.ListJudgmentErrors200JSONResponse{
		Success: true,
		Data:    jsonpresenter.JudgmentErrorList(output.Errors),
	}
	jsonText, err := json.Marshal(payload)
	if err != nil {
		p.logger.Errorf("[StreamJudgmentErrors] Failed to encode the judgment errors: %v", err)
		return false
	}

	if _, err := writer.WriteString("data: " + string(jsonText) + "\n\n"); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamJudgmentErrors] Failed to write a snapshot: %v", err)
		}
		return false
	}
	if err := writer.Flush(); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamJudgmentErrors] Failed to flush a snapshot: %v", err)
		}
		return false
	}
	return true
}

func (p *streamJudgmentErrorsPresenter) Error(err error) api.StreamJudgmentErrorsResponseObject {
	return httperror.From(err).Response()
}
