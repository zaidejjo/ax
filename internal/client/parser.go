package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// ─── Errors ──────────────────────────────────────────────────────────────────

var (
	// ErrEmptyInput is returned when the parser receives an empty string.
	ErrEmptyInput = errors.New("parser: input is empty")

	// ErrParseNoURL is returned when no URL token is found.
	ErrParseNoURL = errors.New("parser: URL is required")
)

// ─── Parse ───────────────────────────────────────────────────────────────────

// Parse parses an xh/httpie-style single-line input into a *Request.
//
// Format: [METHOD] URL [ITEM...]
//
// ITEM can be one of:
//
//	Key:Value      → HTTP header
//	key==value     → JSON body field
//	key=value      → form-encoded field
//
// URL shorthands:
//
//	:port/path     → http://localhost:port/path
//	host/path      → https://host/path (scheme defaults to https)
//
// Method detection:
//
//	If the first token is an HTTP method (GET, POST, PUT, etc.), it is used.
//	If no method is specified and JSON or form body fields are present, POST is used.
//	Otherwise, GET is used.
func Parse(input string) (*Request, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, ErrEmptyInput
	}

	tokens := tokenize(input)
	if len(tokens) == 0 {
		return nil, ErrEmptyInput
	}

	idx := 0

	// ── Detect explicit HTTP method ────────────────────────────────────
	var method string
	if isHTTPMethod(tokens[idx]) {
		method = strings.ToUpper(tokens[idx])
		idx++
	}

	// ── Extract URL ────────────────────────────────────────────────────
	if idx >= len(tokens) {
		return nil, ErrParseNoURL
	}
	rawURL := tokens[idx]
	idx++

	// ── Parse remaining tokens as items ────────────────────────────────
	var headers map[string]string
	jsonFields := make(map[string]string)
	formFields := make(map[string]string)

	for ; idx < len(tokens); idx++ {
		token := tokens[idx]
		switch classifyItem(token) {
		case itemJSONField:
			k, v := splitItem(token, "==")
			jsonFields[k] = v
		case itemFormField:
			k, v := splitItem(token, "=")
			formFields[k] = v
		case itemHeader:
			k, v := splitItem(token, ":")
			if headers == nil {
				headers = make(map[string]string)
			}
			headers[k] = v
		}
	}

	// ── Normalize URL ──────────────────────────────────────────────────
	normalizedURL := normalizeURL(rawURL)
	if normalizedURL == "" {
		return nil, fmt.Errorf("parser: invalid URL: %q", rawURL)
	}

	// ── Determine method ───────────────────────────────────────────────
	if method == "" {
		if len(jsonFields) > 0 || len(formFields) > 0 {
			method = "POST"
		} else {
			method = "GET"
		}
	}

	// ── Build body ─────────────────────────────────────────────────────
	var body []byte
	switch {
	case len(jsonFields) > 0:
		data := make(map[string]string, len(jsonFields))
		for k, v := range jsonFields {
			data[k] = v
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("parser: marshal JSON body: %w", err)
		}
		body = encoded
		if headers == nil {
			headers = make(map[string]string)
		}
		if _, exists := headers["Content-Type"]; !exists {
			headers["Content-Type"] = "application/json"
		}

	case len(formFields) > 0:
		vals := url.Values{}
		for k, v := range formFields {
			vals.Set(k, v)
		}
		body = []byte(vals.Encode())
		if headers == nil {
			headers = make(map[string]string)
		}
		if _, exists := headers["Content-Type"]; !exists {
			headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	}

	return &Request{
		Method:  method,
		URL:     normalizedURL,
		Headers: headers,
		Body:    body,
	}, nil
}

// ─── Item Classification ──────────────────────────────────────────────────────

type itemClass int

const (
	itemUnknown   itemClass = iota
	itemJSONField           // key==value
	itemFormField           // key=value
	itemHeader              // Key:Value
)

// classifyItem determines what kind of shorthand item a token represents.
//
// Priority:
//  1. Contains "=="     → JSON body field
//  2. Contains "="      → form-encoded field
//  3. Contains ":"      → header
//  4. Otherwise         → unknown
func classifyItem(token string) itemClass {
	// A token like "key==value" must have "==" before any single "=".
	if strings.Contains(token, "==") {
		return itemJSONField
	}
	if strings.Contains(token, "=") {
		return itemFormField
	}
	if strings.Contains(token, ":") {
		return itemHeader
	}
	return itemUnknown
}

// splitItem splits a token on the first occurrence of sep and trims
// whitespace from both parts.
func splitItem(token, sep string) (string, string) {
	before, after, _ := strings.Cut(token, sep)
	return strings.TrimSpace(before), strings.TrimSpace(after)
}

// ─── URL Normalisation ───────────────────────────────────────────────────────

// normalizeURL transforms URL shorthands into full URLs:
//
//	:8080/path   → http://localhost:8080/path
//	/path        → http://localhost/path
//	host/path    → https://host/path
func normalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// :port/path → http://localhost:port/path
	if strings.HasPrefix(raw, ":") {
		return "http://localhost" + raw
	}

	// /path → http://localhost/path
	if strings.HasPrefix(raw, "/") {
		return "http://localhost" + raw
	}

	// If it already has a scheme, return as-is.
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}

	// No scheme → prepend https://
	return "https://" + raw
}

// ─── Tokenizer ───────────────────────────────────────────────────────────────

// tokenize splits input into tokens, respecting single and double quotes.
// Quoted sections are treated as a single token with quotes stripped.
func tokenize(input string) []string {
	var tokens []string
	var buf strings.Builder
	inQuote := false
	var quoteChar byte

	for i := 0; i < len(input); i++ {
		ch := input[i]
		switch {
		case inQuote:
			if ch == quoteChar {
				inQuote = false
			} else {
				buf.WriteByte(ch)
			}
		case ch == '\'' || ch == '"':
			inQuote = true
			quoteChar = ch
		case ch == ' ' || ch == '\t':
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		default:
			buf.WriteByte(ch)
		}
	}

	if buf.Len() > 0 {
		tokens = append(tokens, buf.String())
	}

	return tokens
}

// ─── Method Detection ─────────────────────────────────────────────────────────

// isHTTPMethod returns true if token is a known HTTP method (case-insensitive).
func isHTTPMethod(token string) bool {
	switch strings.ToUpper(token) {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS",
		"CONNECT", "TRACE":
		return true
	default:
		return false
	}
}
