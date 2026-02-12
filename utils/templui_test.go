package utils

import (
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestTwMerge(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "single class",
			input:    []string{"bg-red-500"},
			expected: "bg-red-500",
		},
		{
			name:     "conflict resolution",
			input:    []string{"bg-red-500", "bg-blue-500"},
			expected: "bg-blue-500",
		},
		{
			name:     "mixed classes",
			input:    []string{"p-4", "p-2"},
			expected: "p-2",
		},
		{
			name:     "preserving non-conflicting",
			input:    []string{"text-white", "bg-red-500"},
			expected: "text-white bg-red-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TwMerge(tt.input...)
			if got != tt.expected {
				t.Errorf("TwMerge() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIf(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		value     string
		expected  string
	}{
		{
			name:      "true condition",
			condition: true,
			value:     "active",
			expected:  "active",
		},
		{
			name:      "false condition",
			condition: false,
			value:     "active",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := If(tt.condition, tt.value)
			if got != tt.expected {
				t.Errorf("If() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIfElse(t *testing.T) {
	tests := []struct {
		name       string
		condition  bool
		trueValue  string
		falseValue string
		expected   string
	}{
		{
			name:       "true condition",
			condition:  true,
			trueValue:  "yes",
			falseValue: "no",
			expected:   "yes",
		},
		{
			name:       "false condition",
			condition:  false,
			trueValue:  "yes",
			falseValue: "no",
			expected:   "no",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IfElse(tt.condition, tt.trueValue, tt.falseValue)
			if got != tt.expected {
				t.Errorf("IfElse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMergeAttributes(t *testing.T) {
	attr1 := templ.Attributes{"class": "foo", "id": "bar"}
	attr2 := templ.Attributes{"class": "baz", "data-testid": "test"}

	merged := MergeAttributes(attr1, attr2)

	if merged["id"] != "bar" {
		t.Errorf("expected id=bar, got %v", merged["id"])
	}
	if merged["class"] != "baz" {
		t.Errorf("expected class=baz, got %v", merged["class"])
	}
	if merged["data-testid"] != "test" {
		t.Errorf("expected data-testid=test, got %v", merged["data-testid"])
	}
}

func TestRandomID(t *testing.T) {
	id1 := RandomID()
	id2 := RandomID()

	if !strings.HasPrefix(id1, "id-") {
		t.Errorf("expected ID directly start with 'id-', got %v", id1)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, got duplicate %v", id1)
	}
}

func TestScriptURL(t *testing.T) {
	// Test default behavior
	path := "/assets/js/main.js"
	url := ScriptURL(path)
	
	if !strings.Contains(url, path) {
		t.Errorf("expected URL to contain path, got %v", url)
	}
	if !strings.Contains(url, "?v=") {
		t.Errorf("expected URL to contain version parameter, got %v", url)
	}
}
