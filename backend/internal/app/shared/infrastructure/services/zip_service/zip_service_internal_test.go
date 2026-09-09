package zipservice

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsArchiverJunk pins the narrow deny-list that replaced the old
// "skip any dot-prefixed path segment" rule. The distinction that matters:
// entries an archiving TOOL invented (Finder's __MACOSX sidecar, .DS_Store,
// AppleDouble ._ files, Explorer's Thumbs.db) are noise nobody authored,
// while every other dotfile is a real project file a contestant wrote on
// purpose. Anything not on this list must be extracted.
func TestIsArchiverJunk(t *testing.T) {
	junk := []string{
		"__MACOSX/myproject/._Dockerfile",
		"__MACOSX/",
		".DS_Store",
		"src/.DS_Store",
		"._Dockerfile",
		"src/._index.js",
		"Thumbs.db",
		"assets/Thumbs.db",
		"desktop.ini",
	}
	for _, name := range junk {
		t.Run(name+" is junk", func(t *testing.T) {
			assert.True(t, isArchiverJunk(name))
		})
	}

	kept := []string{
		".env",
		".env.production",
		".gitignore",
		".htaccess",
		".dockerignore",
		".github/workflows/ci.yml",
		".vscode/settings.json",
		"src/.eslintrc",
		"Dockerfile",
		"src/index.js",
		// A leading ".." is only an escape when it is the whole segment;
		// "..config.txt" is an ordinary filename.
		"..config.txt",
		// Prefix boundary: a directory merely STARTING with "__MACOSX" is a
		// different directory, so the check must require the separator.
		"__MACOSXfoo/a.txt",
		// __MACOSX is only Finder's sidecar at the archive root. A nested
		// directory a contestant happened to name that way is their file.
		"src/__MACOSX/a.txt",
		// Only a basename STARTING with "._" is AppleDouble.
		"foo._bar",
		// Junk names are matched exactly, not as substrings.
		"my.DS_Store.txt",
	}
	for _, name := range kept {
		t.Run(name+" is kept", func(t *testing.T) {
			assert.False(t, isArchiverJunk(name))
		})
	}
}

// TestIsWithinDir is a regression test for A3 (zip-slip): UnzipToPath used to
// rely entirely on the old isHidden treating any ".." path segment as a
// "hidden" file to skip — an incidental side effect, not an intentional
// containment check. isWithinDir is the explicit guard that makes extraction
// safe on its own, which is what let the dot-prefix rule be narrowed to
// isArchiverJunk without reopening zip-slip.
func TestIsWithinDir(t *testing.T) {
	dir := "/tmp/unzip-abc123"

	t.Run("dir itself is within dir", func(t *testing.T) {
		assert.True(t, isWithinDir(dir, dir))
	})

	t.Run("a nested descendant is within dir", func(t *testing.T) {
		assert.True(t, isWithinDir(dir, dir+"/sub/file.txt"))
	})

	t.Run("a sibling directory sharing a name prefix is NOT within dir", func(t *testing.T) {
		// "/tmp/unzip-abc123-evil" has "/tmp/unzip-abc123" as a naive string
		// prefix but is a sibling, not a descendant — isWithinDir must check
		// the path separator boundary, not just string prefix.
		assert.False(t, isWithinDir(dir, dir+"-evil/secret.txt"))
	})

	t.Run("a parent directory is NOT within dir", func(t *testing.T) {
		assert.False(t, isWithinDir(dir, "/tmp"))
	})

	t.Run("an unrelated absolute path is NOT within dir", func(t *testing.T) {
		assert.False(t, isWithinDir(dir, "/etc/passwd"))
	})

	t.Run("what filepath.Join+Clean produces for a zip-slip entry name is rejected", func(t *testing.T) {
		// Mirrors what UnzipToPath does: filepath.Join(tmpdir, f.Name) before
		// the containment check. A zip entry named "../../../etc/passwd"
		// joins-and-cleans to an absolute path entirely outside dir.
		joined := filepath.Join(dir, "../../../etc/passwd")
		assert.Equal(t, "/etc/passwd", joined)
		assert.False(t, isWithinDir(dir, joined))
	})
}
