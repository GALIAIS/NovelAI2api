package novelai

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"time"
)

const correlationAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func BuildHeaders(token string) http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("x-correlation-id", mustCorrelationID(6))
	h.Set("x-initiated-at", time.Now().UTC().Format(time.RFC3339Nano))
	if token != "" {
		h.Set("Authorization", "Bearer "+token)
	}
	return h
}

func mustCorrelationID(length int) string {
	out := make([]byte, length)
	max := big.NewInt(int64(len(correlationAlphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = correlationAlphabet[i%len(correlationAlphabet)]
			continue
		}
		out[i] = correlationAlphabet[n.Int64()]
	}
	return string(out)
}
