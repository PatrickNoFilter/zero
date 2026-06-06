# Reliability Batch — Design Spec

**Date:** 2026-06-06
**Owner:** Gnanam (runtime core)
**Source:** First slice of the `gnanam-good-modules` integration, per `references/gnanam-portion/GNAMAM_RUNTIME_CORE_REFERENCE_REPORT.md` (P0 #4 + P1 structured-result).

## Overview

Three reliability improvements to the tool-execution path, ported from the
distilled `tools/structured_result.ts`, `utils/secret_scrubber.ts`, and
`utils/embedded_tool.ts` patterns. The goal: no secret ever leaves a tool and
reaches the model/stream/logs, and tool results carry structured metadata
(changed files, display summary) instead of an opaque string.

This is an integration slice, not a greenfield subsystem — Zero already has a
strong secret redactor and directory-exclusion set. The work is wiring existing
capability to the right boundary plus two additive struct fields.

## Scope

**In scope:**
1. Secret scrubbing applied at the tool-output boundary.
2. Structured `ToolResult` fields: `ChangedFiles`, `Display{Summary, Kind}`.
3. Regression test confirming always-excluded directories.

**Explicitly out / already satisfied:**
- **Embedded-binary reliability + `AGENT_RIPGREP_PATH` override (`utils/embedded_tool.ts`)**:
  N/A. Zero's grep is pure-Go (`internal/tools/grep.go` walks files with
  `regexp`); there is no shelled-out ripgrep, so the AV/EDR/corporate
  binary-reliability problem does not exist here.
- **Always-exclude dirs**: already implemented. `ignoredDirectories`
  (`internal/tools/workspace.go:10`) excludes `.git`, `node_modules`, `dist`,
  `build`, `.next`, `.turbo`, `coverage`, `.cache`, `tmp`, `temp`; grep/glob/list
  honor it via `shouldSkipDirectory`. We add a regression test only.
- Real OS sandbox adapters, file checkpoints, model registry depth, MCP surface
  — separate slices, separate specs.

## Current state (grounding)

- `redaction.RedactString(value, Options)` (`internal/redaction/redaction.go:107`)
  already scrubs: OpenAI (`sk-`, `sk-proj-`), Anthropic (`sk-ant-api…`), GitHub
  (`github_pat_`, `ghp_/gho_/…`), GitLab (`glpat-`), Google (`AIza…`), Slack
  (`xox[baprs]-`), AWS (`AKIA/ASIA…`), JWTs (`eyJ…`), plus `KEY=value`,
  sensitive JSON keys, auth headers, and URL credentials. The `textSecretPatterns`
  apply unconditionally (independent of `Options`).
- The gap (per the report): this redactor is **not applied to tool output**
  before it reaches the model/stream.
- `agent.executeToolCall` (`internal/agent/loop.go:100`) is the single chokepoint:
  its returned `ToolResult.Output` becomes the tool message AND is passed to
  `options.OnToolResult`, which fans out to the TUI (`tui/model.go:655`),
  session recording (`cli/exec.go:225`), and stream-json
  (`cli/exec_writer.go:123`).
- `tools.Result` (`internal/tools/types.go:54`) has `Status, Output, Truncated,
  Meta, SandboxDecision`. `streamjson.Event` (`streamjson.go:41`) already mirrors
  `Status, Output, Truncated, Meta`. Missing on both: changed-files + display.

## Architecture

### Component 1 — Secret scrubbing at the boundary

**Placement:** inside `executeToolCall`, immediately after `registry.RunWithOptions`
returns and before the `ToolResult` is constructed/returned. (Chosen over the
registry level so internal callers that may need raw output — e.g. a future
checkpoint differ — are unaffected; chosen over per-tool scrubbing so no tool
can forget.)

**Behavior:**
- `scrubbed := redaction.RedactString(result.Output, redaction.Options{})` —
  using the same zero-value `Options{}` idiom already used in `internal/verify`,
  `internal/doctor`, `internal/selfverify`, and `tui/command_output.go`.
  `textSecretPatterns` apply unconditionally; the default replacement token is
  `RedactedSecret`.
- If `scrubbed != result.Output`: set `Redacted = true` and append a single
  trailing line: `\n[secrets redacted for safety]`.
- New field `Redacted bool` on `tools.Result` and `agent.ToolResult`, mirrored as
  `redacted *bool` (omitempty) on `streamjson.Event` for observability.

**Data flow:** model message, TUI row, session event, and stream-json event all
receive the scrubbed output because they derive from the single returned/
broadcast `ToolResult`.

### Component 2 — Structured ToolResult

**New fields** on `tools.Result`, `agent.ToolResult`, and `streamjson.Event`:
- `ChangedFiles []string` — workspace-relative paths a tool mutated.
- `Display struct { Summary string; Kind string }` — short human/stream summary
  and a kind tag (`file`, `diff`, `search`, `shell`, …).

**Population:**
- `write_file` → `ChangedFiles=[path]`, `Display{Summary:"Wrote <path> (<n> lines)", Kind:"file"}`.
- `edit_file` → `ChangedFiles=[path]`, `Display{Summary:"Edited <path>", Kind:"diff"}`.
- `apply_patch` → `ChangedFiles=[…all touched paths…]`, `Display{Kind:"diff"}`.
- `read_file`/`grep`/`bash` → `Display` summary only (no `ChangedFiles`).
- `streamjson.Event` gains `changedFiles []string` and `display {summary,kind}`
  (both omitempty); `cli/exec_writer.go` and the TUI populate them from
  `ToolResult`. `zerocommands` snapshot updated only if it currently exposes
  tool-result shape (verify; add if so).

### Component 3 — Always-exclude regression test

A test asserting grep (and glob) never descend into `.git`/`node_modules`, so the
guarantee can't silently regress. Plus a one-line doc note in the spec/code that
the embedded-binary override is intentionally not ported.

## Error handling

- Scrubbing is pure string transformation; it cannot fail. If `RedactString`
  somehow panics it would surface through the existing tool-execution path — no
  special handling added.
- Scrubbing runs on every tool result including error outputs (error messages can
  leak secrets too).
- Over-redaction risk is low: `textSecretPatterns` match high-entropy,
  prefix-anchored token shapes; ordinary code/prose is unaffected. Accepted,
  safety-first, per the report's "redaction never leaks" DoD.

## Testing (TDD)

1. `executeToolCall` scrubs `sk-…`/`ghp_…` from `Output` → assert neither the
   returned `ToolResult.Output`, the appended model message, nor the
   `OnToolResult` payload contains the token; `Redacted == true`; reminder present.
2. Clean output is unchanged and `Redacted == false` (no false reminder).
3. `edit_file`/`write_file`/`apply_patch` populate `ChangedFiles` with the right
   relative paths.
4. Each tool sets a sensible `Display.Summary/Kind`.
5. `streamjson.Event` round-trips the new fields (JSON marshal omitempty).
6. Regression: grep/glob skip `.git` and `node_modules`.

## Definition of Done (report DoD)

- [ ] Typed `streamjson` fields for new user/automation-visible data.
- [ ] `zerocommands` snapshot updated if it exposes tool-result state.
- [ ] Unit tests incl. a redaction-never-leaks regression test.
- [ ] Works in headless/stream-json mode (covered by the exec_writer path).
- [ ] `go build ./...`, `go vet ./...`, `go test -race ./...` all green.

## Risks / open questions

- `apply_patch` changed-file extraction: parse the patch/`git apply` summary for
  touched paths. If non-trivial, fall back to the target path(s) it was given.
