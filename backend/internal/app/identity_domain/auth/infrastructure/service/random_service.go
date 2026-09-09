package service

import (
	"crypto/rand"

	"judge-backend/internal/app/identity_domain/auth/application/service"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// maxUnbiased is the largest multiple of len(letterBytes) that fits in a byte.
// A random byte >= this value would skew the distribution under a plain modulo,
// so it is discarded and re-drawn instead — cheap rejection sampling that keeps
// every letter equally likely.
const maxUnbiased = 256 - (256 % len(letterBytes))

type RandomService struct{}

func NewRandomService() service.RandomService {
	return &RandomService{}
}

// RandString returns an n-character token drawn from letterBytes using
// crypto/rand. This is the source of the refresh token, a long-lived bearer
// credential and the Redis lookup key for a session, so it must be
// cryptographically unpredictable — math/rand (a single process-seeded PRNG
// whose state is recoverable from a few outputs) is not acceptable here.
func (s *RandomService) RandString(n int) string {
	b := make([]byte, n)
	// A generous buffer keeps this to one Read in the common case while leaving
	// room to skip the rejected bytes.
	buf := make([]byte, n+n/4+8)
	bi := len(buf)
	for i := 0; i < n; {
		if bi >= len(buf) {
			if _, err := rand.Read(buf); err != nil {
				// crypto/rand.Read never returns an error on supported
				// platforms; if it somehow did, panicking is correct — issuing a
				// predictable token would be worse than failing the request.
				panic(err)
			}
			bi = 0
		}
		c := buf[bi]
		bi++
		if int(c) >= maxUnbiased {
			continue
		}
		b[i] = letterBytes[int(c)%len(letterBytes)]
		i++
	}
	return string(b)
}
