package espresenter

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"time"

	"judge-backend/api"
	jsonpresenter "judge-backend/internal/app/judge_domain/worker/adapter/presenter/json"
	"judge-backend/internal/app/judge_domain/worker/adapter/presenter/types"
	getjudgequeue "judge-backend/internal/app/judge_domain/worker/application/usecase/get_judge_queue"
	sharedtypes "judge-backend/internal/app/shared/types"
	"judge-backend/internal/app/shared/types/httperror"
)

// snapshotInterval is how often a connected client is sent a fresh reading.
//
// The frontend polls the equivalent endpoint every 5–10 seconds, so this is
// deliberately a little faster than that and no faster: each frame is a lock-free
// read of a few in-memory counters, but it is also a write to every open
// connection, and a status page is not improved by sub-second resolution.
const snapshotInterval = 2 * time.Second

type StreamJudgeQueuePresenter types.EsPresenter[struct{}, api.StreamJudgeQueueResponseObject]

type streamJudgeQueuePresenter struct {
	usecase getjudgequeue.Usecase
	logger  sharedtypes.Logger
}

func NewStreamJudgeQueuePresenter(
	usecase getjudgequeue.Usecase,
	logger sharedtypes.Logger,
) StreamJudgeQueuePresenter {
	return &streamJudgeQueuePresenter{
		usecase: usecase,
		logger:  logger,
	}
}

// Output streams a fresh snapshot on a fixed interval.
//
// Ticker-driven, unlike the worker stream next to it, which re-renders when a
// worker publishes a status change. There is no equivalent event here — and
// more importantly there should not be one: oldestQueuedSeconds grows with the
// wall clock, so a stream that only emitted on push/pop would sit frozen at the
// last-observed wait for exactly as long as the pipeline was stuck, which is the
// one situation this page exists for. Pushing on a timer also bounds the write
// rate, which a queue event would not: a burst of a thousand submissions would
// otherwise be a thousand frames.
func (p *streamJudgeQueuePresenter) Output(ctx context.Context, output *struct{}) api.StreamJudgeQueueResponseObject {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		writer := bufio.NewWriter(pw)

		ticker := time.NewTicker(snapshotInterval)
		defer ticker.Stop()

		// The first frame goes out immediately: a client that had to wait a
		// full interval before seeing anything would render an empty panel on
		// every page load.
		for {
			if !p.writeSnapshot(writer) {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	return api.StreamJudgeQueue200TexteventStreamResponse{
		Body: pr,
		Headers: api.StreamJudgeQueue200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

// writeSnapshot emits one frame, reporting whether the stream should continue.
func (p *streamJudgeQueuePresenter) writeSnapshot(writer *bufio.Writer) bool {
	output, err := p.usecase.Execute(&getjudgequeue.Input{})
	if err != nil {
		p.logger.Errorf("[StreamJudgeQueue] Failed to read the queue status: %v", err)
		return false
	}

	// Deliberately the polled endpoint's own response type: a client reading
	// this stream and a client polling GET /judge-queue must be able to use one
	// parser, and reusing the type is what keeps that true when the payload
	// changes.
	payload := api.GetJudgeQueue200JSONResponse{
		Success: true,
		Data:    jsonpresenter.JudgeQueueStatus(output.Status),
	}
	jsonText, err := json.Marshal(payload)
	if err != nil {
		p.logger.Errorf("[StreamJudgeQueue] Failed to encode the queue status: %v", err)
		return false
	}

	if _, err := writer.WriteString("data: " + string(jsonText) + "\n\n"); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamJudgeQueue] Failed to write a snapshot: %v", err)
		}
		return false
	}
	if err := writer.Flush(); err != nil {
		if !isClientDisconnectedError(err) {
			p.logger.Errorf("[StreamJudgeQueue] Failed to flush a snapshot: %v", err)
		}
		return false
	}
	return true
}

func (p *streamJudgeQueuePresenter) Error(err error) api.StreamJudgeQueueResponseObject {
	return httperror.From(err).Response()
}
