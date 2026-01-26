package main

import (
	"strings"
	"testing"
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
