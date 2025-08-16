package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/m-mizutani/goerr"
)

// PayloadVerifier is an interface for verifying Slack request signatures
type PayloadVerifier interface {
	Verify(body []byte, timestamp, signature string) error
}

// Verifier implements Slack signature verification
type Verifier struct {
	signingSecret string
}

// NewVerifier creates a new Slack signature verifier
func NewVerifier(signingSecret string) *Verifier {
	return &Verifier{
		signingSecret: signingSecret,
	}
}

// Verify checks if the request signature is valid
func (v *Verifier) Verify(body []byte, timestamp, signature string) error {
	if v.signingSecret == "" {
		// Skip verification if signing secret is not configured
		return nil
	}

	// Check timestamp to prevent replay attacks (5 minutes window)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return goerr.Wrap(err, "failed to parse timestamp")
	}

	now := time.Now().Unix()
	if now-ts > 60*5 {
		return goerr.New("request timestamp is too old")
	}

	// Create base string for signature
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))

	// Calculate expected signature
	h := hmac.New(sha256.New, []byte(v.signingSecret))
	h.Write([]byte(baseString))
	expectedSig := "v0=" + hex.EncodeToString(h.Sum(nil))

	// Compare signatures
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return goerr.New("invalid signature")
	}

	return nil
}
