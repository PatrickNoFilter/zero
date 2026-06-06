package zenline

import (
	"strings"
	"testing"
)

func TestThemesCount(t *testing.T) {
	if len(Themes) != 5 {
		t.Fatalf("expected 5 themes, got %d", len(Themes))
	}
	for i, th := range Themes {
		if th.Name == "" || th.Dark.Bg == "" || th.Light.Bg == "" {
			t.Errorf("theme %d (%q) incomplete", i, th.Name)
		}
	}
}

func TestRenderHomeAllThemes(t *testing.T) {
	for v := 0; v < len(Themes); v++ {
		out := RenderHome(HomeData{
			Variant: v, Dark: true, Width: 100, Height: 28,
			Header: Header{Cwd: "~/src/zero", Branch: "main", Model: "claude-sonnet-4.5", Provider: "anthropic"},
			Input:  "❯ message zero",
		})
		if !strings.Contains(out, "Own your agent") {
			t.Errorf("theme %d: home missing tagline", v)
		}
	}
}

func TestRenderChatLiveData(t *testing.T) {
	d := ChatData{
		Variant: 1, Dark: true, Width: 100, Height: 30,
		Header: Header{Cwd: "~/src/zero", Branch: "main", Model: "claude-sonnet-4.5", Provider: "anthropic"},
		Rows: []Row{
			{Kind: "user", Text: "refactor the loop"},
			{Kind: "toolcall", Tool: "grep", Detail: "pattern: case"},
			{Kind: "toolresult", Tool: "grep", Status: "ok", Detail: "3 matches"},
			{Kind: "assistant", Text: "Here is the plan."},
		},
	}
	out := RenderChat(d)
	for _, want := range []string{"DONE", "you", "grep", "✦ zero", "claude-sonnet-4.5"} {
		if !strings.Contains(out, want) {
			t.Errorf("chat render missing %q", want)
		}
	}

	// thinking state shows the WORKING mode + an animated thinking line
	d.Rows = d.Rows[:1]
	d.Working = true
	d.Thinking = true
	if w := RenderChat(d); !strings.Contains(w, "WORKING") || !strings.Contains(w, "thinking") {
		t.Error("thinking state not rendered")
	}
	// streaming state shows the live assistant text
	d.Thinking = false
	d.Stream = "streaming-response-here"
	if w := RenderChat(d); !strings.Contains(w, "streaming-response-here") {
		t.Error("streaming text not rendered")
	}

	// permission modal shows the gate choices and BLOCKED mode
	d.Working = false
	d.Perm = &Perm{Tool: "edit_file", Risk: "medium", Reason: "writes a file"}
	p := RenderChat(d)
	for _, want := range []string{"BLOCKED", "permission required", "edit_file", "allow", "always", "deny"} {
		if !strings.Contains(p, want) {
			t.Errorf("permission modal missing %q", want)
		}
	}
}

func TestPermLayoutMatchesRender(t *testing.T) {
	// The buttons row in the rendered modal must sit exactly where PermLayout
	// says, so mouse clicks land on the right choice.
	w, h := 90, 24
	g := PermLayout(w, h)
	out := RenderChat(ChatData{
		Variant: 0, Dark: true, Width: w, Height: h,
		Perm: &Perm{Tool: "edit_file", Risk: "medium", Reason: "writes a file"},
	})
	lines := strings.Split(out, "\n")
	if g.Allow.Y >= len(lines) {
		t.Fatalf("allow row %d beyond frame height %d", g.Allow.Y, len(lines))
	}
	row := lines[g.Allow.Y]
	if !strings.Contains(row, "allow") || !strings.Contains(row, "deny") {
		t.Errorf("button row %d does not contain the buttons: %q", g.Allow.Y, stripANSI(row))
	}
	// hit-test sanity: a click in the middle of each button resolves correctly
	mid := func(r Rect) (int, int) { return r.X + r.W/2, r.Y }
	for name, r := range map[string]Rect{"allow": g.Allow, "always": g.Always, "deny": g.Deny} {
		x, y := mid(r)
		if got := g.Hit(x, y); got != name {
			t.Errorf("Hit(%d,%d) = %q, want %q", x, y, got, name)
		}
	}
	if got := g.Hit(0, 0); got != "" {
		t.Errorf("Hit(0,0) = %q, want empty", got)
	}
}

func TestToolResultRenderingCollapsesAndShows(t *testing.T) {
	d := ChatData{
		Variant: 0, Dark: true, Width: 100, Height: 40,
		Rows: []Row{
			{Kind: "toolcall", Tool: "read_file", Detail: "README.md"},
			{Kind: "toolresult", Tool: "read_file", Status: "ok", Detail: "File: README.md (217 lines)\n\n  1 | THE-RAW-FILE-CONTENT-SHOULD-NOT-APPEAR"},
			{Kind: "toolcall", Tool: "list_directory", Detail: "."},
			{Kind: "toolresult", Tool: "list_directory", Status: "ok", Detail: "Contents of .:\n\na\nb\nc"},
			{Kind: "toolcall", Tool: "edit_file", Detail: "x.go"},
			{Kind: "toolresult", Tool: "edit_file", Status: "ok", Detail: "@@ -1 +1 @@\n-old\n+NEWCODE"},
			{Kind: "toolcall", Tool: "bash", Detail: "go test"},
			{Kind: "toolresult", Tool: "bash", Status: "error", Detail: "exit 1: BUILD-FAILED-HERE"},
		},
	}
	out := stripANSI(RenderChat(d))
	// file read collapses to a count, never dumps content
	if !strings.Contains(out, "217 lines") {
		t.Error("read_file should summarize to a line count")
	}
	if strings.Contains(out, "THE-RAW-FILE-CONTENT-SHOULD-NOT-APPEAR") {
		t.Error("read_file dumped raw file content (should be collapsed)")
	}
	// listing collapses to entry count
	if !strings.Contains(out, "3 entries") {
		t.Error("list_directory should summarize to an entry count")
	}
	// diff body is shown for edits
	if !strings.Contains(out, "NEWCODE") {
		t.Error("edit_file diff body should be shown")
	}
	// errors are surfaced
	if !strings.Contains(out, "BUILD-FAILED-HERE") {
		t.Error("error output should be shown")
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && (r == 'm'):
			inEsc = false
		case inEsc:
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
