package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CmdSetup adds xoji integration instructions to a Xojo project's CLAUDE.md
func CmdSetup(projectPath string) error {
	// Find .xojo_project if not specified
	if projectPath == "" {
		var err error
		projectPath, err = FindXojoProject("")
		if err != nil {
			return err
		}
	} else {
		// Check if projectPath is a directory or a file
		fi, err := os.Stat(projectPath)
		if err != nil {
			return fmt.Errorf("invalid project path: %w", err)
		}

		// If it's a directory, search for .xojo_project in it
		if fi.IsDir() {
			var err error
			projectPath, err = FindXojoProject(projectPath)
			if err != nil {
				return err
			}
		}
	}

	// Get project directory
	projectDir := filepath.Dir(projectPath)

	// Look for CLAUDE.md in the project directory
	claudeMdPath := filepath.Join(projectDir, "CLAUDE.md")

	// Read existing CLAUDE.md if it exists
	var content string
	if _, err := os.Stat(claudeMdPath); err == nil {
		bytes, err := os.ReadFile(claudeMdPath)
		if err != nil {
			return fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		content = string(bytes)
	} else {
		// Create a basic CLAUDE.md if it doesn't exist
		content = "# CLAUDE.md\n\nThis file provides guidance to Claude Code when working with this Xojo project.\n\n"
	}

	// Check if xoji instructions already exist
	if strings.Contains(content, "## Using xoji for token-efficient indexing") {
		fmt.Println("✓ xoji instructions already present in CLAUDE.md")
		return nil
	}

	// Append xoji instructions
	xojiInstructions := `## Using xoji for token-efficient indexing

### Before starting any task:
` + "```bash\n" + `# From project root, run freshness check
xoji check || xoji index
` + "```\n" + `
### Index files in .xojo_index/:
- **codetree.json** — Maps file paths to {entity, methods, properties, events, line numbers}
- **manifest.json** — All files with their types and entity names
- **dependencies.json** — Class inheritance and interface relationships
- **meta.json** — Project hash and file modification times (for freshness)

### How to use them:

1. **Find a method/property**: Query codetree.json
   ` + "```bash\n" + `   cat .xojo_index/codetree.json | grep -A 20 "MainWindow.xojo_window"
   ` + "```\n" + `   Returns: "Button1.Pressed": 78 → method at line 78

2. **Read only what's needed**:
   ` + "```bash\n" + `   sed -n '70,90p' AppSrc/MainWindow.xojo_window
   ` + "```\n" + `   Skip scanning the entire 3000-line file

3. **Understand relationships**: Query dependencies.json
   ` + "```bash\n" + `   cat .xojo_index/dependencies.json | grep -A 5 "OrderForm"
   ` + "```\n" + `   See what OrderForm inherits from and which classes depend on it

### Why this saves tokens:
- Instead of reading entire 3KB–50KB files, read only 20–50 lines
- Instead of blindly scanning all classes, query the dependency graph
- Instead of re-parsing file structure, use pre-computed line numbers
- **Result**: 5–8× fewer tokens per task

`

	// Append to content
	content += "\n" + xojiInstructions

	// Write back to CLAUDE.md
	if err := os.WriteFile(claudeMdPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Printf("✓ Updated %s with xoji instructions\n", claudeMdPath)

	// Check if .claude/hooks directory exists and create if needed
	hooksDir := filepath.Join(projectDir, ".claude", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			fmt.Printf("⚠ Could not create .claude/hooks directory: %v\n", err)
			return nil
		}
		fmt.Printf("✓ Created .claude/hooks directory\n")
	}

	// Copy pre-task.sh hook if it doesn't exist
	hookPath := filepath.Join(hooksDir, "pre-task.sh")
	if _, err := os.Stat(hookPath); err != nil {
		hookScript := `#!/bin/bash
# xoji pre-task hook: Auto-refresh indexes before each Claude Code task
set -e
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
cd "$PROJECT_ROOT"
XOJI="${PROJECT_ROOT}/xoji"
if [ ! -f "$XOJI" ]; then
    XOJI=$(command -v xoji || echo "")
    if [ -z "$XOJI" ]; then
        exit 0
    fi
fi
if ! "$XOJI" check 2>/dev/null; then
    echo "[xoji] Index stale, rebuilding..." >&2
    "$XOJI" index >/dev/null 2>&1
    echo "[xoji] Indexes refreshed" >&2
fi
exit 0
`
		if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
			fmt.Printf("⚠ Could not create pre-task hook: %v\n", err)
			return nil
		}
		fmt.Printf("✓ Created .claude/hooks/pre-task.sh\n")
	} else {
		fmt.Printf("✓ .claude/hooks/pre-task.sh already exists\n")
	}

	fmt.Println("\nSetup complete! Your project is now xoji-enabled.")
	fmt.Println("Next steps:")
	fmt.Println("  1. Run: xoji index")
	fmt.Println("  2. Read CLAUDE.md for integration instructions")
	fmt.Println("  3. AI agents will automatically use the indexes")

	return nil
}
