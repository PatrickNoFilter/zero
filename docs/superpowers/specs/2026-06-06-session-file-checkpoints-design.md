# Session File Checkpoints + Safe Rewind — Design Spec

**Date:** 2026-06-06
**Owner:** Gnanam (runtime core)
**Source:** Slice 2 of `gnanam-good-modules` integration. Report P0 #2 (top safety pick); references `sessions/persistence.ts`, `sessions/checkpoint_store.ts`.
**Scope chosen:** Full safe-rewind (capture + restore files + truncate event log + TUI & headless commands + stream-json events).

## Overview

Today rewind is planning-only (`PlanRewind`/`PlanCompaction`) — there is no file-content
checkpointing and no `ApplyRewind`. This slice adds durable before-mutation file
snapshots and a complete, cross-platform "undo to a checkpoint" for all users (TUI +
headless), so the agent can mutate the workspace autonomously and safely roll back.

## Current state (grounding)

- Sessions: `$XDG_DATA_HOME/zero/sessions/{id}/` with `metadata.json` (atomic tmp-rename)
  + append-only `events.jsonl` (0600); dirs 0700; per-session mutex (`store.go`).
- `EventSessionRewind` type defined but unused (`store.go:35`).
- `PlanRewind`/`PlanCompaction` are planning-only — **no ApplyRewind, no truncation, no
  file checkpoints** (`replay.go:56-158`).
- `OnToolCall` fires in the loop *before* each tool runs (`agent/loop.go`, before
  `executeToolCall`) → the capture hook. `Result.ChangedFiles` (added in slice 1)
  confirms touched paths after.
- Storage idioms: 0700/0600, atomic tmp-rename, `filepath.*` everywhere, redaction available.

## Architecture

### 1. Mutation target discovery — `tools.MutationTargets(name, args) []string`
New pure helper in `internal/tools` (it owns tool arg shapes). Returns workspace-relative
paths a tool *will* touch:
- `write_file` → `[path]`; `edit_file` → `[path]`; `apply_patch` → `changedFilesFromPatch(patch)`.
- `bash` → `nil` (paths unknowable pre-exec — **deferred**, documented).
- Unknown/read-only tools → `nil`.

### 2. Checkpoint store — `internal/sessions/checkpoint.go`
- Blobs: `{sessionDir}/checkpoints/blobs/{sha256}` (0600), content-addressed (dedup).
  Dirs 0700. Written atomically (tmp-rename); a blob that already exists is left as-is.
- Index = a new session event `EventSessionCheckpoint` ("session_checkpoint") appended via
  the existing `Store.AppendEvent` (ordered, rewind-aware, no separate index file). Payload:
  ```
  { sequence, tool, files: [ { path, blob: "<sha256>"|"", absent: bool, skipped: bool, bytes: int } ] }
  ```
  `absent:true` = file did not exist before (restore ⇒ delete). `skipped:true` = exceeded
  size cap (restore can't recover; surfaced as a warning).
- API:
  - `CaptureToolCheckpoint(store, sessionID, workspaceRoot string, seq int, tool string, paths []string) error`
    — reads each path's current bytes, writes/dedups blob, appends the checkpoint event.
  - `RestoreToSequence(store, sessionID, workspaceRoot string, targetSeq int) (RestoreReport, error)`
    — see §4.
- Size cap: `maxCheckpointBytes` default 5 MiB (override via `ZERO_CHECKPOINT_MAX_BYTES`);
  larger files recorded as `skipped`, not blobbed.

### 3. Capture wiring (TUI + headless parity)
A shared helper `sessions.CaptureForToolCall(...)` called from `OnToolCall` in **both**
`internal/cli/exec.go` and `internal/tui/model.go`, after the existing OnToolCall logic.
It computes `tools.MutationTargets`, and if non-empty, calls `CaptureToolCheckpoint` with the
current event sequence. Capture happens before the mutation runs (OnToolCall precedes
`executeToolCall`). Denied tools may produce a harmless no-op checkpoint (restore = same content).

### 4. Restore + ApplyRewind
- `Store.TruncateEvents(sessionID string, keepThroughSequence int) error` — atomically rewrite
  `events.jsonl` keeping events with `Sequence <= keepThroughSequence` (tmp-rename), update
  `metadata.json` EventCount.
- `RestoreToSequence`: iterate `session_checkpoint` events with `sequence > targetSeq` from
  **newest → oldest**; for each file, restore its recorded before-content (write blob bytes;
  `absent` ⇒ remove file). Newest-first means the snapshot closest to the target is applied
  last and wins. Returns `RestoreReport{ FilesRestored, FilesDeleted, Skipped[] }`.
- `ApplyRewind(store, sessionID, workspaceRoot, targetSeq)`: `RestoreToSequence` →
  `TruncateEvents(targetSeq)` → append `EventSessionRewind` marker `{targetSequence, report}`.
  Order matters: restore files first (uses checkpoint events), then truncate, then mark.

### 5. Commands (all users)
- Headless: `zero sessions rewind <id> --to <seq>` (and `--to latest-checkpoint`), printing the
  `RestoreReport`; honors `--json`. Extends `internal/cli/sessions.go` (which already has `rewind-plan`).
- TUI: `/rewind [N|latest]` command (mirrors `/compact`), resolves target, calls `ApplyRewind`,
  truncates the transcript view to match, shows the report.

### 6. Stream-json events
Add `EventCheckpoint` ("checkpoint") and `EventRestore` ("restore") to `streamjson`. Emit a
`checkpoint` event when a checkpoint is captured (sequence, tool, file count, bytes) and a
`restore` event on rewind (target, filesRestored/Deleted/skipped). Headless observers can audit.

### 7. zerocommands snapshot
If `zerocommands` exposes session state, add checkpoint counts/last-checkpoint to the snapshot
(verify; add only if it currently surfaces session shape).

## "All users / all platforms" requirements
- `filepath.*` only; blobs are raw bytes (binary-safe); 0700 dirs / 0600 files; atomic
  tmp-rename for blob, log truncation, metadata.
- **No redaction of blob content** — restore fidelity requires raw bytes; blobs are the user's
  own files stored exactly as securely as the session. (Checkpoint *event payloads* carry only
  paths+hashes, which are safe.) This is an explicit, documented divergence from slice 1.
- Disk discipline: content-addressed dedup; per-file size cap; blobs pruned when a session is
  deleted; orphan-blob prune helper (blobs unreferenced by any checkpoint event).
- Opt-out: `ZERO_CHECKPOINTS=off` disables capture (rewind then reports "no checkpoints").
- Concurrency: capture/truncate run under the existing per-session mutex.
- Large logs: `TruncateEvents` streams line-by-line rather than holding all in memory where feasible.

## Error handling
- Capture failures (unreadable file, disk full) must **never** fail the tool run — log/skip the
  file as `skipped` and continue (checkpointing is best-effort safety, not a gate).
- Restore validates the target sequence exists; missing blob ⇒ report `skipped`, continue others.
- Truncation uses tmp-rename; a crash mid-write leaves the original intact.

## Testing (TDD)
1. `MutationTargets` returns correct paths per tool; `nil` for bash/read-only.
2. Capture writes a dedup'd blob + a `session_checkpoint` event with the right payload; identical
   content reuses one blob.
3. Size cap: a >cap file is recorded `skipped`, no blob written.
4. Round-trip: write_file creates a file → checkpoint → edit it → `RestoreToSequence` reverts to
   the captured content; `absent` before-state ⇒ restore deletes the created file.
5. Newest→oldest precedence: two edits to one file, restore to before-both yields original.
6. `TruncateEvents` keeps `<= seq`, updates EventCount, is atomic (tmp-rename), contiguous.
7. `ApplyRewind` end-to-end: files restored + log truncated + `EventSessionRewind` appended.
8. Capture never fails the tool run when a path is unreadable.
9. `ZERO_CHECKPOINTS=off` disables capture.
10. Headless `zero sessions rewind` and stream-json `checkpoint`/`restore` events round-trip.

## DoD (report)
- [ ] Typed stream-json events for checkpoint/restore.
- [ ] zerocommands snapshot updated iff it exposes session shape.
- [ ] Unit + integration tests incl. restore round-trip and truncation atomicity.
- [ ] Works headless (exec.go) and TUI (model.go) with command parity.
- [ ] `go build`, `go vet`, `go test -race ./...` green.

## Out of scope / deferred
- bash mutation checkpointing (no upfront paths) — future fs-scan / sandbox-reported mutations.
- Compaction execution (`ApplyCompaction`) — separate concern (this slice does rewind, not compaction).
- Cross-session checkpoint sharing / cross-session memory (reference mentions it; later slice).
