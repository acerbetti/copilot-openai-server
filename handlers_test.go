package main

import (
	"net/http"
	"os"
	"strings"
	"testing"

	copilot "github.com/github/copilot-sdk/go"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name     string
		messages []Message
		want     string
		wants    []string // Substrings we expect
		ignores  []string // Substrings we expect NOT to appear
	}{
		{
			name: "Basic user assistant interaction",
			messages: []Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there"},
			},
			wants: []string{
				"[User]: Hello",
				"[Assistant]: Hi there",
			},
		},
		{
			name: "System message should be ignored in prompt",
			messages: []Message{
				{Role: "system", Content: "You are a helpful assistant"},
				{Role: "user", Content: "Hello"},
			},
			wants: []string{
				"[User]: Hello",
			},
			ignores: []string{
				"[System]: You are a helpful assistant",
				"You are a helpful assistant",
			},
		},
		{
			name: "Multiple system messages should be ignored",
			messages: []Message{
				{Role: "system", Content: "Sys 1"},
				{Role: "user", Content: "User 1"},
				{Role: "system", Content: "Sys 2"},
			},
			wants: []string{
				"[User]: User 1",
			},
			ignores: []string{
				"Sys 1",
				"Sys 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPrompt(tt.messages)

			for _, want := range tt.wants {
				if !strings.Contains(got, want) {
					t.Errorf("buildPrompt() missing expected content %q. Got:\n%s", want, got)
				}
			}

			for _, ignore := range tt.ignores {
				if strings.Contains(got, ignore) {
					t.Errorf("buildPrompt() contained prohibited content %q. Got:\n%s", ignore, got)
				}
			}
		})
	}
}

func TestGetAPIKeyFromHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	if got := getAPIKeyFromHeader(req); got != "abc123" {
		t.Errorf("expected abc123, got %q", got)
	}

	req.Header.Set("Authorization", "bearer XYZ")
	if got := getAPIKeyFromHeader(req); got != "XYZ" {
		t.Errorf("case-insensitive Bearer failed, got %q", got)
	}

	req.Header.Set("Authorization", "Token abc")
	if got := getAPIKeyFromHeader(req); got != "" {
		t.Errorf("unexpected non-empty for wrong scheme: %q", got)
	}
}

func TestExtractAPIKey(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", strings.NewReader("{}"))
	some := "foo"
	body := ChatCompletionRequest{ApiKey: some}
	if got := extractAPIKey(req, &body); got != some {
		t.Errorf("expected body key, got %q", got)
	}

	req.Header.Set("Authorization", "Bearer bar")
	if got := extractAPIKey(req, &body); got != "bar" {
		t.Errorf("header should override body, got %q", got)
	}
}

func TestBuildClientEnv_PreservesBaseAndOverridesToken(t *testing.T) {
	t.Setenv("PATH", "/tmp/test-path")
	t.Setenv("HOME", "/tmp/test-home")
	t.Setenv("COPILOT_GITHUB_TOKEN", "old-token")

	env := buildClientEnv("new-token")

	var gotPath, gotHome, gotToken string
	tokenCount := 0
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			gotPath = strings.TrimPrefix(entry, "PATH=")
		}
		if strings.HasPrefix(entry, "HOME=") {
			gotHome = strings.TrimPrefix(entry, "HOME=")
		}
		if strings.HasPrefix(entry, "COPILOT_GITHUB_TOKEN=") {
			gotToken = strings.TrimPrefix(entry, "COPILOT_GITHUB_TOKEN=")
			tokenCount++
		}
	}

	if gotPath != "/tmp/test-path" {
		t.Fatalf("expected PATH to be preserved, got %q", gotPath)
	}
	if gotHome != "/tmp/test-home" {
		t.Fatalf("expected HOME to be preserved, got %q", gotHome)
	}
	if gotToken != "new-token" {
		t.Fatalf("expected token to be overridden, got %q", gotToken)
	}
	if tokenCount != 1 {
		t.Fatalf("expected exactly one COPILOT_GITHUB_TOKEN entry, got %d", tokenCount)
	}

	if len(env) < len(os.Environ()) {
		t.Fatalf("expected env to keep base entries, got fewer entries: %d < %d", len(env), len(os.Environ()))
	}
}

func TestHandleChatCompletions_NoAPIKey(t *testing.T) {
	srv := &Server{clients: make(map[string]*copilot.Client)}
	// no default client
	reqBody := `{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}`
	req, _ := http.NewRequest("POST", "/v1/chat/completions", strings.NewReader(reqBody))
	rw := &responseRecorder{head: http.Header{}}
	srv.HandleChatCompletions(rw, req)
	if rw.status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rw.status)
	}
}

func TestStatusFromSessionError(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want int
	}{
		{
			name: "Parse CAPIError 400",
			msg:  "Execution failed... Last error: CAPIError: 400 400 Bad Request",
			want: http.StatusBadRequest,
		},
		{
			name: "Timeout fallback",
			msg:  "upstream timeout waiting for model",
			want: http.StatusGatewayTimeout,
		},
		{
			name: "Generic fallback",
			msg:  "unexpected upstream failure",
			want: http.StatusBadGateway,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusFromSessionError(tt.msg); got != tt.want {
				t.Fatalf("statusFromSessionError() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUserMessageFromSessionError(t *testing.T) {
	msg := "Failed to get response... Last error: CAPIError: 400 400 Bad Request"
	if got := userMessageFromSessionError(msg); got != "CAPIError: 400 400 Bad Request" {
		t.Fatalf("unexpected user message: %q", got)
	}

	if got := userMessageFromSessionError("   "); got != "Upstream model request failed" {
		t.Fatalf("empty message fallback mismatch: %q", got)
	}
}

func TestOpenAIErrorTypeForStatus(t *testing.T) {
	if got := openAIErrorTypeForStatus(http.StatusUnauthorized); got != "authentication_error" {
		t.Fatalf("401 should map to authentication_error, got %q", got)
	}
	if got := openAIErrorTypeForStatus(http.StatusBadRequest); got != "invalid_request_error" {
		t.Fatalf("400 should map to invalid_request_error, got %q", got)
	}
	if got := openAIErrorTypeForStatus(http.StatusBadGateway); got != "api_error" {
		t.Fatalf("502 should map to api_error, got %q", got)
	}
}

// simple response recorder to capture status for testing handlers

type responseRecorder struct {
	head   http.Header
	body   strings.Builder
	status int
}

func (r *responseRecorder) Header() http.Header { return r.head }
func (r *responseRecorder) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *responseRecorder) WriteHeader(code int) { r.status = code }
