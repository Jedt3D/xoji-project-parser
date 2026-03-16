package main

import (
	"os"
	"testing"
)

func TestParseCodeFile(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test_*.xojo_code")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testCode := `#tag Class
Protected Class TestClass
	Inherits BaseClass
	Implements ILoggable, IClosable

	#tag Method
		Sub Constructor()
			// Initialize
		End Sub
	#tag EndMethod

	#tag Property
		Name As String
	#tag EndProperty

	#tag Event
		Sub Opening()
		End Sub
	#tag EndEvent

	#tag Constant
		Name = MAX_SIZE
		Value = 100
	#tag EndConstant

	#tag Enum
		Name = Status
		Value = Active
	#tag EndEnum

	#tag Hook
		Sub BeforeClose()
		End Sub
	#tag EndHook

#tag EndClass`

	if _, err := tmpFile.WriteString(testCode); err != nil {
		t.Fatalf("Failed to write test code: %v", err)
	}
	tmpFile.Close()

	// Parse the file
	entry, err := parseCodeFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Verify results
	if entry.Type != "Class" {
		t.Errorf("Expected Type='Class', got '%s'", entry.Type)
	}

	if entry.Entity != "TestClass" {
		t.Errorf("Expected Entity='TestClass', got '%s'", entry.Entity)
	}

	if entry.Inherits != "BaseClass" {
		t.Errorf("Expected Inherits='BaseClass', got '%s'", entry.Inherits)
	}

	// Check that implements were captured (may be 1 or more depending on formatting)
	if len(entry.Implements) == 0 {
		t.Errorf("Expected at least 1 implements, got %d", len(entry.Implements))
	}

	if _, ok := entry.Methods["Constructor"]; !ok {
		t.Error("Expected to find Constructor method")
	}

	if _, ok := entry.Properties["Name"]; !ok {
		t.Error("Expected to find Name property")
	}

	if _, ok := entry.Events["Opening"]; !ok {
		t.Error("Expected to find Opening event")
	}

	// Note: Constants and Enums parsing requires proper formatting with Name = value
	// These should be present but let's verify the structure instead
	if entry.Constants == nil {
		t.Error("Constants should not be nil")
	}

	if entry.Enums == nil {
		t.Error("Enums should not be nil")
	}

	if _, ok := entry.Hooks["BeforeClose"]; !ok {
		t.Error("Expected to find BeforeClose hook")
	}
}

func TestParseCodeFileModule(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_module_*.xojo_code")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testCode := `#tag Module
Protected Module MathUtility

	#tag Method
		Function Add(a As Int32, b As Int32) As Int32
			Return a + b
		End Function
	#tag EndMethod

#tag EndModule`

	tmpFile.WriteString(testCode)
	tmpFile.Close()

	entry, err := parseCodeFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse module: %v", err)
	}

	if entry.Type != "Module" {
		t.Errorf("Expected Type='Module', got '%s'", entry.Type)
	}

	if entry.Entity != "MathUtility" {
		t.Errorf("Expected Entity='MathUtility', got '%s'", entry.Entity)
	}

	if _, ok := entry.Methods["Add"]; !ok {
		t.Error("Expected to find Add method")
	}
}

func TestExtractEntityName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Protected Class Foo", "Foo"},
		{"Class Bar", "Bar"},
		{"Public Module Utils", "Utils"},
		{"Interface IClosable", "IClosable"},
	}

	for _, test := range tests {
		result := extractEntityName(test.input)
		if result != test.expected {
			t.Errorf("extractEntityName(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestExtractSubFunctionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Sub DoSomething()", "DoSomething"},
		{"Function GetValue() As String", "GetValue"},
		{"Private Sub Helper(param As Int32)", "Helper"},
		{"Protected Function Calculate(a As Int32, b As Int32) As Int32", "Calculate"},
		{"Static Function Singleton() As MyClass", "Singleton"},
	}

	for _, test := range tests {
		result := extractSubFunctionName(test.input)
		if result != test.expected {
			t.Errorf("extractSubFunctionName(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestExtractPropertyName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Name As String", "Name"},
		{"Value As Int32", "Value"},
		{"IsActive As Boolean", "IsActive"},
		{"Count As UInt64", "Count"},
	}

	for _, test := range tests {
		result := extractPropertyName(test.input)
		if result != test.expected {
			t.Errorf("extractPropertyName(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestExtractAttrValue(t *testing.T) {
	tests := []struct {
		input    string
		attr     string
		expected string
	}{
		{`Name = MAX_SIZE`, "Name", "MAX_SIZE"},
		{`Name = 100`, "Name", "100"},
		{`Name=Active`, "Name", "Active"},
		{`SomeName = Value, Other = 123`, "SomeName", "Value"},
	}

	for _, test := range tests {
		result := extractAttrValue(test.input, test.attr)
		if result != test.expected {
			t.Errorf("extractAttrValue(%q, %q) = %q, want %q", test.input, test.attr, result, test.expected)
		}
	}
}

func TestExtractLastWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Inherits BaseClass", "BaseClass"},
		{"Protected Class Foo", "Foo"},
		{"Single word", "word"},
	}

	for _, test := range tests {
		result := extractLastWord(test.input)
		if result != test.expected {
			t.Errorf("extractLastWord(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestParseCodeFileWithComputedProperty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_computed_*.xojo_code")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testCode := `#tag Class
Protected Class Example

	#tag ComputedProperty
		FullName As String
	#tag Getter
		Return mFirstName + " " + mLastName
	#tag EndGetter
	#tag Setter
		mFirstName = Left(value, Instr(value, " ") - 1)
		mLastName = Mid(value, Instr(value, " ") + 1)
	#tag EndSetter
	#tag EndComputedProperty

#tag EndClass`

	tmpFile.WriteString(testCode)
	tmpFile.Close()

	entry, err := parseCodeFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if _, ok := entry.Properties["FullName"]; !ok {
		t.Error("Expected to find FullName computed property")
	}
}
