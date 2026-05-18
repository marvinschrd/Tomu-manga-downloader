package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
func (c *Client) Get(path string, params url.Values, out interface{}) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

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
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limited by %s — try again later", c.baseURL)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s%s", resp.StatusCode, c.baseURL, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
