package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillContentEmbedded(t *testing.T) {
	if !strings.HasPrefix(skillContent, "---\n") {
		t.Error("Expected SKILL.md to start with YAML frontmatter")
	}
	if !strings.Contains(skillContent, "name: lmk") {
		t.Error("Expected SKILL.md to declare name: lmk")
	}
	if !strings.Contains(skillContent, "description:") {
		t.Error("Expected SKILL.md to declare a description")
	}
}

func TestSkillInstallDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to resolve home: %v", err)
	}

	t.Run("default_global", func(t *testing.T) {
		got, err := skillInstallDir(false, "")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(home, ".claude", "skills", "lmk")
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("project", func(t *testing.T) {
		got, err := skillInstallDir(true, "")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(".claude", "skills", "lmk")
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("custom_path_wins", func(t *testing.T) {
		got, err := skillInstallDir(true, "/tmp/my-skills")
		if err != nil {
			t.Fatal(err)
		}
		if got != "/tmp/my-skills" {
			t.Errorf("expected /tmp/my-skills, got %q", got)
		}
	})
}

func TestInstallSkillWritesFile(t *testing.T) {
	tempDir := t.TempDir()
	installSkill([]string{"--path", tempDir})

	target := filepath.Join(tempDir, "SKILL.md")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("expected SKILL.md at %s: %v", target, err)
	}
	if string(data) != skillContent {
		t.Error("written file does not match embedded content")
	}
}
