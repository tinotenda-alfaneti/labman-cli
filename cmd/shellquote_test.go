package cmd

import "testing"

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "''",
		},
		{
			name:  "simple string",
			input: "hello",
			want:  "'hello'",
		},
		{
			name:  "string with spaces",
			input: "hello world",
			want:  "'hello world'",
		},
		{
			name:  "string with single quote",
			input: "it's",
			want:  "'it'\\''s'",
		},
		{
			name:  "string with multiple single quotes",
			input: "can't won't don't",
			want:  "'can'\\''t won'\\''t don'\\''t'",
		},
		{
			name:  "string with special characters",
			input: "hello$world",
			want:  "'hello$world'",
		},
		{
			name:  "string with backticks",
			input: "hello`world`",
			want:  "'hello`world`'",
		},
		{
			name:  "string with semicolons",
			input: "hello; rm -rf /",
			want:  "'hello; rm -rf /'",
		},
		{
			name:  "string with pipes",
			input: "hello | cat",
			want:  "'hello | cat'",
		},
		{
			name:  "string with newlines",
			input: "hello\nworld",
			want:  "'hello\nworld'",
		},
		{
			name:  "only single quotes",
			input: "'''",
			want:  "''\\'''\\'''\\'''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shellQuote(tt.input)
			if got != tt.want {
				t.Errorf("shellQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestJoinCommand(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "empty parts",
			parts: []string{},
			want:  "",
		},
		{
			name:  "single part",
			parts: []string{"echo"},
			want:  "echo",
		},
		{
			name:  "multiple parts",
			parts: []string{"echo", "hello", "world"},
			want:  "echo hello world",
		},
		{
			name:  "parts with whitespace",
			parts: []string{"  echo  ", "  hello  ", "  world  "},
			want:  "echo hello world",
		},
		{
			name:  "parts with empty strings",
			parts: []string{"echo", "", "hello", "", "world"},
			want:  "echo hello world",
		},
		{
			name:  "parts with only spaces",
			parts: []string{"echo", "   ", "hello"},
			want:  "echo hello",
		},
		{
			name:  "all empty strings",
			parts: []string{"", "", ""},
			want:  "",
		},
		{
			name:  "all whitespace",
			parts: []string{"  ", "   ", "    "},
			want:  "",
		},
		{
			name:  "mixed content",
			parts: []string{"ls", "-la", "", "/home/user", "  ", "documents"},
			want:  "ls -la /home/user documents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinCommand(tt.parts...)
			if got != tt.want {
				t.Errorf("joinCommand(%v) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}
