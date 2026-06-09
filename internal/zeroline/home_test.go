package zeroline

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHomeWordmarkTaglineHintChips(t *testing.T) {
	d := HomeData{
		Variant: 0, Dark: true, Width: 90, Height: 30,
		Header:    Header{Model: "claude-sonnet-4-5"},
		Chips:     []string{"Add a --version flag", "Why is go vet failing?", "Create hello.txt"},
		ChipIndex: 1,
	}
	out := RenderHome(d)
	if h := lipgloss.Height(out); h != 30 {
		t.Fatalf("home height = %d, want 30 (frame-exact)", h)
	}
	for _, line := range strings.Split(out, "\n") {
		if lipgloss.Width(line) > 90 {
			t.Fatalf("home line exceeds width 90: %d (%q)", lipgloss.Width(line), stripANSI(line))
		}
	}
	plain := stripANSI(out)
	for _, want := range []string{
		"std-lib-first", "running", "zero", "against", "claude-sonnet-4-5",
		"Add a --version flag", "Why is go vet failing?", "Create hello.txt", "❯",
	} {
		if !strings.Contains(plain, want) {
			t.Errorf("home missing %q", want)
		}
	}
}

// Chip selection/border behavior is covered by TestChipBoxBorderedAndSelected.
