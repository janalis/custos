package types

import "testing"

// Refined pseudo-types and int<lo, hi> keep their values; a malformed
// range is a plain int.
func TestRefinedDocParts(t *testing.T) {
	for doc, want := range map[string]string{
		"positive-int":          "int<1, max>",
		"negative-int":          "int<min, -1>",
		"non-positive-int":      "int<min, 0>",
		"non-negative-int":      "int<0, max>",
		"non-zero-int":          "int<min, -1>|int<1, max>",
		"non-empty-string":      "string<len 1, max>",
		"non-falsy-string":      "string<len 1, max>",
		"truthy-string":         "string<len 1, max>",
		"numeric-string":        "string<len 1, max>",
		"literal-string":        "string",
		"int<-5, MAX>":          "int<-5, max>",
		"int<min, 0>":           "int<min, 0>",
		"int<1>":                "int",
		"int<1, 2, 3>":          "int",
		"int<2, 1>":             "int",
		"int<a, 1>":             "int",
		"int<1, b>":             "int",
		"int<1, 2>[]":           "int[]<int<1, 2>>",
		"positive-int|int":      "int",
		"int<1, 2>|int<5, 6>":   "int<1, 2>|int<5, 6>",
		"positive-int|null":     "int<1, max>|null",
		"non-empty-string|null": "null|string<len 1, max>",
	} {
		if got := FromDoc(doc, nil).ShapeString(); got != want {
			t.Errorf("%s: got %s, want %s", doc, got, want)
		}
	}
}
