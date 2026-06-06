package tools

// MutationTargets returns the workspace-relative paths a tool call will write to,
// so the session layer can snapshot their before-state for safe rewind. It is a
// pure helper (no I/O beyond path resolution) and returns nil for read-only tools
// and for bash (whose affected paths are not knowable before execution).
func MutationTargets(workspaceRoot string, name string, args map[string]any) []string {
	switch name {
	case "write_file", "edit_file":
		path, err := stringArg(args, "path", "", true)
		if err != nil {
			return nil
		}
		_, relative, err := resolveWorkspaceTargetPath(workspaceRoot, path)
		if err != nil {
			return nil
		}
		return []string{relative}
	case "apply_patch":
		patch, err := stringArg(args, "patch", "", true)
		if err != nil {
			return nil
		}
		paths := changedFilesFromPatch(patch)
		if len(paths) == 0 {
			return nil
		}
		return paths
	default:
		return nil
	}
}
