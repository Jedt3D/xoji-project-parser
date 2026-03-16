package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindXojoProject walks up from the given directory (or current dir) to find a .xojo_project file
func FindXojoProject(startPath string) (string, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	current := startPath
	for {
		// List files in current directory
		entries, err := os.ReadDir(current)
		if err != nil {
			return "", err
		}

		// Look for .xojo_project file
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xojo_project") {
				return filepath.Join(current, entry.Name()), nil
			}
		}

		// Walk up to parent
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			return "", fmt.Errorf("no .xojo_project file found")
		}
		current = parent
	}
}

// ParseXojoProject reads and parses a .xojo_project file
func ParseXojoProject(filePath string) (*Project, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open .xojo_project: %w", err)
	}
	defer file.Close()

	rootPath := filepath.Dir(filePath)
	project := &Project{
		RootPath:  rootPath,
		Items:     []ProjectItem{},
		IndexPath: filepath.Join(rootPath, ".xojo_index"),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and header lines (Type=, RBProjectVersion=, etc.)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		// Parse lines with format: ItemType=Name;RelativePath;&hID;&hParentID;bool
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		itemType := parts[0]
		rest := parts[1]

		// Skip known non-file items
		if itemType == "Type" || itemType == "RBProjectVersion" || itemType == "MinIDEVersion" ||
			itemType == "OrigIDEVersion" || itemType == "BuildSteps" {
			continue
		}

		// Parse the rest: Name;RelativePath;&hID;&hParentID;bool
		semicolons := strings.Split(rest, ";")
		if len(semicolons) != 5 {
			// Invalid line format
			continue
		}

		name := semicolons[0]
		relPath := semicolons[1]
		id := semicolons[2]
		parentID := semicolons[3]
		// bool at semicolons[4] is ignored

		// Skip Folder and Library items (they have no file to parse)
		if itemType == "Folder" || itemType == "Library" {
			continue
		}

		// Skip items with no path (e.g., BuildSteps)
		if relPath == "" {
			continue
		}

		// Skip items without file extensions (directories)
		if !strings.Contains(relPath, ".") {
			continue
		}

		// Add to items list
		project.Items = append(project.Items, ProjectItem{
			ItemType:     itemType,
			DisplayName:  name,
			RelativePath: relPath,
			ID:           id,
			ParentID:     parentID,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return project, nil
}
