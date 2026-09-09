package logreader_test

import (
	"bytes"
	"errors"
	"io"
	buildloglevel "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/buildloglevel"
	buildlog "judge-backend/internal/app/evaluation_environment_domain/execution_environment/domain/entity/build/valueobject/log"
	logreader "judge-backend/internal/app/evaluation_environment_domain/execution_environment/infrastructure/service/log_reader"
	"testing"
	"time"
)

func assertCurrentLogCountAtLeast(t *testing.T, logs <-chan buildlog.Log, expected int) {
	// 2s, not a tighter bound: this is a safety net against a genuine hang,
	// not a precision timing assertion (the loop below returns as soon as
	// expected logs arrive). A tighter budget was observed to flake under
	// -race plus full-suite CPU contention, where scheduling latency alone
	// can exceed a few milliseconds.
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-timer.C:
			t.Error("Test timed out while waiting for logs")
			return
		case _, ok := <-logs:
			if !ok {
				return
			}
			count++
			if count >= expected {
				return
			}
		}
	}
}

func TestLogReader_SimpleRead(t *testing.T) {
	buildOutput := `
	{"errorDetail":{"code":127,"message":"The command '/bin/sh -c apt-getxx update \u0026\u0026 apt-get install -y     curl     \u0026\u0026 rm -rf /var/lib/apt/lists/*' returned a non-zero code: 127"},"error":"The command '/bin/sh -c apt-getxx update \u0026\u0026 apt-get install -y     curl     \u0026\u0026 rm -rf /var/lib/apt/lists/*' returned a non-zero code: 127"}
	`
	logReader := logreader.New(io.NopCloser(bytes.NewReader([]byte(buildOutput))))
	logs, _ := logReader.Read()
	assertCurrentLogCountAtLeast(t, logs, 1)
}

// TestLogReader_StreamingRead proves a single Read() call keeps yielding
// newly-written log lines as they arrive, rather than snapshotting whatever
// was available at call time. It uses io.Pipe so a Write only completes once
// LogReader's background goroutine has actually consumed it — matching how
// the real Docker build response body streams — instead of a bytes.Buffer,
// which returns io.EOF immediately once drained and can't model this.
func TestLogReader_StreamingRead(t *testing.T) {
	pr, pw := io.Pipe()
	logReader := logreader.New(pr)

	logs, err := logReader.Read()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	pw.Write([]byte("{\"stream\":\"Step 1/5 : FROM alpine:3.12\\n\"}\n{\"stream\":\"Step 2/5\\n\"}\n"))
	assertCurrentLogCountAtLeast(t, logs, 2)

	pw.Write([]byte("{\"stream\":\"Step 3/5\\n\"}\n"))
	assertCurrentLogCountAtLeast(t, logs, 1)

	pw.Close()
}

// TestLogReader_ReadTwice is a regression test: LogReader.Read used to spawn
// a new unsynchronized goroutine on every call, all reading the same
// underlying io.ReadCloser concurrently — undefined behavior that go test
// -race correctly flagged as a data race. A second call must now fail
// instead of racing the first call's goroutine over the same stream.
func TestLogReader_ReadTwice(t *testing.T) {
	logReader := logreader.New(io.NopCloser(bytes.NewReader([]byte(`{"stream":"hi\n"}` + "\n"))))

	if _, err := logReader.Read(); err != nil {
		t.Fatalf("expected first Read to succeed, got %v", err)
	}

	if _, err := logReader.Read(); err != logreader.ErrorLogReaderAlreadyStarted {
		t.Fatalf("expected ErrorLogReaderAlreadyStarted, got %v", err)
	}
}

// A build's verdict is read out of this stream, so a stream that dies midway
// must not look like a stream that ended cleanly — otherwise the caller sees
// a truncated log with no errors in it and calls the build a success.
func TestLogReader_StreamBreakingMidwayReportsAnError(t *testing.T) {
	pr, pw := io.Pipe()
	logReader := logreader.New(pr)

	logs, err := logReader.Read()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	go func() {
		pw.Write([]byte("{\"stream\":\"Step 1/5\\n\"}\n"))
		pw.CloseWithError(errors.New("connection reset by peer"))
	}()

	var levels []buildloglevel.Level
	for l := range logs {
		levels = append(levels, l.Level())
	}
	if len(levels) == 0 || levels[len(levels)-1] != buildloglevel.Error {
		t.Fatalf("expected the broken stream to end with an Error entry, got %v", levels)
	}
}

// The counterpart: Close() is this process tearing the stream down on
// purpose (a cancelled build), so the read error it causes must stay silent
// — inventing an Error entry there would turn a cancellation into a failure
// in the build log.
func TestLogReader_DeliberateCloseReportsNothing(t *testing.T) {
	pr, pw := io.Pipe()
	logReader := logreader.New(pr)

	logs, err := logReader.Read()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	go func() {
		pw.Write([]byte("{\"stream\":\"Step 1/5\\n\"}\n"))
	}()
	if first, ok := <-logs; !ok {
		t.Fatal("expected the first log entry before the close")
	} else if first.Level() == buildloglevel.Error {
		t.Fatalf("unexpected error entry: %s", first.Message())
	}

	logReader.Close()

	for l := range logs {
		if l.Level() == buildloglevel.Error {
			t.Fatalf("a deliberate Close must not produce an error entry, got %q", l.Message())
		}
	}
}

func TestLogReader_ReadError(t *testing.T) {
	buildOutput := `
	abc
	`
	logReader := logreader.New(io.NopCloser(bytes.NewReader([]byte(buildOutput))))
	logReader.Close()
	logs, err := logReader.Read()
	if err == nil {
		t.Error("expected error")
	}

	if len(logs) != 0 {
		t.Error("expected 0 logs")
	}
}
