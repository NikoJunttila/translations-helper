package jsontools

// Difference represents the differences between two JSON structures
type Difference struct {
	MissingKeys     []string          // Keys present in base but missing in target
	ExtraKeys       []string          // Keys present in target but not in base
	DifferentValues map[string]Values // Keys with different values
}

// Values holds the base and target values for comparison
type Values struct {
	Base   string
	Target string
}

// CompareJSON compares two flattened JSON maps and returns differences
func CompareJSON(base, target map[string]string) Difference {
	diff := Difference{
		MissingKeys:     []string{},
		ExtraKeys:       []string{},
		DifferentValues: make(map[string]Values),
	}

	// Find missing keys (in base but not in target, or empty in target)
	for key := range base {
		val, exists := target[key]
		if !exists || val == "" {
			diff.MissingKeys = append(diff.MissingKeys, key)
		}
	}

	// Find extra keys (in target but not in base)
	for key := range target {
		if _, exists := base[key]; !exists {
			diff.ExtraKeys = append(diff.ExtraKeys, key)
		}
	}

	// Find different values
	for key, baseValue := range base {
		if targetValue, exists := target[key]; exists {
			if baseValue != targetValue {
				diff.DifferentValues[key] = Values{
					Base:   baseValue,
					Target: targetValue,
				}
			}
		}
	}

	return diff
}

// HasMissingTranslations checks if there are any missing keys
func (d Difference) HasMissingTranslations() bool {
	return len(d.MissingKeys) > 0
}

// TotalDifferences returns the total number of differences
func (d Difference) TotalDifferences() int {
	return len(d.MissingKeys) + len(d.ExtraKeys) + len(d.DifferentValues)
}

// CompletionPercentage calculates the completion percentage
func CompletionPercentage(base, target map[string]string) float64 {
	if len(base) == 0 {
		return 100.0
	}

	diff := CompareJSON(base, target)
	missing := len(diff.MissingKeys)
	total := len(base)

	completed := total - missing
	return (float64(completed) / float64(total)) * 100.0
}
