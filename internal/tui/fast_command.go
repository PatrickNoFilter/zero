package tui

import (
	"fmt"
	"strings"

	"github.com/Gitlawb/zero/internal/modelregistry"
)

// fastPlan is the pure decision `/fast` makes before touching any state: either
// a terminal message to show the user, or a target model to switch to.
type fastPlan struct {
	message  string // terminal notice; shown as-is when targetID is empty
	targetID string // model to switch to ("" means no switch — show message)
	toFast   bool   // the target is a fast model (drives the ⚡ prefix)
}

// planFastCommand resolves what `/fast [on|off]` should do for the current model
// using only the registry — no side effects. State is DERIVED from the model
// identity: the session is "in fast mode" iff the current model has a base
// variant. Returns a terminal message, or a target model to switch to.
func planFastCommand(registry modelregistry.Registry, currentModel, arg string) fastPlan {
	// 1. Parse the argument (case-insensitive): no arg / "on" => enable fast,
	//    "off" => return to base, anything else => usage error.
	enable := true
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "", "on":
		enable = true
	case "off":
		enable = false
	default:
		return fastPlan{message: fmt.Sprintf("Invalid argument %q. Usage: /fast, /fast on, or /fast off", strings.TrimSpace(arg))}
	}

	// 2. Read the current model.
	current := strings.TrimSpace(currentModel)
	if current == "" {
		return fastPlan{message: "No model is currently set for this session"}
	}
	short := modelShortName(registry, current)

	// 3. Derived state: in fast mode iff the current model has a base variant.
	_, isFast := registry.BaseVariant(current)

	// 4. No-op cases — already in the requested state.
	if enable && isFast {
		return fastPlan{message: "⚡ Already in fast mode (" + short + ")"}
	}
	if !enable && !isFast {
		return fastPlan{message: "Already using base model (" + short + ")"}
	}

	// 5. Resolve the target variant.
	var target modelregistry.ModelEntry
	var ok bool
	if enable {
		target, ok = registry.FastVariant(current)
	} else {
		target, ok = registry.BaseVariant(current)
	}
	if !ok {
		return fastPlan{message: "No fast mode available for " + short}
	}

	// Switching TO a fast model iff the target itself has a base variant.
	_, toFast := registry.BaseVariant(target.ID)
	return fastPlan{targetID: target.ID, toFast: toFast}
}

// handleFastCommand implements `/fast [on|off]`: toggle the session model between
// its base and fast variant within the same family. The actual switch is
// delegated to handleModelCommand — the same path /model uses — so provider
// rebuild, the compaction-before-switch guard, and persistence all apply; this
// never mutates model state directly.
func (m model) handleFastCommand(arg string) (model, string) {
	registry, err := modelregistry.DefaultRegistry()
	if err != nil {
		return m, "Failed to load the model registry: " + err.Error()
	}

	plan := planFastCommand(registry, m.modelName, arg)
	if plan.targetID == "" {
		return m, plan.message
	}

	var switchText string
	m, switchText = m.handleModelCommand(plan.targetID)
	if m.modelName != plan.targetID {
		// The switch did not complete — either an error, or a "compact first"
		// guard prompt from handleModelCommand. Surface its own message so the
		// user sees the real reason instead of a misleading "Switched to ...".
		return m, switchText
	}

	short := modelShortName(registry, plan.targetID)
	if plan.toFast {
		return m, "⚡ Switched to " + short
	}
	return m, "Switched to " + short
}

// modelShortName returns a model's short display name, falling back to the id
// when the model is unknown or has no display name.
func modelShortName(registry modelregistry.Registry, id string) string {
	if entry, ok := registry.Get(id); ok {
		if name := strings.TrimSpace(entry.DisplayName); name != "" {
			return name
		}
	}
	return strings.TrimSpace(id)
}
