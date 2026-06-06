package tools

import (
	"reflect"
	"testing"
)

func TestMutationTargets(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		tool string
		args map[string]any
		want []string
	}{
		{"write_file", map[string]any{"path": "a/b.txt", "content": "x"}, []string{"a/b.txt"}},
		{"edit_file", map[string]any{"path": "c.txt", "old_string": "x", "new_string": "y"}, []string{"c.txt"}},
		{"apply_patch", map[string]any{"patch": "--- a/d.txt\n+++ b/d.txt\n@@ -1 +1 @@\n-x\n+y\n"}, []string{"d.txt"}},
		{"bash", map[string]any{"command": "echo hi"}, nil},
		{"read_file", map[string]any{"path": "e.txt"}, nil},
		{"grep", map[string]any{"pattern": "x"}, nil},
	}
	for _, tc := range cases {
		got := MutationTargets(root, tc.tool, tc.args)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.tool, got, tc.want)
		}
	}
}

func TestMutationTargetsRejectsEscapingPaths(t *testing.T) {
	root := t.TempDir()
	if got := MutationTargets(root, "write_file", map[string]any{"path": "../escape.txt", "content": "x"}); len(got) != 0 {
		t.Errorf("expected no targets for escaping path, got %v", got)
	}
}

func TestStripPatchPrefixStripsOnlyOne(t *testing.T) {
	root := t.TempDir()
	// A workspace file under a directory literally named "b".
	got := MutationTargets(root, "apply_patch", map[string]any{
		"patch": "--- a/b/foo.txt\n+++ b/b/foo.txt\n@@ -1 +1 @@\n-x\n+y\n",
	})
	if len(got) != 1 || got[0] != "b/foo.txt" {
		t.Fatalf("expected [b/foo.txt], got %v", got)
	}
}
