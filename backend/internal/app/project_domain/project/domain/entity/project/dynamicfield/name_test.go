package dynamicfield

import "testing"

func TestNewName(t *testing.T) {
	valid := []string{"answer", "answer.txt", "my_field-1", "資料檔"}
	for _, v := range valid {
		if _, err := NewName(v); err != nil {
			t.Errorf("NewName(%q) rejected a legitimate key: %v", v, err)
		}
	}

	unsafe := []string{
		"",
		".",
		"..",
		"../../config.yaml",
		"a/b",
		`a\b`,
		"foo/../bar",
		"with\x00null",
	}
	for _, v := range unsafe {
		if _, err := NewName(v); err == nil {
			t.Errorf("NewName(%q) accepted an unsafe key", v)
		}
	}
}
