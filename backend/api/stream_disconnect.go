package api

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"syscall"
)

func IsClientDisconnectedError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, io.ErrClosedPipe) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		if errors.Is(netErr.Err, os.ErrClosed) || errors.Is(netErr.Err, syscall.EPIPE) || errors.Is(netErr.Err, syscall.ECONNRESET) {
			return true
		}
	}
	return false
}
