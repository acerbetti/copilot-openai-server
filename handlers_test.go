package main

import (
	"net/http"
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

// simple response recorder to capture status for testing handlers

type responseRecorder struct {
	head   http.Header
	body   strings.Builder
	status int
}

func (r *responseRecorder) Header() http.Header { return r.head }
func (r *responseRecorder) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *responseRecorder) WriteHeader(code int) { r.status = code }
