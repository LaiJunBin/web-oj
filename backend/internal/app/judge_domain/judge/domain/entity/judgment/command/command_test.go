package command_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/command"
	"testing"
)

func TestCommand(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		cmd := command.New("echo hello")
		if cmd.String() != "echo hello" {
			t.Errorf("expected 'echo hello', got %s", cmd.String())
		}
	})

	t.Run("Split", func(t *testing.T) {
		t.Run("should split simple command by spaces", func(t *testing.T) {
			cmd := command.New("echo hello world")
			parts := cmd.Split()

			expected := []string{"echo", "hello", "world"}
			if len(parts) != len(expected) {
				t.Fatalf("expected %d parts, got %d: %v", len(expected), len(parts), parts)
			}
			for i, v := range expected {
				if parts[i] != v {
					t.Errorf("expected parts[%d] to be %q, got %q", i, v, parts[i])
				}
			}
		})

		t.Run("should handle quoted strings", func(t *testing.T) {
			cmd := command.New(`echo "hello world" test`)
			parts := cmd.Split()

			expected := []string{"echo", "hello world", "test"}
			if len(parts) != len(expected) {
				t.Fatalf("expected %d parts, got %d: %v", len(expected), len(parts), parts)
			}
			for i, v := range expected {
				if parts[i] != v {
					t.Errorf("expected parts[%d] to be %q, got %q", i, v, parts[i])
				}
			}
		})

		t.Run("should handle single word command", func(t *testing.T) {
			cmd := command.New("ls")
			parts := cmd.Split()

			if len(parts) != 1 || parts[0] != "ls" {
				t.Errorf("expected ['ls'], got %v", parts)
			}
		})

		t.Run("should handle empty command", func(t *testing.T) {
			cmd := command.New("")
			parts := cmd.Split()

			if len(parts) != 0 {
				t.Errorf("expected empty slice, got %v", parts)
			}
		})

		t.Run("should handle multiple quoted segments", func(t *testing.T) {
			cmd := command.New(`cmd "arg one" "arg two"`)
			parts := cmd.Split()

			expected := []string{"cmd", "arg one", "arg two"}
			if len(parts) != len(expected) {
				t.Fatalf("expected %d parts, got %d: %v", len(expected), len(parts), parts)
			}
			for i, v := range expected {
				if parts[i] != v {
					t.Errorf("expected parts[%d] to be %q, got %q", i, v, parts[i])
				}
			}
		})

		t.Run("should handle empty quoted string", func(t *testing.T) {
			cmd := command.New(`echo "" end`)
			parts := cmd.Split()

			expected := []string{"echo", "", "end"}
			if len(parts) != len(expected) {
				t.Fatalf("expected %d parts, got %d: %v", len(expected), len(parts), parts)
			}
			for i, v := range expected {
				if parts[i] != v {
					t.Errorf("expected parts[%d] to be %q, got %q", i, v, parts[i])
				}
			}
		})
	})
}
