package enrow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.enrow.io"

type Client struct {
	Email        *EmailFinder
	Verify       *EmailVerifier
	Phone        *PhoneFinder
	ReverseEmail *ReverseEmailResource
	Account      *AccountResource

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
	c.ReverseEmail = &ReverseEmailResource{client: c}
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
	}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		apiErr.Message = string(body)
	}

	switch status {
	case 401:
		return &AuthenticationError{EnrowError{Status: status, Err: apiErr.Error, Message: apiErr.Message}}
	case 402:
		return &InsufficientBalanceError{EnrowError{Status: status, Err: apiErr.Error, Message: apiErr.Message}}
	case 429:
		return &RateLimitError{EnrowError: EnrowError{Status: status, Err: apiErr.Error, Message: apiErr.Message}}
	default:
		return &EnrowError{Status: status, Err: apiErr.Error, Message: apiErr.Message}
	}
}

func poll[T any](fetcher func() (*T, error), isDone func(*T) bool, interval, timeout time.Duration) (*T, error) {
	if interval == 0 {
		interval = 2 * time.Second
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	start := time.Now()
	for {
		result, err := fetcher()
		if err != nil {
			return nil, err
		}
		if isDone(result) {
			return result, nil
		}
		if time.Since(start) >= timeout {
			return result, nil
		}
		time.Sleep(interval)
	}
}

// PollOptions controls auto-polling behavior.
type PollOptions struct {
	WaitForResult bool
	PollInterval  time.Duration
	Timeout       time.Duration
}

func defaultPoll(opts *PollOptions) *PollOptions {
	if opts == nil {
		return &PollOptions{}
	}
	return opts
}

func formatPath(pattern string, args ...any) string {
	return fmt.Sprintf(pattern, args...)
}
