package docker

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readStdOutputFrom(frames []byte, maxBytes int64) (string, string, error) {
	return readStdOutput(bufio.NewReader(bytes.NewReader(frames)), maxBytes)
}

func TestReadStdOutput(t *testing.T) {
	t.Run("normalizes lines and routes streams", func(t *testing.T) {
		frames := append(muxFrame(1, "hello\nworld\n"), muxFrame(2, "oops\n")...)
		stdout, stderr, err := readStdOutputFrom(frames, 0)
		assert.NoError(t, err)
		assert.Equal(t, "hello\nworld\n", stdout)
		assert.Equal(t, "oops\n", stderr)
	})

	t.Run("truncates combined output past the cap and marks it", func(t *testing.T) {
		// 16 stdout bytes with a 10-byte budget: 10 kept, the rest dropped.
		stdout, _, err := readStdOutputFrom(muxFrame(1, "abcdefghijklmnop"), 10)
		assert.NoError(t, err)
		assert.Contains(t, stdout, "abcdefghij")
		assert.NotContains(t, stdout, "klmnop")
		assert.Contains(t, stdout, "truncated")
	})

	t.Run("shares one budget across stdout and stderr", func(t *testing.T) {
		// 8 stdout + 8 stderr with a 10-byte budget: stdout takes its 8, stderr
		// only gets 2 before the shared budget is gone.
		frames := append(muxFrame(1, "AAAAAAAA"), muxFrame(2, "BBBBBBBB")...)
		stdout, stderr, err := readStdOutputFrom(frames, 10)
		assert.NoError(t, err)
		assert.Equal(t, "AAAAAAAA\n", stdout[:9])
		assert.Contains(t, strings.TrimSpace(stderr), "BB")
		assert.NotContains(t, stderr, "BBBBBBBB")
	})

	t.Run("does not drop a line longer than the old 64KiB scanner limit", func(t *testing.T) {
		long := strings.Repeat("a", 100_000)
		stdout, _, err := readStdOutputFrom(muxFrame(1, long), 0)
		assert.NoError(t, err)
		// Full line preserved plus the trailing newline normalizeLines adds.
		assert.Equal(t, 100_001, len(stdout))
	})
}
