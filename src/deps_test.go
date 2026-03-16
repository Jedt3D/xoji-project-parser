package main

import (
	"encoding/json"
	"testing"
)

func TestBuildDependencies(t *testing.T) {
	// Create test data
	codeTree := make(CodeTree)

	// Add some test entries
	codeEntry1 := &CodeEntry{
		Type:       "Class",
		Entity:     "MainForm",
		Inherits:   "BaseForm",
		Implements: []string{"IClosable", "IRefreshable"},
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
		Constants:  make(map[string]int),
		Enums:      make(map[string]int),
		Hooks:      make(map[string]int),
	}

	codeEntry2 := &CodeEntry{
		Type:        "Class",
		Entity:      "SettingsForm",
		Inherits:    "BaseForm",
		Implements:  []string{"IClosable"},
		Methods:     make(map[string]int),
		Properties:  make(map[string]int),
		Events:      make(map[string]int),
		Constants:   make(map[string]int),
		Enums:       make(map[string]int),
		Hooks:       make(map[string]int),
	}

	codeEntry3 := &CodeEntry{
		Type:        "Class",
		Entity:      "BaseForm",
		Inherits:    "",
		Implements:  []string{},
		Methods:     make(map[string]int),
		Properties:  make(map[string]int),
		Events:      make(map[string]int),
		Constants:   make(map[string]int),
		Enums:       make(map[string]int),
		Hooks:       make(map[string]int),
	}

	codeTree["MainForm.xojo_code"] = codeEntry1
	codeTree["SettingsForm.xojo_code"] = codeEntry2
	codeTree["BaseForm.xojo_code"] = codeEntry3

	// Build dependencies
	deps := buildDependencies(codeTree)

	// Verify classes map
	if len(deps.Classes) != 3 {
		t.Errorf("Expected 3 classes in dependency graph, got %d", len(deps.Classes))
	}

	// Check MainForm
	mainFormDeps, ok := deps.Classes["MainForm"]
	if !ok {
		t.Error("Expected MainForm in classes")
	}
	if mainFormDeps.Inherits != "BaseForm" {
		t.Errorf("MainForm should inherit from BaseForm, got %s", mainFormDeps.Inherits)
	}
	if len(mainFormDeps.Implements) != 2 {
		t.Errorf("MainForm should implement 2 interfaces, got %d", len(mainFormDeps.Implements))
	}

	// Check reverse map (inheritance)
	baseFormDerived, ok := deps.Reverse["BaseForm"]
	if !ok {
		t.Error("Expected BaseForm in reverse map")
	}
	if len(baseFormDerived) != 2 {
		t.Errorf("Expected 2 classes to inherit from BaseForm, got %d", len(baseFormDerived))
	}

	// Verify the derived classes
	derivedSet := make(map[string]bool)
	for _, derived := range baseFormDerived {
		derivedSet[derived] = true
	}
	if !derivedSet["MainForm"] {
		t.Error("Expected MainForm to be in BaseForm's reverse map")
	}
	if !derivedSet["SettingsForm"] {
		t.Error("Expected SettingsForm to be in BaseForm's reverse map")
	}
}

func TestBuildDependenciesWithInterfaces(t *testing.T) {
	codeTree := make(CodeTree)

	codeEntry := &CodeEntry{
		Type:       "Class",
		Entity:     "MyClass",
		Inherits:   "",
		Implements: []string{"IClosable", "IRefreshable", "IDisposable"},
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
		Constants:  make(map[string]int),
		Enums:      make(map[string]int),
		Hooks:      make(map[string]int),
	}

	codeTree["MyClass.xojo_code"] = codeEntry

	deps := buildDependencies(codeTree)

	myClassDeps := deps.Classes["MyClass"]
	if len(myClassDeps.Implements) != 3 {
		t.Errorf("Expected 3 implemented interfaces, got %d", len(myClassDeps.Implements))
	}
}

func TestBuildDependenciesNoInheritance(t *testing.T) {
	codeTree := make(CodeTree)

	codeEntry := &CodeEntry{
		Type:       "Class",
		Entity:     "Standalone",
		Inherits:   "",
		Implements: []string{},
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
		Constants:  make(map[string]int),
		Enums:      make(map[string]int),
		Hooks:      make(map[string]int),
	}

	codeTree["Standalone.xojo_code"] = codeEntry

	deps := buildDependencies(codeTree)

	if len(deps.Reverse) != 0 {
		t.Errorf("Expected empty reverse map for standalone class, got %d entries", len(deps.Reverse))
	}
}

func TestBuildDependenciesJSON(t *testing.T) {
	codeTree := make(CodeTree)

	codeEntry := &CodeEntry{
		Type:       "Class",
		Entity:     "MyClass",
		Inherits:   "BaseClass",
		Implements: []string{"IInterface"},
		Methods:    make(map[string]int),
		Properties: make(map[string]int),
		Events:     make(map[string]int),
		Constants:  make(map[string]int),
		Enums:      make(map[string]int),
		Hooks:      make(map[string]int),
	}

	codeTree["MyClass.xojo_code"] = codeEntry

	deps := buildDependencies(codeTree)

	// Verify it can be marshaled to JSON
	jsonBytes, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal dependencies to JSON: %v", err)
	}

	// Verify it can be unmarshaled
	var unmarshaled DependencyGraph
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal dependencies from JSON: %v", err)
	}

	if len(unmarshaled.Classes) != 1 {
		t.Errorf("Expected 1 class after unmarshal, got %d", len(unmarshaled.Classes))
	}
}
