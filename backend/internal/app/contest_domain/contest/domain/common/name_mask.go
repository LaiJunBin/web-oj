package common

import "strings"

// MaskName is THE masking rule for contestant display names on
// public/contestant-facing scoreboards (admin views never mask). It is
// deliberately a single free function so the masking policy can be changed in
// one place. Current rule: keep the first and last rune, replace everything
// in between with '*'; names shorter than three runes degrade to "first rune
// + *" / "*" so nothing of a 1–2 character name leaks besides its initial.
func MaskName(name string) string {
	runes := []rune(name)
	switch {
	case len(runes) == 0:
		return ""
	case len(runes) == 1:
		return "*"
	case len(runes) == 2:
		return string(runes[0]) + "*"
	default:
		return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
	}
}
