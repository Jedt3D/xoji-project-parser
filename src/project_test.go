package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseXojoProject(t *testing.T) {
	// Create a temporary directory and project file
	tmpDir, err := os.MkdirTemp("", "test_project_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectFile := filepath.Join(tmpDir, "TestProject.xojo_project")

	projectContent := `Type=Desktop
RBProjectVersion=2025.031
MinIDEVersion=20210300
OrigIDEVersion=00000000
Class=App;App.xojo_code;&h000000001BA8AB47;&h0000000000000000;false
Folder=AppSrc;AppSrc;&h00000000352B4FFF;&h0000000000000000;false
Class=MainWindow;AppSrc/MainWindow.xojo_window;&h000000005E851226;&h00000000352B4FFF;false
Module=Utils;AppSrc/Utils.xojo_code;&h0000000076B00FFF;&h00000000352B4FFF;false
Interface=IClosable;AppSrc/IClosable.xojo_code;&hFFFFFFFFA0010FFF;&h00000000352B4FFF;false
Folder=Images;Images;&h000000003ABEFFFF;&h0000000000000000;false
MultiImage=Icon;Images/Icon.xojo_image;&h000000001AF1D7FF;&h000000003ABEFFFF;false
Library=SharedLib;SharedLib;&h000000002560F7FF;&h0000000000000000;false
MenuBar=MainMenu;MainMenu.xojo_menu;&h00000000726AEEEA;&h0000000000000000;false
DesktopToolbar=MainToolbar;MainToolbar.xojo_toolbar;&h000000004C5F6799;&h0000000000000000;false`

	err = os.WriteFile(projectFile, []byte(projectContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write project file: %v", err)
	}

	// Parse the project
	project, err := ParseXojoProject(projectFile)
	if err != nil {
		t.Fatalf("Failed to parse project: %v", err)
	}

	// Verify basic properties
	if project.RootPath != tmpDir {
		t.Errorf("Expected RootPath %s, got %s", tmpDir, project.RootPath)
	}

	expectedIndexPath := filepath.Join(tmpDir, ".xojo_index")
	if project.IndexPath != expectedIndexPath {
		t.Errorf("Expected IndexPath %s, got %s", expectedIndexPath, project.IndexPath)
	}

	// Verify items were parsed
	if len(project.Items) == 0 {
		t.Fatal("Expected at least one project item, got zero")
	}

	// Should have 5 file items (folders and libraries are skipped)
	// App.xojo_code, MainWindow.xojo_window, Utils.xojo_code, IClosable.xojo_code, Icon.xojo_image, MainMenu.xojo_menu, MainToolbar.xojo_toolbar
	expectedCount := 7
	if len(project.Items) != expectedCount {
		t.Errorf("Expected %d items, got %d", expectedCount, len(project.Items))
	}

	// Verify specific items
	itemMap := make(map[string]*ProjectItem)
	for i := range project.Items {
		itemMap[project.Items[i].RelativePath] = &project.Items[i]
	}

	// Check App.xojo_code
	appItem, ok := itemMap["App.xojo_code"]
	if !ok {
		t.Error("Expected to find App.xojo_code")
	} else {
		if appItem.ItemType != "Class" {
			t.Errorf("App.xojo_code should be Class, got %s", appItem.ItemType)
		}
		if appItem.DisplayName != "App" {
			t.Errorf("App.xojo_code DisplayName should be App, got %s", appItem.DisplayName)
		}
	}

	// Check MainWindow (should handle both .xojo_window extensions)
	mainWindowItem, ok := itemMap["AppSrc/MainWindow.xojo_window"]
	if !ok {
		t.Error("Expected to find AppSrc/MainWindow.xojo_window")
	} else {
		if mainWindowItem.ItemType != "Class" {
			t.Errorf("MainWindow should be Class type, got %s", mainWindowItem.ItemType)
		}
	}

	// Verify Folders are skipped
	_, ok = itemMap["AppSrc"]
	if ok {
		t.Error("Folder items should be skipped")
	}

	// Verify Library items are skipped
	_, ok = itemMap["SharedLib"]
	if ok {
		t.Error("Library items should be skipped")
	}
}

func TestParseXojoProjectNonexistent(t *testing.T) {
	_, err := ParseXojoProject("/nonexistent/path/to/project.xojo_project")
	if err == nil {
		t.Error("Should return error for nonexistent project file")
	}
}

func TestParseXojoProjectEmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_empty_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectFile := filepath.Join(tmpDir, "Empty.xojo_project")
	os.WriteFile(projectFile, []byte(""), 0644)

	project, err := ParseXojoProject(projectFile)
	if err != nil {
		t.Fatalf("Failed to parse empty project: %v", err)
	}

	if len(project.Items) != 0 {
		t.Errorf("Expected 0 items in empty project, got %d", len(project.Items))
	}
}

func TestParseXojoProjectRelativePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_relative_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectFile := filepath.Join(tmpDir, "TestProject.xojo_project")

	projectContent := `Type=Desktop
Class=App;App.xojo_code;&h0000000001;&h0000000000;false
Class=Form;Subfolder/Form.xojo_code;&h0000000002;&h0000000000;false`

	os.WriteFile(projectFile, []byte(projectContent), 0644)

	project, err := ParseXojoProject(projectFile)
	if err != nil {
		t.Fatalf("Failed to parse project: %v", err)
	}

	itemMap := make(map[string]*ProjectItem)
	for i := range project.Items {
		itemMap[project.Items[i].RelativePath] = &project.Items[i]
	}

	if _, ok := itemMap["App.xojo_code"]; !ok {
		t.Error("Expected App.xojo_code")
	}

	if _, ok := itemMap["Subfolder/Form.xojo_code"]; !ok {
		t.Error("Expected Subfolder/Form.xojo_code")
	}
}

func TestFindXojoProject(t *testing.T) {
	// Create a nested directory structure
	tmpDir, err := os.MkdirTemp("", "test_find_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	nestedDir := filepath.Join(tmpDir, "nested", "deep", "directory")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create project file in the root temp directory
	projectFile := filepath.Join(tmpDir, "MyProject.xojo_project")
	os.WriteFile(projectFile, []byte("Type=Desktop"), 0644)

	// Search from nested directory
	found, err := FindXojoProject(nestedDir)
	if err != nil {
		t.Fatalf("Failed to find project: %v", err)
	}

	if found != projectFile {
		t.Errorf("Expected to find %s, got %s", projectFile, found)
	}
}

func TestFindXojoProjectNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_not_found_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = FindXojoProject(tmpDir)
	if err == nil {
		t.Error("Should return error when project not found")
	}
}

func TestFindXojoProjectMultiple(t *testing.T) {
	// Create a directory with project file
	tmpDir, err := os.MkdirTemp("", "test_multiple_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	projectFile := filepath.Join(tmpDir, "MyProject.xojo_project")
	os.WriteFile(projectFile, []byte("Type=Desktop"), 0644)

	// Search from the directory containing the project
	found, err := FindXojoProject(tmpDir)
	if err != nil {
		t.Fatalf("Failed to find project: %v", err)
	}

	if found != projectFile {
		t.Errorf("Expected to find %s, got %s", projectFile, found)
	}
}
