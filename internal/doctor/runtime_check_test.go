package doctor

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

// fakeFileInfo is a minimal os.FileInfo for the runtime-check stat injection.
type fakeFileInfo struct{ dir bool }

func (f fakeFileInfo) Name() string       { return "zero" }
func (f fakeFileInfo) Size() int64        { return 1 }
func (f fakeFileInfo) Mode() fs.FileMode  { return 0o755 }
func (f fakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeFileInfo) IsDir() bool        { return f.dir }
func (f fakeFileInfo) Sys() any           { return nil }

func runtimeCheckFor(t *testing.T, exe func() (string, error), stat func(string) (os.FileInfo, error)) Check {
	t.Helper()
	return runtimeCheck(Options{Runtime: "go1.26", Executable: exe, StatExecutable: stat})
}

// A present, runnable binary passes AND reports its resolved path — the check is
// no longer a hardcoded green; it names which binary answered.
func TestRuntimeCheckPassesAndReportsBinaryPath(t *testing.T) {
	c := runtimeCheckFor(t,
		func() (string, error) { return "/opt/homebrew/lib/node_modules/@gitlawb/zero/zero", nil },
		func(string) (os.FileInfo, error) { return fakeFileInfo{}, nil })
	if c.Status != StatusPass {
		t.Fatalf("status = %s, want pass", c.Status)
	}
	if c.Details["binaryPath"] != "/opt/homebrew/lib/node_modules/@gitlawb/zero/zero" {
		t.Fatalf("binaryPath = %v, want the resolved path in details", c.Details["binaryPath"])
	}
	if !strings.Contains(c.Message, "/opt/homebrew/lib/node_modules/@gitlawb/zero/zero") {
		t.Fatalf("message should surface the binary path, got %q", c.Message)
	}
}

// The vanished-binary case FAILS loudly instead of the old unconditional pass —
// this is the #405 regression: a broken runtime no longer gets a green check.
func TestRuntimeCheckFailsWhenBinaryMissing(t *testing.T) {
	c := runtimeCheckFor(t,
		func() (string, error) { return "/path/to/zero", nil },
		func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	if c.Status != StatusFail {
		t.Fatalf("#405: a missing/inaccessible binary must FAIL, got %s (%q)", c.Status, c.Message)
	}
	if !strings.Contains(c.Message, "/path/to/zero") {
		t.Fatalf("failure message should name the path, got %q", c.Message)
	}
}

// os.Executable resolution failure fails loudly too.
func TestRuntimeCheckFailsWhenExecutableUnresolvable(t *testing.T) {
	c := runtimeCheckFor(t,
		func() (string, error) { return "", errors.New("no executable path") },
		func(string) (os.FileInfo, error) { return fakeFileInfo{}, nil })
	if c.Status != StatusFail {
		t.Fatalf("unresolvable executable must FAIL, got %s", c.Status)
	}
}

// A path that resolves to a directory is not a runnable binary → FAIL.
func TestRuntimeCheckFailsWhenPathIsDirectory(t *testing.T) {
	c := runtimeCheckFor(t,
		func() (string, error) { return "/some/dir", nil },
		func(string) (os.FileInfo, error) { return fakeFileInfo{dir: true}, nil })
	if c.Status != StatusFail {
		t.Fatalf("a directory path must FAIL, got %s", c.Status)
	}
}

// Default injection (nil Executable/StatExecutable) resolves the real running
// test binary — proving the production default path works and passes.
func TestRuntimeCheckDefaultsToRealBinary(t *testing.T) {
	c := runtimeCheck(Options{})
	if c.Status != StatusPass {
		t.Fatalf("default runtime check should pass for the running test binary, got %s (%q)", c.Status, c.Message)
	}
	if strings.TrimSpace(c.Details["binaryPath"].(string)) == "" {
		t.Fatal("default runtime check should report the running binary path")
	}
}
