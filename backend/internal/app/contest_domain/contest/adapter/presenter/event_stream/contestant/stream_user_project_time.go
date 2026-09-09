package espresenter

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"judge-backend/api"
	"judge-backend/internal/app/contest_domain/contest/adapter/presenter/types"
	streamuserprojecttime "judge-backend/internal/app/contest_domain/contest/application/usecase/contestant/stream_user_project_time"
	"judge-backend/internal/app/shared/types/httperror"
	"time"
)

type StreamUserProjectTimePresenter types.EsPresenter[streamuserprojecttime.Output, api.StreamUserProjectTimeResponseObject]
type streamUserProjectTimePresenter struct {
}

func NewStreamUserProjectTimePresenter() StreamUserProjectTimePresenter {
	return &streamUserProjectTimePresenter{}
}

func (p *streamUserProjectTimePresenter) Output(ctx context.Context, output *streamuserprojecttime.Output) api.StreamUserProjectTimeResponseObject {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		writer := bufio.NewWriter(pw)
		prev := 0
		timeType := "wait_start_submit"
		var timeLeft time.Duration
		for {
			if err := ctx.Err(); err != nil {
				return
			}

			now := time.Now()
			if timeType == "wait_start_submit" {
				timeLeft = output.ProjectTime.TimeLeftToStartSubmit(now)
				if timeLeft.Truncate(time.Second) <= 0 {
					if prev != 0 {
						if err := p.writeTimeEvent(writer, timeType, 0); err != nil {
							return
						}
						sleepDuration := time.Duration((timeLeft.Milliseconds()%1000 + 1)) * time.Millisecond
						if !p.sleepForRemainingTime(ctx, sleepDuration) {
							return
						}
					}
					timeType = "submit"
				}
			}

			if timeType == "submit" {
				now = time.Now()
				timeLeft = output.ProjectTime.TimeLeftToSubmit(now)
				if timeLeft <= 0 {
					_ = p.writeTimeEvent(writer, timeType, 0)
					break
				}
			}

			if p.shouldSleepBeforeNextIteration(&prev, timeLeft) {
				continue
			}

			if err := p.writeTimeEvent(writer, timeType, timeLeft); err != nil {
				return
			}

			if !p.sleepForRemainingTime(ctx, time.Duration(timeLeft.Milliseconds()%1000+1)*time.Millisecond) {
				return
			}
		}
	}()

	return api.StreamUserProjectTime200TexteventStreamResponse{
		Body: pr,
		Headers: api.StreamUserProjectTime200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *streamUserProjectTimePresenter) Error(err error) api.StreamUserProjectTimeResponseObject {
	return httperror.From(err).Response()
}

func (p *streamUserProjectTimePresenter) shouldSleepBeforeNextIteration(prev *int, timeLeft time.Duration) bool {
	if *prev == int(timeLeft.Seconds()) {
		return true
	}
	*prev = int(timeLeft.Seconds())
	return false
}

func (p *streamUserProjectTimePresenter) writeTimeEvent(writer *bufio.Writer, timeType string, timeLeft time.Duration) error {
	jsonText, err := json.Marshal(map[string]interface{}{
		"time_type": timeType,
		"time_left": timeLeft.Truncate(time.Second).String(),
	})
	if err != nil {
		return err
	}

	_, err = writer.WriteString("data: " + string(jsonText) + "\n\n")
	if err != nil {
		if isClientDisconnectedError(err) {
			return nil
		}
		return err
	}

	err = writer.Flush()
	if isClientDisconnectedError(err) {
		return nil
	}
	return err
}

func (p *streamUserProjectTimePresenter) sleepForRemainingTime(ctx context.Context, sleepDuration time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(sleepDuration):
		return true
	}
}
