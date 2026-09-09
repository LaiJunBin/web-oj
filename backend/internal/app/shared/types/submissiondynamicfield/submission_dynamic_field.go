package submissiondynamicfield

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

var (
	typeFile = Type{"file"}
	typeText = Type{"text"}
	typeBool = Type{"bool"}
)

type Type struct {
	value string
}

func (t Type) Equal(other Type) bool { return t.value == other.value }
func (t Type) IsFile() bool         { return t.Equal(typeFile) }
func (t Type) IsText() bool         { return t.Equal(typeText) }
func (t Type) IsBool() bool         { return t.Equal(typeBool) }

// SubmissionDynamicField represents a project-defined extra field
// submitted by a contestant alongside their project code.
type SubmissionDynamicField struct {
	typ   Type
	key   string
	value interface{}
}

func newField(t Type, key string, value interface{}) *SubmissionDynamicField {
	return &SubmissionDynamicField{typ: t, key: key, value: value}
}

func NewText(key, value string) *SubmissionDynamicField {
	return newField(typeText, key, value)
}

func NewBool(key, value string) (*SubmissionDynamicField, error) {
	if value != "true" && value != "false" {
		return nil, fmt.Errorf("invalid bool value: %s", value)
	}
	return newField(typeBool, key, value == "true"), nil
}

func NewFile(key, base64Str string) (*SubmissionDynamicField, error) {
	if base64Str == "" {
		return nil, fmt.Errorf("empty base64 value")
	}
	parts := strings.Split(base64Str, ",")
	if len(parts) != 2 {
		return nil, errors.New("invalid base64 format")
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	return newField(typeFile, key, decoded), nil
}

func (f *SubmissionDynamicField) Type() Type       { return f.typ }
func (f *SubmissionDynamicField) Key() string      { return f.key }
func (f *SubmissionDynamicField) Value() interface{} { return f.value }
