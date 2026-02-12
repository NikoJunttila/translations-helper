package jsontools

import (
	"reflect"
	"testing"
)

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single placeholder",
			text:     "Hello {name}",
			expected: []string{"{name}"},
		},
		{
			name:     "multiple placeholders",
			text:     "{greeting} {name}",
			expected: []string{"{greeting}", "{name}"},
		},
		{
			name:     "no placeholders",
			text:     "Hello World",
			expected: []string{},
		},
		{
			name:     "duplicates",
			text:     "{a} {b} {a}",
			expected: []string{"{a}", "{b}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPlaceholders(tt.text)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ExtractPlaceholders() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePlaceholders(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		target  string
		wantErr bool
	}{
		{
			name:    "valid",
			base:    "Hello {name}",
			target:  "Hola {name}",
			wantErr: false,
		},
		{
			name:    "missing placeholder in target",
			base:    "Hello {name}",
			target:  "Hola",
			wantErr: true,
		},
		{
			name:    "extra placeholder in target",
			base:    "Hello",
			target:  "Hola {name}",
			wantErr: false, // ExtractPlaceholders checks if target contains all base placeholders, extra is fine usually
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlaceholders(tt.base, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlaceholders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
