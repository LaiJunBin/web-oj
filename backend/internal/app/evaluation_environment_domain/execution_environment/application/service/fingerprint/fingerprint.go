// Package fingerprint identifies an image's CONTENT, so an import can tell
// "I already have this" from "I have something else under this name".
//
// The tag cannot answer that. Two installations both holding
// "test-nginx:latest" say nothing about whether the two are the same thing,
// and an importer with only the tag to go on has two equally bad options:
// refuse every image the venue already has, or silently overwrite one that
// several live contests are judging against.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// FileReader is the slice of the filesystem this needs. Narrow on purpose:
// the fingerprint is pure computation over bytes, and a test should not have
// to stand up a filesystem to check it.
type FileReader interface {
	ListFileRecursive(path string) ([]string, error)
	ReadFile(path string) ([]byte, error)
}

// OfBuildContext hashes the files a source bundle carries.
//
// Paths are sorted and both the path and the length are folded in, so the
// result does not depend on the order the filesystem happened to list things,
// and two different file layouts cannot collide by concatenating to the same
// bytes ("ab" + "c" and "a" + "bc" hash differently).
//
// Timestamps and permissions are deliberately excluded: they differ between
// two checkouts of the same content, and an importer that reported a conflict
// because a file's mtime moved would train its operator to always force.
func OfBuildContext(fs FileReader, root string) (string, error) {
	files, err := fs.ListFileRecursive(root)
	if err != nil {
		return "", err
	}
	sort.Strings(files)

	h := sha256.New()
	for _, f := range files {
		body, err := fs.ReadFile(f)
		if err != nil {
			return "", err
		}
		// The path is recorded relative to the context root, so the same
		// context fingerprints the same on a machine that stores it
		// elsewhere.
		rel := strings.TrimPrefix(strings.TrimPrefix(f, root), "/")
		fmt.Fprintf(h, "%s\x00%d\x00", rel, len(body))
		h.Write(body)
	}

	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

// OfDigest wraps the digest the daemon reports for a built image. Already a
// content hash, so it is used as-is rather than re-hashed.
func OfDigest(digest string) string {
	return digest
}
