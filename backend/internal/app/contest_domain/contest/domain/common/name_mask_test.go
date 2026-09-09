package common

import "testing"

func TestMaskName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single rune", "王", "*"},
		{"two runes", "小明", "小*"},
		{"three runes keeps ends", "王小明", "王*明"},
		{"ascii name", "Alice", "A***e"},
		{"unicode middle fully masked", "歐陽小小明", "歐***明"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := MaskName(c.input); got != c.expected {
				t.Errorf("MaskName(%q) = %q, expected %q", c.input, got, c.expected)
			}
		})
	}
}
