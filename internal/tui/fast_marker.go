package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Gitlawb/zero/internal/modelregistry"
)

// fastMarkerGlyph is the compact "f" badge shown next to a model that has an
// available fast variant — i.e. one where `/fast on` would switch down to a
// faster same-family model. It marks the BASE model, never the fast one.
const fastMarkerGlyph = "f"

// fastVariantEntryFor resolves the fast variant of an arbitrary model id, if any.
// Like modelContextWindow, it reads the default registry on demand rather than
// caching derived state, so a marker can never go stale after a model switch.
// DefaultRegistry is memoized, so even the per-frame callers (the active-model
// chip renders every frame; the pickers render per keystroke) pay only a map
// lookup here, not a registry rebuild.
func fastVariantEntryFor(id string) (modelregistry.ModelEntry, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return modelregistry.ModelEntry{}, false
	}
	registry, err := modelregistry.DefaultRegistry()
	if err != nil {
		return modelregistry.ModelEntry{}, false
	}
	return registry.FastVariant(id)
}

// currentModelFastVariant resolves the fast variant of the session's active model.
func (m model) currentModelFastVariant() (modelregistry.ModelEntry, bool) {
	return fastVariantEntryFor(m.modelName)
}

// renderFastBadge renders the "f" fast-mode marker in the given (surface-aware)
// style, prefixed with a space so it reads as a distinct badge rather than a
// trailing letter of the adjacent model name.
func renderFastBadge(style lipgloss.Style) string {
	return " " + style.Render(fastMarkerGlyph)
}

// fastVariantHint is the one-line reveal shown against a focused/selected
// fast-capable model. The TUI has no floating-tooltip primitive, so the "hover"
// affordance is this inline detail on the active row: it names the target model
// and the command that gets there.
func fastVariantHint(variantName string) string {
	variantName = strings.TrimSpace(variantName)
	if variantName == "" {
		return ""
	}
	return "⚡ fast: " + variantName + " · /fast on"
}

// fastToastDuration is how long a /fast toast stays up before auto-dismissing.
const fastToastDuration = 3 * time.Second

// fastToastExpiredMsg clears a fast-lane toast after fastToastDuration. It is
// seq-gated so a newer toast is never dismissed by an older timer.
type fastToastExpiredMsg struct{ seq int }

// showFastToast raises a transient fast-lane notice in the row above the composer
// and schedules its dismissal. It replaces the persistent transcript line the
// /fast command used to append, so repeated toggles never pile up in history.
func (m model) showFastToast(text string) (model, tea.Cmd) {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if text == "" {
		return m, nil
	}
	m.fastToastSeq++
	m.fastToast = text
	seq := m.fastToastSeq
	return m, tea.Tick(fastToastDuration, func(time.Time) tea.Msg {
		return fastToastExpiredMsg{seq: seq}
	})
}

// renderFastToast styles the toast as a filled brand pill so it reads as a
// transient popup rather than a permanent status line.
func renderFastToast(text string) string {
	return zeroTheme.badge.Render(" " + strings.TrimSpace(text) + " ")
}
