// Package bundlecodec holds the JSON rules every export/import bundle obeys.
//
// Shared rather than repeated per bundle kind because the three decisions
// below have to be the SAME decisions everywhere. Two copies drift: somebody
// turns off HTML escaping in one and a year later the other is still writing
// < into shell commands, and nothing fails until a person tries to read
// the file.
package bundlecodec

import (
	"bytes"
	"encoding/json"
)

// indent is fixed so two exports of unchanged data produce identical bytes.
const indent = "  "

// Marshal renders a bundle document.
//
// Deterministic by construction: bundle documents are structs, and Go emits
// struct fields in declaration order. Nothing in a bundle may become a map —
// Go sorts map keys, which looks stable right up until somebody relies on the
// order matching the declaration.
//
// HTML escaping is off. Bundles carry shell commands, and escaping < > & would
// leave a file that is technically correct and unreadable to whoever has to
// edit it.
func Marshal(v any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", indent)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal parses a bundle document, rejecting fields it does not know.
//
// Strict on purpose. A bundle may be hand-written, and the typos that hurt are
// the ones on OPTIONAL fields: a misspelled boolean leaves its default in
// place, the import succeeds, and the mistake surfaces later as behaviour
// nobody asked for. Required fields are already caught by validation, because
// their zero value fails it; unknown-field rejection covers the rest.
//
// The cost is that ADDING a field is a breaking change for older importers, so
// any schema change bumps that bundle's format version — additive ones
// included. A version bump is a visible one-line decision; a silently ignored
// field is a contest that behaves differently than its author intended.
func Unmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
