package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gitlawb/zero/internal/config"
	"github.com/Gitlawb/zero/internal/modelregistry"
	"github.com/Gitlawb/zero/internal/zeroruntime"
)

func fastTestRegistry(t *testing.T) modelregistry.Registry {
	t.Helper()
	registry, err := modelregistry.DefaultRegistry()
	if err != nil {
		t.Fatalf("DefaultRegistry() error = %v", err)
	}
	return registry
}

// planFastCommand is the pure decision layer — every argument/state case is
// exercised here without touching the switch machinery.
func TestPlanFastCommand(t *testing.T) {
	registry := fastTestRegistry(t)

	cases := []struct {
		name        string
		model       string
		arg         string
		wantMessage string // substring expected when the plan is terminal
		wantTarget  string // model id expected when the plan switches
		wantToFast  bool
	}{
		{name: "invalid arg", model: "gpt-4.1", arg: "turbo", wantMessage: `/fast takes on or off, not "turbo"`},
		{name: "no model set", model: "", arg: "on", wantMessage: "Pick a model first"},
		{name: "enable while already fast", model: "gpt-4.1-mini", arg: "on", wantMessage: "Already in the fast lane"},
		{name: "disable while already base", model: "gpt-4.1", arg: "off", wantMessage: "Already on base"},
		{name: "enable with no fast variant", model: "claude-opus-4.1", arg: "on", wantMessage: "No fast lane"},
		{name: "enable (no arg) resolves fast", model: "gpt-4.1", arg: "", wantTarget: "gpt-4.1-mini", wantToFast: true},
		{name: "enable (on) resolves fast", model: "gpt-4.1", arg: "ON", wantTarget: "gpt-4.1-mini", wantToFast: true},
		{name: "disable resolves base", model: "gpt-4.1-mini", arg: "off", wantTarget: "gpt-4.1", wantToFast: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := planFastCommand(registry, tc.model, tc.arg)
			if tc.wantTarget != "" {
				canon, ok := registry.Resolve(tc.wantTarget)
				if !ok {
					t.Fatalf("test fixture %q does not resolve", tc.wantTarget)
				}
				if plan.targetID != canon.ID {
					t.Fatalf("targetID = %q, want %q", plan.targetID, canon.ID)
				}
				if plan.toFast != tc.wantToFast {
					t.Fatalf("toFast = %v, want %v", plan.toFast, tc.wantToFast)
				}
				if plan.message != "" {
					t.Fatalf("switching plan should carry no message, got %q", plan.message)
				}
				return
			}
			if plan.targetID != "" {
				t.Fatalf("expected a terminal message, got target %q", plan.targetID)
			}
			if !strings.Contains(plan.message, tc.wantMessage) {
				t.Fatalf("message = %q, want it to contain %q", plan.message, tc.wantMessage)
			}
		})
	}
}

// handleFastCommand's terminal cases return the plan message without switching,
// so a bare model exercises them.
func TestHandleFastCommandTerminalCases(t *testing.T) {
	cases := []struct {
		name         string
		model        string
		arg          string
		wantContains string
	}{
		{"invalid arg", "gpt-4.1", "sideways", `/fast takes on or off, not "sideways"`},
		{"no model", "", "on", "Pick a model first"},
		{"already fast", "gpt-4.1-mini", "on", "Already in the fast lane"},
		{"already base", "gpt-4.1", "off", "Already on base"},
		{"no fast variant", "claude-opus-4.1", "on", "No fast lane"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := model{modelName: tc.model}
			got, notice := m.handleFastCommand(tc.arg)
			if got.modelName != tc.model {
				t.Fatalf("model must not change on a terminal case: %q -> %q", tc.model, got.modelName)
			}
			if !strings.Contains(notice, tc.wantContains) {
				t.Fatalf("notice = %q, want it to contain %q", notice, tc.wantContains)
			}
		})
	}
}

func newFastTestModel(t *testing.T, modelID string) model {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "zero.json")
	profile := config.ProviderProfile{
		Name:         "openai",
		ProviderKind: config.ProviderKindOpenAI,
		BaseURL:      config.OpenAIBaseURL,
		APIKey:       "sk-test",
		Model:        modelID,
	}
	if _, err := config.UpsertProvider(configPath, profile, true); err != nil {
		t.Fatalf("write user config: %v", err)
	}
	return newModel(context.Background(), Options{
		UserConfigPath:  configPath,
		ProviderName:    "openai",
		ModelName:       modelID,
		ProviderProfile: profile,
		Provider:        &fakeProvider{},
		NewProvider: func(config.ProviderProfile) (zeroruntime.Provider, error) {
			return &fakeProvider{}, nil
		},
	})
}

// Enabling switches base -> fast with a ⚡ notice; disabling switches back to the
// base with a plain notice. Uses the same switch path /model uses.
func TestHandleFastCommandSwitchesToFastAndBack(t *testing.T) {
	registry := fastTestRegistry(t)
	m := newFastTestModel(t, "gpt-4.1")

	m, notice := m.handleFastCommand("on")
	if m.modelName != "gpt-4.1-mini" {
		t.Fatalf("enable: modelName = %q, want gpt-4.1-mini", m.modelName)
	}
	if want := "⚡ Fast lane on · " + modelShortName(registry, "gpt-4.1-mini"); notice != want {
		t.Fatalf("enable notice = %q, want exactly %q (bolt + fast-lane-on + target short name)", notice, want)
	}

	m, notice = m.handleFastCommand("off")
	if m.modelName != "gpt-4.1" {
		t.Fatalf("disable: modelName = %q, want gpt-4.1", m.modelName)
	}
	if want := "Fast lane off · " + modelShortName(registry, "gpt-4.1"); notice != want {
		t.Fatalf("disable notice = %q, want exactly %q (plain fast-lane-off, no bolt)", notice, want)
	}
}

// The no-op and no-variant notices must carry their full spec decoration — the ⚡
// prefix and the (<short>) name — so a regression dropping either fails loudly
// rather than passing a bare substring assertion.
func TestFastCommandNoOpAndNoVariantMessagesAreDecorated(t *testing.T) {
	registry := fastTestRegistry(t)

	plan := planFastCommand(registry, "gpt-4.1-mini", "on")
	if want := "⚡ Already in the fast lane · " + modelShortName(registry, "gpt-4.1-mini"); plan.message != want {
		t.Fatalf("already-fast message = %q, want %q", plan.message, want)
	}

	plan = planFastCommand(registry, "gpt-4.1", "off")
	if want := "Already on base · " + modelShortName(registry, "gpt-4.1"); plan.message != want {
		t.Fatalf("already-base message = %q, want %q", plan.message, want)
	}

	// An unknown / off-catalog model id also reaches the no-variant branch, where
	// <short> falls back to the raw id.
	plan = planFastCommand(registry, "some-custom-live-provider-model", "on")
	if want := "⚡ No fast lane for some-custom-live-provider-model"; plan.message != want {
		t.Fatalf("no-variant (unknown id) message = %q, want %q", plan.message, want)
	}
}

// When the underlying switch fails, /fast surfaces the error and leaves the
// model unchanged (here: a valid base model but no provider profile to rebuild).
func TestHandleFastCommandSurfacesSwitchError(t *testing.T) {
	m := model{modelName: "gpt-4.1"} // no providerProfile / newProvider wired
	got, notice := m.handleFastCommand("on")
	if got.modelName != "gpt-4.1" {
		t.Fatalf("model must be unchanged on switch failure, got %q", got.modelName)
	}
	if strings.HasPrefix(notice, "⚡ Fast lane on ") || strings.HasPrefix(notice, "Fast lane off ") {
		t.Fatalf("a failed switch must not claim success, got %q", notice)
	}
	if strings.TrimSpace(notice) == "" {
		t.Fatal("a failed switch must surface a reason")
	}
}

func TestModelShortName(t *testing.T) {
	registry := fastTestRegistry(t)
	if got := modelShortName(registry, "gpt-4.1"); strings.TrimSpace(got) == "" {
		t.Fatal("a known model should resolve to a non-empty short name")
	}
	if got := modelShortName(registry, "totally-unknown-model-xyz"); got != "totally-unknown-model-xyz" {
		t.Fatalf("an unknown model short name = %q, want the raw id", got)
	}
}
