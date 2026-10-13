package enrow

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.enrow.io"

const (
	defaultPollInterval = 2 * time.Second
	defaultPollTimeout  = 30 * time.Second

	// ongoing is the qualification of a search or verification still running.
	ongoing = "ongoing"
)

type Client struct {
	Email   *EmailFinder
	Verify  *EmailVerifier
	Phone   *PhoneFinder
	Account *AccountResource

	apiKey  string
	baseURL string
	http    *http.Client
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.http = hc }
}

func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}

	c.Email = &EmailFinder{client: c}
	c.Verify = &EmailVerifier{client: c}
	c.Phone = &PhoneFinder{client: c}
	c.Account = &AccountResource{client: c}

	return c
}

func (c *Client) doPost(path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

func (c *Client) doGet(path string, params map[string]string, result any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", c.apiKey)

	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseError(resp.StatusCode, body)
	}

	return json.Unmarshal(body, result)
}

func parseError(status int, body []byte) error {
	var apiErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		// A 402 on a single endpoint answers {"reason": "...", "success": false}.
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		apiErr.Message = strings.TrimSpace(string(body))
	}
	if apiErr.Message == "" {
		apiErr.Message = apiErr.Reason
	}

	base := EnrowError{Status: status, Err: apiErr.Error, Message: apiErr.Message}
	switch status {
	case http.StatusUnauthorized:
		return &AuthenticationError{base}
	case http.StatusPaymentRequired:
		return &InsufficientBalanceError{base}
	case http.StatusTooManyRequests:
		return &RateLimitError{EnrowError: base}
	default:
		return &base
	}
}

// PollOptions makes a single search or verification wait for its result: once
// it is launched, Get is called every PollInterval (default 2s) until it has
// ended or Timeout (default 30s) has elapsed. On timeout the result is returned
// as it is, with Qualification "ongoing" and its ID, so that Get can be called
// later. If a Get fails, the error is returned along with the launched result,
// whose ID stays usable.
type PollOptions struct {
	WaitForResult bool
	PollInterval  time.Duration
	Timeout       time.Duration
}

func pollOptions(opts []*PollOptions) PollOptions {
	if len(opts) > 0 && opts[0] != nil {
		return *opts[0]
	}
	return PollOptions{}
}

// poll waits interval, calls fetch, and repeats until done reports true or
// timeout has elapsed. It returns the last result either way.
func poll[T any](fetch func() (*T, error), done func(*T) bool, interval, timeout time.Duration) (*T, error) {
	if interval <= 0 {
		interval = defaultPollInterval
	}
	if timeout <= 0 {
		timeout = defaultPollTimeout
	}

	deadline := time.Now().Add(timeout)
	for {
		wait := interval
		if remaining := time.Until(deadline); remaining < wait {
			wait = remaining
		}
		if wait > 0 {
			time.Sleep(wait)
		}

		result, err := fetch()
		if err != nil {
			return nil, err
		}
		if done(result) || !time.Now().Before(deadline) {
			return result, nil
		}
	}
}
