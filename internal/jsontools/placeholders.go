package jsontools

import (
	"fmt"
	"regexp"
	"sort"
)

var placeholderRegex = regexp.MustCompile(`\{[^}]+\}`)

// ExtractPlaceholders returns a sorted list of unique placeholders found in the text
func ExtractPlaceholders(text string) []string {
	matches := placeholderRegex.FindAllString(text, -1)
	if matches == nil {
		return []string{}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			unique = append(unique, match)
		}
	}
	sort.Strings(unique)
	return unique
}

// ValidatePlaceholders checks if target contains all placeholders from base
func ValidatePlaceholders(base, target string) error {
	basePlaceholders := ExtractPlaceholders(base)
	targetPlaceholders := ExtractPlaceholders(target)

	// Create a set for target placeholders for O(1) lookup
	targetSet := make(map[string]bool)
	for _, p := range targetPlaceholders {
		targetSet[p] = true
	}

	// Check if all base placeholders exist in target
	for _, p := range basePlaceholders {
		if !targetSet[p] {
			return fmt.Errorf("missing required placeholder: %s", p)
		}
	}

	return nil
}
