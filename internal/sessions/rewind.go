package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RestoreReport summarizes a workspace restore.
type RestoreReport struct {
	TargetSequence int      `json:"targetSequence"`
	FilesRestored  int      `json:"filesRestored"`
	FilesDeleted   int      `json:"filesDeleted"`
	Skipped        []string `json:"skipped,omitempty"` // paths whose before-state was not recoverable
}

// RewindMarker is the payload of the EventSessionRewind event appended after a rewind.
type RewindMarker struct {
	TargetSequence int           `json:"targetSequence"`
	Report         RestoreReport `json:"report"`
}

// RestoreToSequence reverts workspace files to their state at targetSeq by applying
// the before-snapshots of every checkpoint after the target, newest-first (so the
// snapshot closest to the target wins). It does not modify the event log.
func (store *Store) RestoreToSequence(sessionID, workspaceRoot string, targetSeq int) (RestoreReport, error) {
	report := RestoreReport{TargetSequence: targetSeq}
	lock := store.sessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	checkpoints, err := store.sortedCheckpointsAfter(sessionID, targetSeq)
	if err != nil {
		return report, err
	}
	// Apply newest -> oldest; a path touched by several checkpoints ends at the
	// oldest (closest-to-target) before-state, which is applied last.
	restored := map[string]bool{}
	for _, ev := range checkpoints {
		var payload CheckpointPayload
		if err := json.Unmarshal(ev.Payload, &payload); err != nil {
			continue
		}
		for _, f := range payload.Files {
			// Defense in depth: never write/delete outside the workspace, even if
			// a checkpoint event was tampered with (path traversal via "../").
			abs, ok := resolveWithinWorkspace(workspaceRoot, f.Path)
			if !ok {
				if !restored[f.Path] {
					report.Skipped = append(report.Skipped, f.Path)
				}
				restored[f.Path] = true
				continue
			}
			switch {
			case f.Skipped:
				if !restored[f.Path] {
					report.Skipped = append(report.Skipped, f.Path)
				}
			case f.Absent:
				if err := os.Remove(abs); err == nil || os.IsNotExist(err) {
					report.FilesDeleted++
				} else {
					report.Skipped = append(report.Skipped, f.Path)
				}
			case f.Blob != "":
				content, rerr := store.readBlob(sessionID, f.Blob)
				if rerr != nil {
					report.Skipped = append(report.Skipped, f.Path)
					continue
				}
				if err := store.writeFileAtomic(abs, content); err != nil {
					report.Skipped = append(report.Skipped, f.Path)
					continue
				}
				report.FilesRestored++
			}
			restored[f.Path] = true
		}
	}
	return report, nil
}

// resolveWithinWorkspace joins rel to root and confirms the result stays inside
// root, rejecting traversal ("../") and absolute escapes.
func resolveWithinWorkspace(root, rel string) (string, bool) {
	abs := filepath.Join(root, rel)
	cleanRoot := filepath.Clean(root)
	within, err := filepath.Rel(cleanRoot, abs)
	if err != nil {
		return "", false
	}
	if within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
		return "", false
	}
	return abs, true
}

// TruncateEvents atomically rewrites events.jsonl keeping only events with
// Sequence <= keepThroughSequence, and updates metadata EventCount.
func (store *Store) TruncateEvents(sessionID string, keepThroughSequence int) error {
	if !ValidSessionID(sessionID) {
		return fmt.Errorf("invalid zero session id %q", sessionID)
	}
	lock := store.sessionLock(sessionID)
	lock.Lock()
	defer lock.Unlock()

	events, err := store.ReadEvents(sessionID)
	if err != nil {
		return err
	}
	var kept [][]byte
	keptCount := 0
	for _, ev := range events {
		if ev.Sequence > keepThroughSequence {
			continue
		}
		data, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("encode kept event: %w", err)
		}
		kept = append(kept, data)
		keptCount++
	}
	var encoded []byte
	if len(kept) > 0 {
		encoded = append(bytes.Join(kept, []byte{'\n'}), '\n')
	}
	path := store.eventsPath(sessionID)
	tmp := fmt.Sprintf("%s.tmp-%d", path, store.idCounter.Add(1))
	if err := os.WriteFile(tmp, encoded, 0o600); err != nil {
		return fmt.Errorf("write truncated events: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("commit truncated events: %w", err)
	}
	session, err := store.readMetadata(sessionID)
	if err != nil {
		return err
	}
	session.EventCount = keptCount
	session.UpdatedAt = store.timestamp()
	return store.writeMetadata(session)
}

// ApplyRewind performs a full safe rewind: restore workspace files to targetSeq,
// truncate the event log, prune now-orphaned blobs, and append an EventSessionRewind
// marker. Returns the restore report.
func (store *Store) ApplyRewind(sessionID, workspaceRoot string, targetSeq int) (RestoreReport, error) {
	report, err := store.RestoreToSequence(sessionID, workspaceRoot, targetSeq)
	if err != nil {
		return report, err
	}
	if err := store.TruncateEvents(sessionID, targetSeq); err != nil {
		return report, err
	}
	_, _ = store.pruneOrphanBlobs(sessionID)
	if _, err := store.AppendEvent(sessionID, AppendEventInput{
		Type:    EventSessionRewind,
		Payload: RewindMarker{TargetSequence: targetSeq, Report: report},
	}); err != nil {
		return report, err
	}
	return report, nil
}

func (store *Store) writeFileAtomic(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.zero-restore-tmp-%d", path, store.idCounter.Add(1))
	if err := os.WriteFile(tmp, content, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
