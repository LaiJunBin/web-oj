package espresenter

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"judge-backend/api"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/adapter/presenter/types"
	streambuildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/stream_build_log"
	buildstatus "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/status"
	"time"
)

type StreamBuildLogPresenter struct{}

func NewStreamBuildLogPresenter() types.StreamBuildLogPresenterType {
	return &StreamBuildLogPresenter{}
}

func (p *StreamBuildLogPresenter) Output(ctx context.Context, output *streambuildlog.Output) api.StreamLogsResponseObject {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		writer := bufio.NewWriter(pw)
		i := 0
		for {
			if err := ctx.Err(); err != nil {
				return
			}
			logs := output.Logs()
			for i < len(logs) {
				jsonText, err := json.Marshal(map[string]string{
					"level":   logs[i].Level().String(),
					"message": logs[i].Message(),
					"time":    logs[i].Time().Format(time.RFC3339),
				})
				if err != nil {
					return
				}
				if _, err = writer.WriteString("data: " + string(jsonText) + "\n\n"); err != nil {
					if isClientDisconnectedError(err) {
						return
					}
					return
				}
				if err := writer.Flush(); err != nil {
					if isClientDisconnectedError(err) {
						return
					}
					return
				}
				i++
			}
			if !output.Status().Equals(buildstatus.PENDING) && !output.Status().Equals(buildstatus.BUILDING) {
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()

	return api.StreamLogs200TexteventStreamResponse{
		Body: pr,
		Headers: api.StreamLogs200ResponseHeaders{
			CacheControl:     "no-cache",
			Connection:       "keep-alive",
			TransferEncoding: "chunked",
		},
	}
}

func (p *StreamBuildLogPresenter) Error(err error) api.StreamLogsResponseObject {
	switch {
	case errors.Is(err, streambuildlog.ErrorTagIsRequired),
		errors.Is(err, streambuildlog.ErrorBuildNotInProgress),
		errors.Is(err, streambuildlog.ErrorBuildNotStarted):
		return &api.StreamLogs400JSONResponse{
			Success: false,
			Message: api.SharedAPIBadRequestResponseMessage(err.Error()),
		}
	case errors.Is(err, streambuildlog.ErrorImageNotFound),
		errors.Is(err, streambuildlog.ErrorBuildNotFound):
		return &api.StreamLogs404JSONResponse{
			Success: false,
			Message: api.SharedAPINotFoundResponseMessage(err.Error()),
		}
	case errors.Is(err, streambuildlog.ErrorInternalServer):
		fallthrough
	default:
		return &api.StreamLogs500JSONResponse{
			Success: false,
			Message: api.SharedAPIInternalServerErrorResponseMessage(err.Error()),
		}
	}
}
