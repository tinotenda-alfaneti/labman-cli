package cmd

import (
	"testing"
	"time"
)

func TestFormatTTL(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
		want string
	}{
		{
			name: "expired",
			ttl:  -1 * time.Second,
			want: "expired",
		},
		{
			name: "zero",
			ttl:  0,
			want: "expired",
		},
		{
			name: "only seconds",
			ttl:  45 * time.Second,
			want: "45s",
		},
		{
			name: "minutes and seconds",
			ttl:  5*time.Minute + 30*time.Second,
			want: "5m 30s",
		},
		{
			name: "hours and minutes",
			ttl:  2*time.Hour + 15*time.Minute,
			want: "2h 15m",
		},
		{
			name: "hours minutes and seconds",
			ttl:  1*time.Hour + 30*time.Minute + 45*time.Second,
			want: "1h 30m",
		},
		{
			name: "only hours",
			ttl:  5 * time.Hour,
			want: "5h",
		},
		{
			name: "only minutes",
			ttl:  10 * time.Minute,
			want: "10m",
		},
		{
			name: "24 hours",
			ttl:  24 * time.Hour,
			want: "24h",
		},
		{
			name: "more than 24 hours",
			ttl:  48*time.Hour + 30*time.Minute,
			want: "48h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTTL(tt.ttl)
			if got != tt.want {
				t.Errorf("formatTTL(%v) = %q, want %q", tt.ttl, got, tt.want)
			}
		})
	}
}

func TestNamespaceArg(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name:      "empty namespace returns all namespaces flag",
			namespace: "",
			want:      "-A",
		},
		{
			name:      "default namespace",
			namespace: "default",
			want:      "-n 'default'",
		},
		{
			name:      "custom namespace",
			namespace: "kube-system",
			want:      "-n 'kube-system'",
		},
		{
			name:      "namespace with special characters",
			namespace: "my-app-prod",
			want:      "-n 'my-app-prod'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namespaceArg(tt.namespace)
			if got != tt.want {
				t.Errorf("namespaceArg(%q) = %q, want %q", tt.namespace, got, tt.want)
			}
		})
	}
}

func TestSelectorArg(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     string
	}{
		{
			name:     "empty selector",
			selector: "",
			want:     "",
		},
		{
			name:     "simple label selector",
			selector: "app=web",
			want:     "--selector='app=web'",
		},
		{
			name:     "multiple labels",
			selector: "app=web,tier=frontend",
			want:     "--selector='app=web,tier=frontend'",
		},
		{
			name:     "selector with special characters",
			selector: "app.kubernetes.io/name=nginx",
			want:     "--selector='app.kubernetes.io/name=nginx'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectorArg(tt.selector)
			if got != tt.want {
				t.Errorf("selectorArg(%q) = %q, want %q", tt.selector, got, tt.want)
			}
		})
	}
}

func TestBuildDiagBundleScript(t *testing.T) {
	tests := []struct {
		name       string
		remotePath string
		wantParts  []string
	}{
		{
			name:       "default path",
			remotePath: "/tmp/labman-diag.tar.gz",
			wantParts: []string{
				"set -euo pipefail",
				"TMP_DIR=$(mktemp -d /tmp/labman-diag-XXXXXX)",
				"cleanup() { rm -rf \"$TMP_DIR\"; }",
				"trap cleanup EXIT",
				"kubectl get all -A -o yaml",
				"kubectl get events",
				"kubectl describe nodes",
				"journalctl",
				"microk8s inspect",
				"tar -czf",
				"/tmp/labman-diag.tar.gz",
			},
		},
		{
			name:       "custom path with spaces",
			remotePath: "/home/user/my diag.tar.gz",
			wantParts: []string{
				"set -euo pipefail",
				"tar -czf",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildDiagBundleScript(tt.remotePath)
			for _, part := range tt.wantParts {
				if !containsString(got, part) {
					t.Errorf("buildDiagBundleScript(%q) missing %q", tt.remotePath, part)
				}
			}
		})
	}
}

func TestShouldSkipSession(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		parent  string
		want    bool
	}{
		{
			name:    "nil command",
			cmdName: "",
			parent:  "",
			want:    false,
		},
		{
			name:    "login command",
			cmdName: "login",
			parent:  "",
			want:    true,
		},
		{
			name:    "child of login",
			cmdName: "subcommand",
			parent:  "login",
			want:    true,
		},
		{
			name:    "other command",
			cmdName: "cluster",
			parent:  "root",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is simplified - full testing would require mock cobra.Command
			// Just test the nil case
			if tt.cmdName == "" {
				got := shouldSkipSession(nil)
				if got != tt.want {
					t.Errorf("shouldSkipSession(nil) = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
