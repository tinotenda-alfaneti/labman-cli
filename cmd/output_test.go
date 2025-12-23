package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPrintBanner(t *testing.T) {
	var buf bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buf)

	printBanner(cmd)

	output := buf.String()
	if !strings.Contains(output, "LabMan") {
		t.Errorf("printBanner() output should contain 'LabMan', got %q", output)
	}
}

func TestPrintSection(t *testing.T) {
	tests := []struct {
		name  string
		title string
		body  string
		want  []string
	}{
		{
			name:  "simple section",
			title: "TEST",
			body:  "Hello World",
			want:  []string{"TEST", "Hello World", "+", "|", "-"},
		},
		{
			name:  "multiline body",
			title: "MULTI",
			body:  "Line 1\nLine 2\nLine 3",
			want:  []string{"MULTI", "Line 1", "Line 2", "Line 3"},
		},
		{
			name:  "empty body",
			title: "EMPTY",
			body:  "",
			want:  []string{"EMPTY", "+", "|"},
		},
		{
			name:  "body with trailing newlines",
			title: "TRAILING",
			body:  "Content\n\n\n",
			want:  []string{"TRAILING", "Content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)

			printSection(cmd, tt.title, tt.body)

			output := buf.String()
			for _, expected := range tt.want {
				if !strings.Contains(output, expected) {
					t.Errorf("printSection() output missing %q\nGot: %s", expected, output)
				}
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single line",
			input: "hello",
			want:  []string{"hello"},
		},
		{
			name:  "multiple lines",
			input: "line1\nline2\nline3",
			want:  []string{"line1", "line2", "line3"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{""},
		},
		{
			name:  "trailing newline",
			input: "hello\n",
			want:  []string{"hello"},
		},
		{
			name:  "multiple trailing newlines",
			input: "hello\n\n\n",
			want:  []string{"hello"},
		},
		{
			name:  "only newlines",
			input: "\n\n\n",
			want:  []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitLines() got %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitLines() line %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
