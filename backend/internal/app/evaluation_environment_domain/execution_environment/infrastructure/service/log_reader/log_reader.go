package logreader

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	logreader "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/service/log_reader"
	"strings"
	"sync"
	"time"
)

type BuildOutput struct {
	Stream string `json:"stream"`
	Aux    struct {
		ID string `json:"ID"`
	} `json:"aux"`
	ErrorDetail struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errorDetail"`
	Error          string `json:"error"`
	Status         string `json:"status"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
}

var (
	ErrorLogReaderClosed = fmt.Errorf("log reader is closed")
	// ErrorLogReaderAlreadyStarted is returned by a second call to Read: the
	// underlying io.ReadCloser is not safe for concurrent reads, so a second
	// call would race the first call's background goroutine over the same
	// stream instead of returning any new output.
	ErrorLogReaderAlreadyStarted = fmt.Errorf("log reader has already started reading")
)

type LogReader struct {
	reader  io.ReadCloser
	mu      sync.Mutex // guards closed/started; the background goroutine started by Read owns r.reader exclusively once running
	closed  bool
	started bool
}

func New(
	reader io.ReadCloser,
) logreader.LogReader {
	return &LogReader{
		reader: reader,
	}
}

func (r *LogReader) Read() (<-chan buildlog.Log, error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, ErrorLogReaderClosed
	}
	if r.started {
		r.mu.Unlock()
		return nil, ErrorLogReaderAlreadyStarted
	}
	r.started = true
	r.mu.Unlock()

	reader := bufio.NewReader(r.reader)
	logs := make(chan buildlog.Log, 100)
	go func() {
		defer close(logs)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				// io.EOF is the stream ending normally, and a read error
				// after Close() is this process tearing the stream down on
				// purpose (a cancelled build). Anything else means the
				// stream DIED midway — and since a build's verdict is read
				// out of this very stream, staying silent here would let the
				// caller see a truncated log with no errors in it and call
				// the build a success. Report it as an Error-level entry so
				// it surfaces both in the log and in that verdict.
				if err != io.EOF && !r.isClosed() {
					logs <- buildlog.New(buildloglevel.Error,
						fmt.Sprintf("build log stream ended unexpectedly: %v", err), time.Now())
				}
				break
			}

			if len(strings.TrimSpace(string(line))) == 0 {
				continue
			}

			var buildOutput BuildOutput
			if err := json.Unmarshal(line, &buildOutput); err != nil {
				logs <- buildlog.New(buildloglevel.Error, err.Error(), time.Now())
				continue
			}

			if buildOutput.Status != "" {
				if buildOutput.Progress != "" {
					logs <- buildlog.New(buildloglevel.Info, fmt.Sprintf("%s %s %s", buildOutput.Status, buildOutput.ID, buildOutput.Progress), time.Now())
				} else {
					logs <- buildlog.New(buildloglevel.Info, fmt.Sprintf("%s %s", buildOutput.Status, buildOutput.ID), time.Now())
				}
			}

			if buildOutput.Stream != "" {
				logs <- buildlog.New(buildloglevel.Info, buildOutput.Stream, time.Now())
			}

			if buildOutput.Aux.ID != "" {
				logs <- buildlog.New(buildloglevel.Success, fmt.Sprintf("writing image %s", buildOutput.Aux.ID), time.Now())
			}

			if buildOutput.Error != "" {
				logs <- buildlog.New(buildloglevel.Error, buildOutput.Error, time.Now())
			}
		}
	}()

	return logs, nil
}

// isClosed reports whether Close has been called. Close sets the flag before
// closing the underlying reader, so a read error caused by that Close is
// always seen as deliberate.
func (r *LogReader) isClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

func (r *LogReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrorLogReaderClosed
	}

	r.closed = true
	return r.reader.Close()
}
