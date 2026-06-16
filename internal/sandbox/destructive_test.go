package sandbox

import (
	"context"
	"testing"
)

// TestScopedDestructivePromptsWhileCatastrophicDenies pins the split between a
// scoped delete a user can approve and the irrecoverable system-level forms that
// stay a hard block even in unsafe mode.
func TestScopedDestructivePromptsWhileCatastrophicDenies(t *testing.T) {
	engine := NewEngine(EngineOptions{WorkspaceRoot: t.TempDir(), Policy: DefaultPolicy()})

	shellReq := func(command string, mode PermissionMode) Request {
		return Request{
			ToolName:       "bash",
			SideEffect:     SideEffectShell,
			Permission:     PermissionPrompt,
			PermissionMode: mode,
			Autonomy:       AutonomyHigh,
			Args:           map[string]any{"command": command},
		}
	}

	// rm -rf <subdir> in ask mode is now a prompt the user can approve, not a hard
	// block — and it must be "destructive" but NOT "destructive_catastrophic".
	scoped := engine.Evaluate(context.Background(), shellReq("rm -rf site", PermissionModeAsk))
	if scoped.Action != ActionPrompt {
		t.Fatalf("rm -rf <subdir> in ask mode = %#v, want ActionPrompt", scoped)
	}
	if !HasRiskCategory(scoped.Risk, "destructive") || HasRiskCategory(scoped.Risk, "destructive_catastrophic") {
		t.Fatalf("rm -rf <subdir> categories = %v, want destructive (not catastrophic)", scoped.Risk.Categories)
	}

	// The same scoped delete is allowed in unsafe mode (the operator opted in).
	if d := engine.Evaluate(context.Background(), shellReq("rm -rf site", PermissionUnsafe)); d.Action != ActionAllow {
		t.Fatalf("rm -rf <subdir> in unsafe = %#v, want ActionAllow", d)
	}

	// A nested relative path is still scoped, not catastrophic.
	if d := engine.Evaluate(context.Background(), shellReq("rm -rf build/output", PermissionModeAsk)); d.Action != ActionPrompt {
		t.Fatalf("rm -rf build/output in ask = %#v, want ActionPrompt", d)
	}

	// Catastrophic system-level commands stay a hard block — even in unsafe mode.
	catastrophic := []string{
		"rm -rf /",
		"rm -rf $HOME",
		"rm -rf ~",
		"mkfs.ext4 /dev/sda1",
		"dd if=/dev/zero of=/dev/sda",
		":(){ :|:& };:",
	}
	for _, command := range catastrophic {
		d := engine.Evaluate(context.Background(), shellReq(command, PermissionUnsafe))
		if d.Action != ActionDeny || d.Violation == nil || d.Violation.Code != ViolationDestructiveCommand {
			t.Fatalf("catastrophic %q in unsafe = %#v, want destructive deny", command, d)
		}
	}
}
