package client_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zaidejjo/ax/internal/client"
)

// ─── Unit Tests ─────────────────────────────────────────────────────────────

func TestNewClient_Defaults(t *testing.T) {
	c := client.New()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestDo_NoURL(t *testing.T) {
	c := client.New()
	_, err := c.Do(&client.Request{Method: "GET", URL: ""})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestDo_NilRequest(t *testing.T) {
	c := client.New()
	_, err := c.Do(nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestDo_EmptyMethodDefaultsToGet(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Do(&client.Request{
		Method: "",
		URL:    ts.URL,
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_Timeout(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // longer than our timeout
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New(client.WithTimeout(1 * time.Nanosecond))
	_, err := c.Do(&client.Request{
		Method: "GET",
		URL:    ts.URL,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestDo_InvalidMethod(t *testing.T) {
	// An empty method should default to GET (not error), but a truly bogus
	// method containing spaces will cause http.NewRequest to fail.
	c := client.New()
	_, err := c.Do(&client.Request{Method: "GE T", URL: "http://example.com"})
	if err == nil {
		t.Fatal("expected error for malformed method containing spaces")
	}
}

// ─── Options Tests ───────────────────────────────────────────────────────────

func TestWithInsecureSkipVerify(t *testing.T) {
	c := client.New(client.WithInsecureSkipVerify())
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithDisableRedirects(t *testing.T) {
	c := client.New(client.WithDisableRedirects())
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithMaxBodySize(t *testing.T) {
	c := client.New(client.WithMaxBodySize(1024))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithTimeout(t *testing.T) {
	c := client.New(client.WithTimeout(5 * time.Second))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// ─── Integration Tests (httptest-based, no external deps) ────────────────────

func TestDo_GET(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Do(&client.Request{
		Method: "GET",
		URL:    ts.URL,
	})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if resp.Body == nil {
		t.Fatal("expected non-nil body")
	}
	if string(resp.Body) != `{"status":"ok"}` {
		t.Errorf("unexpected body: %s", resp.Body)
	}
	if resp.BodySize != int64(len(`{"status":"ok"}`)) {
		t.Errorf("expected body size %d, got %d", len(`{"status":"ok"}`), resp.BodySize)
	}
	if resp.Duration <= 0 {
		t.Error("expected positive duration")
	}
	if resp.Proto == "" {
		t.Error("expected non-empty proto")
	}
	if resp.RawRequest == "" {
		t.Error("expected non-empty raw request dump")
	}
	if resp.RawResponse == "" {
		t.Error("expected non-empty raw response dump")
	}
}

func TestDo_POST(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", contentType)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"received":true}`)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Post(
		ts.URL,
		map[string]string{"Content-Type": "application/json"},
		[]byte(`{"hello":"world"}`),
	)
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_PUT(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Put(ts.URL, nil, []byte(`{"update":true}`))
	if err != nil {
		t.Fatalf("PUT failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_PATCH(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Patch(ts.URL, nil, []byte(`{"patch":true}`))
	if err != nil {
		t.Fatalf("PATCH failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_DELETE(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Delete(ts.URL, nil)
	if err != nil {
		t.Fatalf("DELETE failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDo_HEAD(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("expected HEAD, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Head(ts.URL, nil)
	if err != nil {
		t.Fatalf("HEAD failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_OPTIONS(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			t.Errorf("expected OPTIONS, got %s", r.Method)
		}
		w.Header().Set("Allow", "GET, POST, PUT, DELETE, OPTIONS")
		w.WriteHeader(http.StatusNoContent)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Options(ts.URL, nil)
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	if resp.Headers.Get("Allow") == "" {
		t.Error("expected Allow header")
	}
}

// ─── Headers ─────────────────────────────────────────────────────────────────

func TestDo_CustomHeaders(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "test-value" {
			t.Errorf("expected X-Custom: test-value, got: %s", r.Header.Get("X-Custom"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Authorization header")
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Do(&client.Request{
		Method: "GET",
		URL:    ts.URL,
		Headers: map[string]string{
			"X-Custom":      "test-value",
			"Authorization": "Bearer test-token",
		},
	})
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_DefaultUserAgent(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != "ax/1.0" {
			t.Errorf("expected User-Agent ax/1.0, got %s", r.UserAgent())
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDo_CustomUserAgentOverridesDefault(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		expectedUA := "MyCustom/2.0"
		if r.UserAgent() != expectedUA {
			t.Errorf("expected User-Agent %s, got %s", expectedUA, r.UserAgent())
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, map[string]string{"User-Agent": "MyCustom/2.0"})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── Body Size Limit ─────────────────────────────────────────────────────────

func TestMaximumBodySize(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Send 4 KB of data.
		body := strings.Repeat("x", 4096)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	})
	defer ts.Close()

	// Set a tiny limit — 1 KB.
	c := client.New(client.WithMaxBodySize(1024))
	_, err := c.Get(ts.URL, nil)
	if err == nil {
		t.Fatal("expected ErrBodyTooBig for 4 KB response with 1 KB limit")
	}
}

func TestBodySizeLimitEdge_ExactMatch(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body := strings.Repeat("x", 100)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	})
	defer ts.Close()

	c := client.New(client.WithMaxBodySize(100))
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("expected success for exact limit match: %v", err)
	}
	if resp.BodySize != 100 {
		t.Errorf("expected 100 bytes, got %d", resp.BodySize)
	}
}

func TestBodySizeLimitEdge_OneOver(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body := strings.Repeat("x", 101)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	})
	defer ts.Close()

	c := client.New(client.WithMaxBodySize(100))
	_, err := c.Get(ts.URL, nil)
	if err == nil {
		t.Fatal("expected ErrBodyTooBig for 101 bytes with 100 byte limit")
	}
}

// ─── Redirects ───────────────────────────────────────────────────────────────

func TestRedirect_FollowedByDefault(t *testing.T) {
	redirectCount := 0
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status":"done"}`)
			return
		}
		redirectCount++
		http.Redirect(w, r, ts.URL+"/final", http.StatusFound)
	}))
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET with redirect failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 after redirect, got %d", resp.StatusCode)
	}
	if redirectCount != 1 {
		t.Errorf("expected 1 redirect, got %d", redirectCount)
	}
}

func TestRedirect_Disabled(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ts.URL+"/new-location", http.StatusFound)
	}))
	defer ts.Close()

	c := client.New(client.WithDisableRedirects())
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET with redirects disabled failed: %v", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
}

// ─── Status Codes ────────────────────────────────────────────────────────────

func TestStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"NoContent", http.StatusNoContent},
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"Forbidden", http.StatusForbidden},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})
			defer ts.Close()

			c := client.New()
			resp, err := c.Get(ts.URL, nil)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tt.statusCode {
				t.Errorf("expected %d, got %d", tt.statusCode, resp.StatusCode)
			}
		})
	}
}

// ─── Response Metadata ───────────────────────────────────────────────────────

func TestResponse_Proto(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	// The test server uses HTTP/1.1.
	if resp.Proto != "HTTP/1.1" {
		t.Errorf("expected HTTP/1.1, got %s", resp.Proto)
	}
}

func TestResponse_RawRequestContainsMethod(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if !strings.Contains(resp.RawRequest, "GET") {
		t.Error("raw request should contain GET method")
	}
	if !strings.Contains(resp.RawRequest, "Host:") {
		t.Error("raw request should contain Host header")
	}
}

func TestResponse_RawResponseContainsStatusCode(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"not found"}`)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if !strings.Contains(resp.RawResponse, "404") {
		t.Error("raw response should contain 404 status")
	}
}

// ─── Large Body ──────────────────────────────────────────────────────────────

func TestLargeResponse(t *testing.T) {
	const size = 100_000 // 100 KB
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body := strings.Repeat("ABCDEFGHIJ", size/10)
		fmt.Fprint(w, body)
	})
	defer ts.Close()

	c := client.New(client.WithMaxBodySize(200_000))
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("large GET failed: %v", err)
	}
	if resp.BodySize != size {
		t.Errorf("expected %d bytes, got %d", size, resp.BodySize)
	}
}

// ─── JSON Response ───────────────────────────────────────────────────────────

func TestJSONResponse(t *testing.T) {
	type testPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		payload := testPayload{Name: "test", Value: 42}
		json.NewEncoder(w).Encode(payload)
	})
	defer ts.Close()

	c := client.New()
	resp, err := c.Get(ts.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	var result testPayload
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if result.Name != "test" || result.Value != 42 {
		t.Errorf("unexpected payload: %+v", result)
	}
}

// ─── Helper ──────────────────────────────────────────────────────────────────

// newTestServer creates an httptest.Server with common settings.
func newTestServer(handler func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}
