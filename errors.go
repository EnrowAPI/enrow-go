package enrow

import "fmt"

type EnrowError struct {
	Status  int
	Err     string
	Message string
}

func (e *EnrowError) Error() string {
	return fmt.Sprintf("enrow: %d %s — %s", e.Status, e.Err, e.Message)
}

type AuthenticationError struct {
	EnrowError
}

type InsufficientBalanceError struct {
	EnrowError
}

type RateLimitError struct {
	EnrowError
	RetryAfter int
}
