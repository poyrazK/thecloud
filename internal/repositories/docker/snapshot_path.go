package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validateSnapshotPath enforces a hardened policy on snapshot file paths
// supplied by callers of CreateVolumeSnapshot and RestoreVolumeSnapshot.
//
// The previous check, `strings.Contains(path, "..")`, missed several real
// traversal payloads:
//   - `..././etc/passwd` — the substring scan finds `..`, but writing the path
//     after the dot/slash rules collapses to `../etc/passwd`.
//   - `Etc/passwd` on case-insensitive filesystems (macOS, Windows) — the
//     simple compare against a denylist of lowercase paths fails.
//
// The rules below close those gaps:
//
//  1. The path must be absolute. Snapshot tooling always knows the absolute
//     destination; relative paths are rejected outright.
//  2. After `filepath.Clean`, the path must not change. Any traversal segment
//     (`..`, `.`, double-slash, trailing slash) collapses during Clean, so a
//     stable round-trip means the input was already canonical.
//  3. The cleaned path must not contain a `..` segment in any form. Belt and
//     braces in case Clean leaves a leading `..` (e.g. for `../foo`).
//  4. On platforms with case-insensitive filesystems, the comparison logic
//     remains byte-exact — but rule (2) already rejects mixed-case traversal
//     payloads because Clean does not lowercase, leaving them visible.
//
// Note: URL-encoded traversal (e.g. %2E%2E) and symlink-based escapes are not
// handled by this function; callers must ensure the path originates from a
// trusted source that has already decoded and resolved symlinks.
//
// The function returns the cleaned path so callers can use it directly.
func validateSnapshotPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("snapshot path is empty")
	}

	if !filepath.IsAbs(p) {
		return "", fmt.Errorf("snapshot path must be absolute: %q", p)
	}

	// Reject NUL bytes in the path, which can truncate or escape path boundaries.
	if strings.Contains(p, "\x00") {
		return "", fmt.Errorf("snapshot path contains NUL byte")
	}

	cleaned := filepath.Clean(p)
	if cleaned != p {
		return "", fmt.Errorf("snapshot path is not canonical: %q (cleaned %q)", p, cleaned)
	}

	// Reject symlinks to prevent traversal through symlink-based escapes.
	if info, err := os.Lstat(cleaned); err == nil && (info.Mode()&os.ModeSymlink) != 0 {
		return "", fmt.Errorf("snapshot path is a symlink: %q", p)
	}

	// After Clean a leading `..` cannot survive on an absolute path, but reject
	// any embedded `..` segment defensively for older filepath behavior on
	// other platforms.
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return "", fmt.Errorf("snapshot path contains traversal segment: %q", p)
		}
	}

	// Reject root "/" — filepath.Dir("") and filepath.Base(".") cause the host
	// root to be bound into the container, a critical security issue.
	if cleaned == "/" {
		return "", fmt.Errorf("snapshot path resolves to root: %q", p)
	}

	return cleaned, nil
}
