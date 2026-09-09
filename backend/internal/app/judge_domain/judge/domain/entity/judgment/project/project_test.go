package project_test

import (
	"judge-backend/internal/app/judge_domain/judge/domain/entity/judgment/project"
	"testing"
)

func TestProject(t *testing.T) {
	t.Run("NewProject", func(t *testing.T) {
		t.Run("should create project with all config values", func(t *testing.T) {
			p := project.NewProject(&project.ProjectConfig{
				ID:           42,
				ImageTag:     "gcc:12",
				TestImageTag: "judge-test:1.0",
				Env:          "CC=gcc\nCFLAGS=-O2",
				TestEnv:      "TIMEOUT=30\nVERBOSE=1",
				MountPath:    "/workspace",
			})

			if p.ID() != 42 {
				t.Errorf("expected ID to be 42, got %d", p.ID())
			}
			if p.ImageTag() != "gcc:12" {
				t.Errorf("expected ImageTag to be 'gcc:12', got %s", p.ImageTag())
			}
			if p.TestImageTag() != "judge-test:1.0" {
				t.Errorf("expected TestImageTag to be 'judge-test:1.0', got %s", p.TestImageTag())
			}
			if p.MountPath() != "/workspace" {
				t.Errorf("expected MountPath to be '/workspace', got %s", p.MountPath())
			}
		})

		t.Run("should split env by newlines and remove empty entries", func(t *testing.T) {
			p := project.NewProject(&project.ProjectConfig{
				ID:           1,
				ImageTag:     "python:3.11",
				TestImageTag: "test:1.0",
				Env:          "A=1\n\nB=2\n",
				TestEnv:      "X=10\nY=20",
				MountPath:    "/app",
			})

			env := p.Env()
			if len(env) != 2 {
				t.Fatalf("expected 2 env vars, got %d: %v", len(env), env)
			}
			if env[0] != "A=1" {
				t.Errorf("expected env[0] to be 'A=1', got %s", env[0])
			}
			if env[1] != "B=2" {
				t.Errorf("expected env[1] to be 'B=2', got %s", env[1])
			}

			testEnv := p.TestEnv()
			if len(testEnv) != 2 {
				t.Fatalf("expected 2 test env vars, got %d: %v", len(testEnv), testEnv)
			}
			if testEnv[0] != "X=10" {
				t.Errorf("expected testEnv[0] to be 'X=10', got %s", testEnv[0])
			}
		})

		t.Run("should handle empty env strings", func(t *testing.T) {
			p := project.NewProject(&project.ProjectConfig{
				ID:           1,
				ImageTag:     "python:3.11",
				TestImageTag: "test:1.0",
				Env:          "",
				TestEnv:      "",
				MountPath:    "/app",
			})

			if len(p.Env()) != 0 {
				t.Errorf("expected empty env, got %v", p.Env())
			}
			if len(p.TestEnv()) != 0 {
				t.Errorf("expected empty test env, got %v", p.TestEnv())
			}
		})
	})
}
