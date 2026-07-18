// Package client provides a pure Go HTTP client wrapper for the ax TUI API client.
//
// It builds on net/http with safety limits, structured responses, and a clean
// API for the TUI layer. It has zero external dependencies beyond the standard
// library, making it fully cross-platform without requiring curl or any other
// external tool.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// ─── Errors ──────────────────────────────────────────────────────────────────

var (
	// ErrNoURL indicates no URL was provided in the request.
	ErrNoURL = errors.New("client: URL is required")

	// ErrBodyTooBig indicates the response body exceeded the configured limit.
	ErrBodyTooBig = errors.New("client: response body exceeds maximum size")

	// ErrInvalidMethod indicates the HTTP method is not recognized.
	ErrInvalidMethod = errors.New("client: invalid HTTP method")
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	// DefaultTimeout is the default per-request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxBodySize is the maximum response body we will read (10 MB).
	DefaultMaxBodySize = 10 * 1024 * 1024
)

// ─── Types ───────────────────────────────────────────────────────────────────

// Request describes an HTTP request to execute.
type Request struct {
	// Method is the HTTP method (GET, POST, PUT, etc.). Empty defaults to GET.
	Method string

	// URL is the full request URL (required).
	URL string

	// Headers are key-value pairs to set on the request.
	Headers map[string]string

	// Body is the raw request body bytes.
	Body []byte
}

// Response contains the complete result of an HTTP request.
type Response struct {
	// StatusCode is the HTTP status code (e.g., 200, 404, 500).
	StatusCode int

	// Status is the human-readable status (e.g., "200 OK").
	Status string

	// Proto is the HTTP protocol version (e.g., "HTTP/1.1").
	Proto string

	// Headers contains the response headers.
	Headers http.Header

	// Body contains the raw response body bytes.
	Body []byte

	// BodySize is the size of Body in bytes.
	BodySize int64

	// Duration is the total request-response time.
	Duration time.Duration

	// RawRequest is the HTTP wire representation of the outgoing request.
	RawRequest string

	// RawResponse is the HTTP wire representation of the incoming response
	// (status line + headers only; body excluded since it is already in Body).
	RawResponse string
}

// Client wraps net/http.Client with safety limits and convenience features.
type Client struct {
	httpClient         *http.Client
	timeout            time.Duration
	maxBodySize        int64
	insecureSkipVerify bool
	disableRedirects   bool
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// ─── Options ─────────────────────────────────────────────────────────────────

// WithTimeout sets the per-request timeout. Set to 0 to disable timeouts.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithInsecureSkipVerify disables TLS certificate verification.
// Useful for testing with self-signed certificates.
func WithInsecureSkipVerify() Option {
	return func(c *Client) { c.insecureSkipVerify = true }
}

// WithDisableRedirects prevents the client from automatically following
// HTTP redirects. The redirect response itself will be returned.
func WithDisableRedirects() Option {
	return func(c *Client) { c.disableRedirects = true }
}

// WithMaxBodySize sets the maximum response body size in bytes.
// Bodies larger than this will return ErrBodyTooBig.
func WithMaxBodySize(n int64) Option {
	return func(c *Client) { c.maxBodySize = n }
}

// ─── Constructor ─────────────────────────────────────────────────────────────

// New creates a new Client with the given options.
//
// Usage:
//
//	c := client.New(
//	    client.WithTimeout(10 * time.Second),
//	    client.WithInsecureSkipVerify(),
//	)
func New(opts ...Option) *Client {
	c := &Client{
		timeout:     DefaultTimeout,
		maxBodySize: DefaultMaxBodySize,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.buildHTTPClient()
	return c
}

func (c *Client) buildHTTPClient() {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.insecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		},
		// Good defaults from the standard library (Go 1.26 sets these sanely).
	}

	client := &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}

	if c.disableRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	c.httpClient = client
}

// ─── Core Execute ────────────────────────────────────────────────────────────

// Do executes an HTTP request and returns a structured Response.
//
// It handles method defaulting (GET), header injection, body streaming, size
// limiting, timing, and raw request/response capture for display purposes.
func (c *Client) Do(req *Request) (*Response, error) {
	// ── Validate ──────────────────────────────────────────────────────────
	if req == nil {
		return nil, errors.New("client: request is nil")
	}
	if req.URL == "" {
		return nil, ErrNoURL
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	if !isValidMethod(method) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidMethod, req.Method)
	}

	// ── Build *http.Request ──────────────────────────────────────────────
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("client: failed to create request: %w", err)
	}

	// Set user headers.
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set a default User-Agent if none was provided.
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "ax/1.0")
	}

	// ── Capture raw request ──────────────────────────────────────────────
	rawReq, _ := httputil.DumpRequestOut(httpReq, true)

	// ── Execute ──────────────────────────────────────────────────────────
	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		// Check for timeout specifically.
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("client: request timed out after %v", c.timeout)
		}
		return nil, fmt.Errorf("client: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// ── Capture raw response (status + headers; body captured separately) ─
	rawResp, _ := httputil.DumpResponse(httpResp, false)

	// ── Read body with size limit ────────────────────────────────────────
	body, err := c.readLimitedBody(httpResp.Body)
	if err != nil {
		return nil, err
	}

	// ── Assemble result ──────────────────────────────────────────────────
	resp := &Response{
		StatusCode:  httpResp.StatusCode,
		Status:      httpResp.Status,
		Proto:       httpResp.Proto,
		Headers:     httpResp.Header.Clone(),
		Body:        body,
		BodySize:    int64(len(body)),
		Duration:    duration,
		RawRequest:  string(rawReq),
		RawResponse: string(rawResp),
	}

	return resp, nil
}

// readLimitedBody reads the response body up to maxBodySize+1 bytes.
// If the body exceeds maxBodySize, ErrBodyTooBig is returned and the
// already-read data is discarded.
func (c *Client) readLimitedBody(r io.Reader) ([]byte, error) {
	// Read up to maxBodySize+1 so we can detect overflow.
	limited := io.LimitReader(r, c.maxBodySize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("client: failed to read response body: %w", err)
	}
	if int64(len(body)) > c.maxBodySize {
		return nil, ErrBodyTooBig
	}
	return body, nil
}

// ─── Convenience Methods ─────────────────────────────────────────────────────

// Get executes a GET request.
func (c *Client) Get(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{Method: http.MethodGet, URL: url, Headers: headers})
}

// Post executes a POST request with an optional body.
func (c *Client) Post(url string, headers map[string]string, body []byte) (*Response, error) {
	return c.Do(&Request{Method: http.MethodPost, URL: url, Headers: headers, Body: body})
}

// Put executes a PUT request with an optional body.
func (c *Client) Put(url string, headers map[string]string, body []byte) (*Response, error) {
	return c.Do(&Request{Method: http.MethodPut, URL: url, Headers: headers, Body: body})
}

// Patch executes a PATCH request with an optional body.
func (c *Client) Patch(url string, headers map[string]string, body []byte) (*Response, error) {
	return c.Do(&Request{Method: http.MethodPatch, URL: url, Headers: headers, Body: body})
}

// Delete executes a DELETE request.
func (c *Client) Delete(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{Method: http.MethodDelete, URL: url, Headers: headers})
}

// Head executes a HEAD request.
func (c *Client) Head(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{Method: http.MethodHead, URL: url, Headers: headers})
}

// Options executes an OPTIONS request.
func (c *Client) Options(url string, headers map[string]string) (*Response, error) {
	return c.Do(&Request{Method: http.MethodOptions, URL: url, Headers: headers})
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// isValidMethod checks whether method is a known HTTP method.
func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost,
		http.MethodPut, http.MethodPatch, http.MethodDelete,
		http.MethodConnect, http.MethodOptions, http.MethodTrace:
		return true
	default:
		// Allow arbitrary methods for extensibility (WebDAV, etc.).
		return len(method) > 0
	}
}
