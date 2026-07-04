package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/Gitlawb/zero/internal/modelregistry"
)

// hasFastBadge reports whether a rendered (styled) line ends with the " f"
// fast-mode badge, ignoring trailing layout padding. Labels used in these tests
// deliberately contain no trailing " f" so the badge is unambiguous.
func hasFastBadge(rendered string) bool {
	return strings.HasSuffix(strings.TrimRight(ansi.Strip(rendered), " "), " "+fastMarkerGlyph)
}

func TestFastVariantHint(t *testing.T) {
	if got := fastVariantHint("Claude Haiku 4.5"); got != "⚡ fast: Claude Haiku 4.5 · /fast on" {
		t.Errorf("fastVariantHint = %q", got)
	}
	if got := fastVariantHint("   "); got != "" {
		t.Errorf("fastVariantHint(blank) = %q, want empty", got)
	}
}

func TestRenderFastBadge(t *testing.T) {
	if got := ansi.Strip(renderFastBadge(zeroTheme.accent)); got != " "+fastMarkerGlyph {
		t.Errorf("renderFastBadge visible text = %q, want %q", got, " "+fastMarkerGlyph)
	}
}

// The active-model chip carries the "f" badge only when the current model has a
// fast variant — base models yes, fast models and unpaired top-tier models no.
func TestTitleModelSegmentFastBadge(t *testing.T) {
	cases := []struct {
		model string
		want  bool
	}{
		{"gpt-4.1", true},          // base with a fast variant
		{"gpt-4.1-mini", false},    // the fast variant itself has none
		{"claude-opus-4.1", false}, // top-tier, intentionally unpaired
		{"definitely-not-real", false},
		{"", false},
	}
	for _, tc := range cases {
		seg := model{modelName: tc.model}.titleModelSegment()
		if got := hasFastBadge(seg); got != tc.want {
			t.Errorf("titleModelSegment(%q) badge = %v, want %v (visible %q)", tc.model, got, tc.want, ansi.Strip(seg))
		}
	}
}

// A /model picker row shows the badge iff its FastVariant field is populated,
// and the field is populated only for base models by decorateFastVariantItems.
func TestRenderModelPickerRowFastBadge(t *testing.T) {
	withFast := renderModelPickerRow(60, false, pickerItem{Label: "gpt-4.1", Value: "gpt-4.1", FastVariant: "GPT-4.1 mini"})
	if !hasFastBadge(withFast) {
		t.Errorf("fast-capable row missing badge: %q", ansi.Strip(withFast))
	}
	withoutFast := renderModelPickerRow(60, false, pickerItem{Label: "gpt-4.1", Value: "gpt-4.1"})
	if hasFastBadge(withoutFast) {
		t.Errorf("row without FastVariant should have no badge: %q", ansi.Strip(withoutFast))
	}
}

func TestModelPickerItemDetailIncludesFastHint(t *testing.T) {
	detail := modelPickerItemDetail(pickerItem{Label: "GPT-4.1", Value: "gpt-4.1", FastVariant: "GPT-4.1 mini"})
	if !strings.Contains(detail, "⚡ fast: GPT-4.1 mini · /fast on") {
		t.Errorf("detail missing fast hint: %q", detail)
	}
	plain := modelPickerItemDetail(pickerItem{Label: "Claude Opus 4.1", Value: "claude-opus-4.1"})
	if strings.Contains(plain, "fast:") {
		t.Errorf("non-fast detail should have no hint: %q", plain)
	}
}

// decorateFastVariantItems stamps the fast variant's display name onto base-model
// rows only, using the real registry — the single source of truth for the badge.
func TestDecorateFastVariantItems(t *testing.T) {
	registry, err := modelregistry.DefaultRegistry()
	if err != nil {
		t.Fatalf("DefaultRegistry: %v", err)
	}
	items := []pickerItem{
		{Value: "gpt-4.1"},           // base -> GPT-4.1 mini
		{Value: "gpt-4.1-mini"},      // fast, no variant
		{Value: "claude-opus-4.1"},   // unpaired
		{Value: "some-custom-model"}, // not in registry
		{Value: ""},
	}
	decorateFastVariantItems(registry, items)
	if items[0].FastVariant != "GPT-4.1 mini" {
		t.Errorf("gpt-4.1 FastVariant = %q, want %q", items[0].FastVariant, "GPT-4.1 mini")
	}
	for _, i := range []int{1, 2, 3, 4} {
		if items[i].FastVariant != "" {
			t.Errorf("items[%d].FastVariant = %q, want empty", i, items[i].FastVariant)
		}
	}
}

func TestShowFastToast(t *testing.T) {
	// A non-empty message raises the toast, bumps the seq, and schedules dismissal.
	m, cmd := model{}.showFastToast("⚡ Fast lane on · GPT-4.1 mini")
	if m.fastToast != "⚡ Fast lane on · GPT-4.1 mini" {
		t.Fatalf("fastToast = %q", m.fastToast)
	}
	if m.fastToastSeq != 1 {
		t.Fatalf("fastToastSeq = %d, want 1", m.fastToastSeq)
	}
	if cmd == nil {
		t.Fatal("showFastToast must return a dismissal tick cmd")
	}

	// A multi-line switch-guard message is flattened so the toast stays one line
	// (the footer slot is fixed-height).
	m2, _ := model{}.showFastToast("Model\nCannot switch while a run is active.")
	if strings.Contains(m2.fastToast, "\n") {
		t.Fatalf("toast must be flattened to one line, got %q", m2.fastToast)
	}

	// Blank text is a no-op: no toast, no timer.
	m3, cmd3 := model{}.showFastToast("   ")
	if m3.fastToast != "" || cmd3 != nil {
		t.Fatalf("blank text should not raise a toast, got %q / cmd!=nil=%v", m3.fastToast, cmd3 != nil)
	}
}

func TestRenderFastToastShowsMessage(t *testing.T) {
	got := ansi.Strip(renderFastToast("⚡ No fast lane for gpt-5.5"))
	if !strings.Contains(got, "⚡ No fast lane for gpt-5.5") {
		t.Fatalf("rendered toast = %q, want it to contain the message", got)
	}
}

func TestProviderWizardModelFastBadgeAndHint(t *testing.T) {
	wizard := &providerWizardState{selectedModel: 0}
	withFast := wizard.renderSelectableModel(60, 0, providerWizardModel{ID: "gpt-4.1"})
	if !hasFastBadge(withFast) {
		t.Errorf("provider wizard base-model row missing badge: %q", ansi.Strip(withFast))
	}
	plain := wizard.renderSelectableModel(60, 1, providerWizardModel{ID: "claude-opus-4.1"})
	if hasFastBadge(plain) {
		t.Errorf("provider wizard unpaired-model row should have no badge: %q", ansi.Strip(plain))
	}
	if detail := providerWizardModelDetail(providerWizardModel{ID: "gpt-4.1"}); !strings.Contains(detail, "⚡ fast:") {
		t.Errorf("provider wizard detail missing fast hint: %q", detail)
	}
}
