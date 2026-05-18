package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const maxRetries = 4

// Client is a generic JSON HTTP client shared by all source implementations.
type Client struct {
	baseURL    string
	headers    map[string]string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		headers:    map[string]string{},
		httpClient: &http.Client{},
	}
}

// WithHeader returns a copy of the client with an additional default header.
// Use this for API keys, auth tokens, or User-Agent overrides.
func (c *Client) WithHeader(key, value string) *Client {
	copy := &Client{
		baseURL:    c.baseURL,
		httpClient: c.httpClient,
		headers:    make(map[string]string, len(c.headers)+1),
	}
	for k, v := range c.headers {
		copy.headers[k] = v
	}
	copy.headers[key] = value
	return copy
}

// Get performs a GET request and decodes the JSON response into out.
// Retries up to maxRetries times on HTTP 429, honouring Retry-After when present.
func (c *Client) Get(path string, params url.Values, out interface{}) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	delay := time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := parseRetryAfter(resp.Header.Get("Retry-After"), delay)
			resp.Body.Close()
			if attempt == maxRetries-1 {
				return fmt.Errorf("rate limited by %s — giving up after %d retries", c.baseURL, maxRetries)
			}
			fmt.Printf("  rate limited — retrying in %v...\n", wait.Round(time.Second))
			time.Sleep(wait)
			delay *= 2
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("HTTP %d from %s%s", resp.StatusCode, c.baseURL, path)
		}
		err = json.NewDecoder(resp.Body).Decode(out)
		resp.Body.Close()
		return err
	}
	return fmt.Errorf("rate limited by %s — giving up", c.baseURL)
}

func parseRetryAfter(header string, fallback time.Duration) time.Duration {
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return fallback
}
