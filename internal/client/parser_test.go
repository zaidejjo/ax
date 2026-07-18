package client

import (
	"testing"
)

func TestParse_EmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParse_WhitespaceInput(t *testing.T) {
	_, err := Parse("   ")
	if err == nil {
		t.Fatal("expected error for whitespace input")
	}
}

func TestParse_JustURL(t *testing.T) {
	req, err := Parse("https://api.example.com/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "GET" {
		t.Fatalf("expected GET, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Fatalf("expected https://api.example.com/users, got %s", req.URL)
	}
	if len(req.Body) != 0 {
		t.Fatalf("expected empty body, got %d bytes", len(req.Body))
	}
}

func TestParse_ExplicitMethod(t *testing.T) {
	req, err := Parse("POST https://api.example.com/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %s", req.Method)
	}
}

func TestParse_MethodCaseInsensitive(t *testing.T) {
	req, err := Parse("post https://api.example.com/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %s", req.Method)
	}
}

func TestParse_LocalhostShorthand(t *testing.T) {
	req, err := Parse(":8080/api/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL != "http://localhost:8080/api/users" {
		t.Fatalf("expected http://localhost:8080/api/users, got %s", req.URL)
	}
}

func TestParse_PathOnlyShorthand(t *testing.T) {
	req, err := Parse("/api/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL != "http://localhost/api/users" {
		t.Fatalf("expected http://localhost/api/users, got %s", req.URL)
	}
}

func TestParse_NoSchemeDefaultsHTTPS(t *testing.T) {
	req, err := Parse("api.example.com/users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL != "https://api.example.com/users" {
		t.Fatalf("expected https://api.example.com/users, got %s", req.URL)
	}
}

func TestParse_Headers(t *testing.T) {
	req, err := Parse("https://example.com Authorization:Bearer\ntoken123 X-Custom:value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["Authorization"] != "Bearer\ntoken123" {
		t.Fatalf("expected Authorization header with value including newline, got %q", req.Headers["Authorization"])
	}
	// Wait, the \n is a literal newline in the test string. Let me reconsider...
	// Actually the parser doesn't handle newlines specially, they're just bytes.
	// Let me write a cleaner test.
}

func TestParse_HeadersSimple(t *testing.T) {
	req, err := Parse("https://example.com Authorization:token123 X-Custom:value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["Authorization"] != "token123" {
		t.Fatalf("expected Authorization: token123, got %q", req.Headers["Authorization"])
	}
	if req.Headers["X-Custom"] != "value" {
		t.Fatalf("expected X-Custom: value, got %q", req.Headers["X-Custom"])
	}
}

func TestParse_HeadersWithSpaces(t *testing.T) {
	req, err := Parse("https://example.com 'Content-Type: application/json'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type: application/json, got %q", req.Headers["Content-Type"])
	}
}

func TestParse_JSONBodyFields(t *testing.T) {
	req, err := Parse("https://example.com/data name==John age==30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST (auto-detected from JSON fields), got %s", req.Method)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type: application/json, got %q", req.Headers["Content-Type"])
	}
	if len(req.Body) == 0 {
		t.Fatal("expected non-empty body")
	}
	// Body should be valid JSON.
	body := string(req.Body)
	if body != `{"age":"30","name":"John"}` && body != `{"name":"John","age":"30"}` {
		t.Fatalf("unexpected JSON body: %s", body)
	}
}

func TestParse_JSONBodyFieldsWithExplicitMethod(t *testing.T) {
	req, err := Parse("PUT https://example.com/data name==Updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "PUT" {
		t.Fatalf("expected PUT (explicit), got %s", req.Method)
	}
}

func TestParse_FormBodyFields(t *testing.T) {
	req, err := Parse("https://example.com/login username=john password=secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST (auto-detected from form fields), got %s", req.Method)
	}
	if req.Headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Fatalf("expected Content-Type: application/x-www-form-urlencoded, got %q", req.Headers["Content-Type"])
	}
	body := string(req.Body)
	if body != "password=secret&username=john" && body != "username=john&password=secret" {
		t.Fatalf("unexpected form body: %s", body)
	}
}

func TestParse_JSONBodyPrecedenceOverForm(t *testing.T) {
	// When both == and = are present, == (JSON) should take precedence.
	req, err := Parse("https://example.com/data jsonField==value formField=other")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected JSON Content-Type, got %q", req.Headers["Content-Type"])
	}
}

func TestParse_MixedHeadersAndBody(t *testing.T) {
	req, err := Parse("https://example.com/data Authorization:token123 name==John")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["Authorization"] != "token123" {
		t.Fatalf("expected Authorization: token123, got %q", req.Headers["Authorization"])
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected JSON Content-Type, got %q", req.Headers["Content-Type"])
	}
}

func TestParse_ExplicitContentTypeNotOverridden(t *testing.T) {
	req, err := Parse(`https://example.com/data Content-Type:text/plain name==John`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit Content-Type should NOT be overridden by parser.
	if req.Headers["Content-Type"] != "text/plain" {
		t.Fatalf("expected Content-Type: text/plain (explicit), got %q", req.Headers["Content-Type"])
	}
}

func TestParse_QuotedValue(t *testing.T) {
	req, err := Parse(`https://example.com/data name=="John Doe"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(req.Body)
	if body != `{"name":"John Doe"}` {
		t.Fatalf("expected JSON with quoted value, got %s", body)
	}
}

func TestParse_QuotedHeader(t *testing.T) {
	req, err := Parse(`https://example.com 'User-Agent: ax/2.0'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Headers["User-Agent"] != "ax/2.0" {
		t.Fatalf("expected User-Agent: ax/2.0, got %q", req.Headers["User-Agent"])
	}
}

func TestParse_URLOnlyNoItems(t *testing.T) {
	req, err := Parse(":3000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL != "http://localhost:3000" {
		t.Fatalf("expected http://localhost:3000, got %s", req.URL)
	}
	if req.Method != "GET" {
		t.Fatalf("expected GET (default), got %s", req.Method)
	}
}

func TestParse_URLOnlyWithMethod(t *testing.T) {
	req, err := Parse("DELETE :8080/api/resource/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "DELETE" {
		t.Fatalf("expected DELETE, got %s", req.Method)
	}
	if req.URL != "http://localhost:8080/api/resource/1" {
		t.Fatalf("expected http://localhost:8080/api/resource/1, got %s", req.URL)
	}
}

func TestParse_OnlyMethodNoURL(t *testing.T) {
	_, err := Parse("GET")
	if err == nil {
		t.Fatal("expected error for method without URL")
	}
}

func TestParse_UnknownTokenIgnored(t *testing.T) {
	req, err := Parse("https://example.com some-unknown-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.URL != "https://example.com" {
		t.Fatalf("expected https://example.com, got %s", req.URL)
	}
}

func TestParse_EmptyFormField(t *testing.T) {
	req, err := Parse("https://example.com flag=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(req.Body)
	if body != "flag=" {
		t.Fatalf("expected form body 'flag=', got %s", body)
	}
}

func TestParse_MultipleMethods_OnlyFirstCounts(t *testing.T) {
	// If the URL itself contains a method-like word, only the first token matters.
	req, err := Parse("GET https://example.com/get-resource")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "GET" {
		t.Fatalf("expected GET, got %s", req.Method)
	}
	if req.URL != "https://example.com/get-resource" {
		t.Fatalf("expected https://example.com/get-resource, got %s", req.URL)
	}
}
