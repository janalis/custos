// Package pattern compiles the custom rules a project declares in its
// configuration: PHP code patterns with metavariables, matched on resolved
// names, with an optional quick-fix template.
package pattern

import (
	"encoding/json"
	"errors"
)

// Definition is the JSON shape of one custom rule.
type Definition struct {
	ID    string           `json:"id"`
	Match Patterns         `json:"match"`
	Where map[string]Where `json:"where,omitempty"`
	// Inside lists code the match must lie within (one must match an
	// enclosing node); NotInside, code it must not lie within.
	Inside      Patterns `json:"inside,omitempty"`
	NotInside   Patterns `json:"notInside,omitempty"`
	Message     string   `json:"message"`
	Severity    string   `json:"severity,omitempty"` // error | warning (default) | info
	Enabled     *bool    `json:"enabled,omitempty"`  // default true
	Description string   `json:"description,omitempty"`
	PHP         string   `json:"php,omitempty"` // minimum target PHP version
	Paths       []string `json:"paths,omitempty"`
	Exclude     []string `json:"exclude,omitempty"`
	Fix         *string  `json:"fix,omitempty"`
	FixTitle    string   `json:"fixTitle,omitempty"`
}

// Where constrains what a metavariable may bind.
type Where struct {
	// Type is a PHPDoc type the bound expression's inferred type must be
	// accepted by, with fully qualified class names.
	Type string `json:"type,omitempty"`
	// Regex must match the bound source text.
	Regex string `json:"regex,omitempty"`
	// Not lists patterns the bound code must not match.
	Not Patterns `json:"not,omitempty"`
}

// Patterns is one pattern or a list of them; JSON accepts both forms.
type Patterns []string

// UnmarshalJSON accepts a string or an array of strings.
func (p *Patterns) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*p = Patterns{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return errors.New("want a pattern or a list of patterns")
	}
	*p = many
	return nil
}
