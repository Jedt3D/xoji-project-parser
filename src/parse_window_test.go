package main

import (
	"os"
	"testing"
)

func TestParseWindowFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.xojo_window")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testWindow := `#tag DesktopWindow
Begin DesktopWindow MainWindow
	Height          =   600
	Width           =   800
	Title           =   "Main Application"

	Begin DesktopButton Button1
		Caption         =   "Click Me"
		Left            =   10
		Top             =   10
		Width           =   100
		Height          =   40
	End

	Begin DesktopTextfield TextField1
		Left            =   10
		Top             =   60
		Width           =   200
		Height          =   25
	End

	Begin DesktopListBox ListBox1
		Left            =   10
		Top             =   100
		Width           =   300
		Height          =   400
	End

End

#tag WindowCode
	#tag Method
		Sub LoadData()
			// Load data
		End Sub
	#tag EndMethod

	#tag Property
		mData As Variant
	#tag EndProperty

	#tag Event
		Sub Opening()
			LoadData()
		End Sub
	#tag EndEvent
#tag EndWindowCode

#tag Events Button1
	#tag Event
		Sub Pressed()
			MessageBox("Button clicked")
		End Sub
	#tag EndEvent
#tag EndEvents

#tag Events ListBox1
	#tag Event
		Sub Change()
			// Handle change
		End Sub
	#tag EndEvent
#tag EndEvents

#tag ViewBehavior
	/* Comments */
#tag EndViewBehavior

#tag EndDesktopWindow`

	tmpFile.WriteString(testWindow)
	tmpFile.Close()

	entry, err := parseWindowFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse window: %v", err)
	}

	if entry.Type != "DesktopWindow" {
		t.Errorf("Expected Type='DesktopWindow', got '%s'", entry.Type)
	}

	if entry.Entity != "MainWindow" {
		t.Errorf("Expected Entity='MainWindow', got '%s'", entry.Entity)
	}

	// Check controls
	if len(entry.Controls) != 3 {
		t.Errorf("Expected 3 controls, got %d", len(entry.Controls))
	}

	// Verify specific controls
	controlNames := make(map[string]bool)
	for _, ctrl := range entry.Controls {
		controlNames[ctrl.Name] = true
	}

	if !controlNames["Button1"] {
		t.Error("Expected to find Button1 control")
	}
	if !controlNames["TextField1"] {
		t.Error("Expected to find TextField1 control")
	}
	if !controlNames["ListBox1"] {
		t.Error("Expected to find ListBox1 control")
	}

	// Check window-level method
	if _, ok := entry.Methods["LoadData"]; !ok {
		t.Error("Expected to find LoadData method in window")
	}

	// Check window-level property
	if _, ok := entry.Properties["mData"]; !ok {
		t.Error("Expected to find mData property in window")
	}

	// Check window-level event
	if _, ok := entry.Events["Opening"]; !ok {
		t.Error("Expected to find Opening event in window")
	}

	// Check per-control events
	if _, ok := entry.Events["Button1.Pressed"]; !ok {
		t.Error("Expected to find Button1.Pressed event")
	}

	if _, ok := entry.Events["ListBox1.Change"]; !ok {
		t.Error("Expected to find ListBox1.Change event")
	}
}

func TestParseWindowFileMinimal(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_minimal_*.xojo_window")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testWindow := `#tag Window
Begin Window SimpleWindow
	Height = 300
	Width = 400
End
#tag EndWindow`

	tmpFile.WriteString(testWindow)
	tmpFile.Close()

	entry, err := parseWindowFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse window: %v", err)
	}

	if entry.Type != "Window" {
		t.Errorf("Expected Type='Window', got '%s'", entry.Type)
	}

	if entry.Entity != "SimpleWindow" {
		t.Errorf("Expected Entity='SimpleWindow', got '%s'", entry.Entity)
	}

	if len(entry.Controls) != 0 {
		t.Errorf("Expected 0 controls in minimal window, got %d", len(entry.Controls))
	}
}

func TestParseWindowWithNestedControls(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_nested_*.xojo_window")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Window with a GroupBox containing nested controls
	testWindow := `#tag DesktopWindow
Begin DesktopWindow TestWindow

	Begin DesktopGroupBox GroupBox1
		Caption = "Options"

		Begin DesktopCheckBox Checkbox1
			Caption = "Option 1"
		End

		Begin DesktopCheckBox Checkbox2
			Caption = "Option 2"
		End

	End

	Begin DesktopButton Button1
		Caption = "OK"
	End

End
#tag EndDesktopWindow`

	tmpFile.WriteString(testWindow)
	tmpFile.Close()

	entry, err := parseWindowFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse window: %v", err)
	}

	// Should have GroupBox1 and Button1 (depth 1 only)
	if len(entry.Controls) != 2 {
		t.Errorf("Expected 2 top-level controls, got %d", len(entry.Controls))
	}

	controlNames := make(map[string]bool)
	for _, ctrl := range entry.Controls {
		controlNames[ctrl.Name] = true
	}

	if !controlNames["GroupBox1"] {
		t.Error("Expected to find GroupBox1 control")
	}
	if !controlNames["Button1"] {
		t.Error("Expected to find Button1 control")
	}
}
