package zeroline

import (
	"strings"
	"testing"
)

func TestLooksLikeDiff(t *testing.T) {
	cases := map[string]bool{
		"@@ -1 +1 @@\n-a\n+b":       true,
		"--- a\n+++ b\n-x\n+y":      true,
		"+just an addition":         true,
		"plain prose with no diff":  false,
		"wrote 12 lines to file.go": false,
	}
	for in, want := range cases {
		if got := looksLikeDiff(in); got != want {
			t.Errorf("looksLikeDiff(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestAssistantRowRendersPlainMuted(t *testing.T) {
	// .blk-say is plain MUTED prose (no markdown formatting, no panel): markers are
	// preserved verbatim rather than rendered through glamour.
	d := ChatData{
		Variant: 0, Dark: true, Width: 100, Height: 40,
		Rows: []Row{
			{Kind: "assistant", Text: "Use **bold** and a list:\n\n- one\n- two"},
		},
	}
	out := stripANSI(RenderChat(d))
	if !strings.Contains(out, "**bold**") {
		t.Errorf("assistant prose should render plainly (keep markers): %q", out)
	}
	if !strings.Contains(out, "- one") || !strings.Contains(out, "- two") {
		t.Errorf("assistant list lines missing: %q", out)
	}
}

func TestStreamingAssistantStaysPlain(t *testing.T) {
	d := ChatData{
		Variant: 0, Dark: true, Width: 100, Height: 40,
		Stream: "partial **incomplete",
	}
	out := stripANSI(RenderChat(d))
	// Streaming text is shown verbatim (not run through glamour): the raw markdown
	// markers must survive. Glamour would strip/render "**" into a bold span, so
	// asserting the literal "**incomplete" is preserved proves the verbatim path.
	if !strings.Contains(out, "partial **incomplete") {
		t.Errorf("streaming text not verbatim (markdown markers stripped?): %q", out)
	}
}
