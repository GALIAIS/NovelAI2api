package novelai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Client struct {
	APIBase   string
	ImageBase string
	TextBase  string
	HTTP      *http.Client
}

func NewClient(apiBase, imageBase, textBase string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = newDefaultHTTPClient()
	}
	return &Client{APIBase: apiBase, ImageBase: imageBase, TextBase: textBase, HTTP: httpClient}
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   20,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, url string, token string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = BuildHeaders(token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.HTTP.Do(req)
}

func (c *Client) doJSON(ctx context.Context, method, url string, token string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	resp, err := c.do(ctx, method, url, token, "application/json", reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, _ := io.ReadAll(resp.Body)
		return &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) doBytes(ctx context.Context, method, url string, token string, contentType string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	resp, err := c.do(ctx, method, url, token, contentType, reader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, readErr := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &UpstreamError{StatusCode: resp.StatusCode, Body: buf}
	}
	if readErr != nil {
		return nil, readErr
	}
	return buf, nil
}
