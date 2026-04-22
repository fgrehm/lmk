package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

//go:embed skill/SKILL.md
var skillContent string

// handleSkill dispatches the `lmk skill` subcommand.
// With no args, prints SKILL.md to stdout. With `install`, writes it to disk.
func handleSkill(args []string) {
	if len(args) == 0 {
		fmt.Print(skillContent)
		return
	}

	if args[0] == "install" {
		installSkill(args[1:])
		return
	}

	log.Fatalf("Unknown skill subcommand: %s\nRun `lmk skill` to print, or `lmk skill install` to install.", args[0])
}

func installSkill(args []string) {
	skillFlags := flag.NewFlagSet("skill install", flag.ExitOnError)
	project := skillFlags.Bool("project", false, "Install to ./.claude/skills/lmk/ in the current directory")
	customPath := skillFlags.String("path", "", "Install to a specific directory (overrides --project and the default)")
	dryRun := skillFlags.Bool("dry-run", false, "Show where the skill would be written without writing it")
	skillFlags.Parse(args)

	dir, err := skillInstallDir(*project, *customPath)
	if err != nil {
		log.Fatalf("Failed to resolve install directory: %v", err)
	}
	target := filepath.Join(dir, "SKILL.md")

	if *dryRun {
		fmt.Println("🔍 Dry run mode - no files were written")
		fmt.Println()
		fmt.Printf("Would write skill to: %s\n", target)
		return
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(target, []byte(skillContent), 0644); err != nil {
		log.Fatalf("Failed to write skill file: %v", err)
	}

	fmt.Println("✅ Claude Code skill installed successfully!")
	fmt.Println()
	fmt.Printf("Written to: %s\n", target)
	fmt.Println()
	fmt.Println("Trigger phrases: \"lmk when done\", \"let me know when X finishes\", \"notify me when ready\"")
	fmt.Println()
	fmt.Println("Try it: Start a Claude Code session and ask \"run the tests and lmk when done\".")
}

// skillInstallDir resolves the target directory for the skill.
// Precedence: --path > --project > user-global (~/.claude/skills/lmk).
func skillInstallDir(project bool, customPath string) (string, error) {
	if customPath != "" {
		abs, err := filepath.Abs(customPath)
		if err != nil {
			return "", err
		}
		return abs, nil
	}
	if project {
		return filepath.Join(".claude", "skills", "lmk"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills", "lmk"), nil
}
