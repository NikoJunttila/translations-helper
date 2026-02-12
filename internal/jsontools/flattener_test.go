package jsontools

import (
	"reflect"
	"testing"
)

func TestFlattenJSON(t *testing.T) {
	input := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
		"active": true,
	}

	expected := map[string]string{
		"user.name": "John",
		"user.age":  "30",
		"active":    "true",
	}

	result := FlattenJSON(input, "")

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("FlattenJSON() = %+v, want %+v", result, expected)
	}
}

func TestUnflattenJSON(t *testing.T) {
	input := map[string]string{
		"user.name": "John",
		"user.age":  "30",
		"active":    "true",
	}

	// We expect numeric values to come back as strings/float64 depending on how they were parsed originally.
	// But UnflattenJSON logic puts them into map[string]interface{}.
	// Let's verify structure.

	result := UnflattenJSON(input)

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected user to be a map, got %T", result["user"])
	}

	if user["name"] != "John" {
		t.Errorf("expected user.name to be John, got %v", user["name"])
	}
	if user["age"] != "30" {
		t.Errorf("expected user.age to be 30, got %v", user["age"])
	}
	if result["active"] != "true" {
		t.Errorf("expected active to be true, got %v", result["active"])
	}
}

func TestFlattenUnflattenRoundTrip(t *testing.T) {
	input := map[string]interface{}{
		"param": "value",
		"nested": map[string]interface{}{
			"key": "val",
		},
	}

	flat := FlattenJSON(input, "")
	unflat := UnflattenJSON(flat)

	// Note: Round trip might change types (int -> string) because Flatten stores everything as string.
	// We verify structure and string values.

	if unflat["param"] != "value" {
		t.Errorf("expected param=value, got %v", unflat["param"])
	}

	nested, ok := unflat["nested"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested to be map, got %T", unflat["nested"])
	}

	if nested["key"] != "val" {
		t.Errorf("expected nested.key=val, got %v", nested["key"])
	}
}
