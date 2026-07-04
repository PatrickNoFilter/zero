package tui

import "testing"

func TestSessionMatchesWorkspace(t *testing.T) {
	cases := []struct {
		name        string
		sessionCwd  string
		workspace   string
		wantVisible bool
	}{
		{"same workspace", "/home/u/proj", "/home/u/proj", true},
		{"trailing slash normalizes", "/home/u/proj/", "/home/u/proj", true},
		{"different workspace hidden", "/home/u/other", "/home/u/proj", false},
		{"session with no cwd stays visible", "", "/home/u/proj", true},
		{"unknown current workspace keeps all", "/home/u/other", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sessionMatchesWorkspace(tc.sessionCwd, tc.workspace); got != tc.wantVisible {
				t.Fatalf("sessionMatchesWorkspace(%q, %q) = %v, want %v", tc.sessionCwd, tc.workspace, got, tc.wantVisible)
			}
		})
	}
}
