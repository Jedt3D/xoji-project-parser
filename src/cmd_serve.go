package main

import (
	"fmt"
	"os"
)

// CmdServe handles `xoji serve` - watches for file changes and auto-reindexes
func CmdServe(projectPath string) error {
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

	// TODO: implement watch mode (Phase 8)
	return fmt.Errorf("CmdServe: not yet implemented")
}
