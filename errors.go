package enrow

import "fmt"

// EnrowError is an error answered by the API. Status is the HTTP status code
// and Message the description sent by the API.
type EnrowError struct {
	Status  int
	Err     string
	Message string
}

func (e *EnrowError) Error() string {
	msg := e.Message
	if e.Err != "" {
		if msg != "" {
			msg = e.Err + ": " + msg
		} else {
			msg = e.Err
		}
	}
	if msg == "" {
		return fmt.Sprintf("enrow: HTTP %d", e.Status)
	}
	return fmt.Sprintf("enrow: HTTP %d: %s", e.Status, msg)
}

// AuthenticationError is a 401: missing or invalid API key, or an account
// without access to the feature.
type AuthenticationError struct {
	EnrowError
}

// InsufficientBalanceError is a 402: not enough credits.
type InsufficientBalanceError struct {
	EnrowError
}

// RateLimitError is a 429: too many requests (10 per second per API key on
// POST endpoints). Retry with an exponential backoff.
type RateLimitError struct {
	EnrowError
	RetryAfter int
}
