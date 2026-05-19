package osm

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type RateLimitedClient struct {
	Client   *http.Client
	UserAgent string
	Ticker   *time.Ticker
	done     chan struct{}
}

func NewRateLimitedClient(requestsPerSecond int) *RateLimitedClient {
	if requestsPerSecond < 1 {
		requestsPerSecond = 1
	}
	interval := time.Second / time.Duration(requestsPerSecond)
	r := &RateLimitedClient{
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
		UserAgent: "Sunpath/1.0 (solar analysis tool)",
		Ticker:    time.NewTicker(interval),
		done:      make(chan struct{}),
	}
	return r
}

func (r *RateLimitedClient) Stop() {
	r.Ticker.Stop()
	select {
	case r.done <- struct{}{}:
	default:
	}
}

func (r *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	return r.doWithRetry(req, 0)
}

func (r *RateLimitedClient) doWithRetry(req *http.Request, attempt int) (*http.Response, error) {
	<-r.Ticker.C

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < 3 {
		resp.Body.Close()
		backoff := time.Duration(math.Pow(2, float64(attempt+2))) * time.Second
		if backoff > 15*time.Second {
			backoff = 15 * time.Second
		}
		time.Sleep(backoff)
		return r.doWithRetry(req, attempt+1)
	}

	return resp, nil
}

func (r *RateLimitedClient) DoRequest(method, urlStr string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", r.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return r.Do(req)
}
