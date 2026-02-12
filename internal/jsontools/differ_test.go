package jsontools

import (
	"reflect"
	"sort"
	"testing"
)

func TestCompareJSON(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]string
		target   map[string]string
		expected Difference
	}{
		{
			name:   "identical maps",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{"a": "1", "b": "2"},
			expected: Difference{
				MissingKeys:     []string{},
				ExtraKeys:       []string{},
				DifferentValues: make(map[string]Values),
			},
		},
		{
			name:   "missing keys",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{"a": "1"},
			expected: Difference{
				MissingKeys:     []string{"b"},
				ExtraKeys:       []string{},
				DifferentValues: make(map[string]Values),
			},
		},
		{
			name:   "extra keys",
			base:   map[string]string{"a": "1"},
			target: map[string]string{"a": "1", "b": "2"},
			expected: Difference{
				MissingKeys:     []string{},
				ExtraKeys:       []string{"b"},
				DifferentValues: make(map[string]Values),
			},
		},
		{
			name:   "different values",
			base:   map[string]string{"a": "1"},
			target: map[string]string{"a": "2"},
			expected: Difference{
				MissingKeys: []string{},
				ExtraKeys:   []string{},
				DifferentValues: map[string]Values{
					"a": {Base: "1", Target: "2"},
				},
			},
		},
		{
			name:   "mixed differences",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{"a": "3", "c": "4"},
			expected: Difference{
				MissingKeys: []string{"b"},
				ExtraKeys:   []string{"c"},
				DifferentValues: map[string]Values{
					"a": {Base: "1", Target: "3"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareJSON(tt.base, tt.target)

			// Sort slices for comparison
			sort.Strings(got.MissingKeys)
			sort.Strings(got.ExtraKeys)
			sort.Strings(tt.expected.MissingKeys)
			sort.Strings(tt.expected.ExtraKeys)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("CompareJSON() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestCompletionPercentage(t *testing.T) {
	tests := []struct {
		name   string
		base   map[string]string
		target map[string]string
		want   float64
	}{
		{
			name:   "100% complete",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{"a": "1", "b": "2"},
			want:   100.0,
		},
		{
			name:   "50% complete",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{"a": "1"},
			want:   50.0,
		},
		{
			name:   "0% complete",
			base:   map[string]string{"a": "1", "b": "2"},
			target: map[string]string{},
			want:   0.0,
		},
		{
			name:   "empty base",
			base:   map[string]string{},
			target: map[string]string{"a": "1"},
			want:   100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompletionPercentage(tt.base, tt.target); got != tt.want {
				t.Errorf("CompletionPercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}
