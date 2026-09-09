package command

import "strings"

type Command struct {
	value string
}

func New(value string) Command {
	return Command{value: value}
}

func (t Command) String() string {
	return t.value
}

func (t Command) Split() []string {
	a := []string{}
	sb := &strings.Builder{}
	quoted := false
	for _, r := range t.value {
		if r == '"' {
			quoted = !quoted
		} else if !quoted && r == ' ' {
			a = append(a, sb.String())
			sb.Reset()
		} else {
			sb.WriteRune(r)
		}
	}

	if sb.Len() > 0 {
		a = append(a, sb.String())
	}

	return a
}
