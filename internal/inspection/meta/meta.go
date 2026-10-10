// Package meta holds the catalogue of rule facts: identifiers, groups,
// default severities and option schemas. Upstream-modelled facts are generated
// by tools/extract; native facts are authored independently. Both catalogues
// contain facts only — never descriptions or messages.
package meta

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"custos/internal/diagnostic"
)

// Option describes one configurable rule option.
type Option struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // bool | int | string | enum | list
	Default any    `json:"default,omitempty"`
}

// Rule is the fact sheet of one rule.
type Rule struct {
	Native           bool                `json:"-"` // authored for custos; no upstream counterpart
	ID               string              `json:"id"`
	LegacyID         string              `json:"legacyId"` // PhpStorm short name, honoured by @noinspection
	Group            string              `json:"group"`
	Severity         diagnostic.Severity `json:"severity"`
	EnabledByDefault bool                `json:"enabledByDefault"`
	HasFix           bool                `json:"hasFix"`
	Experimental     bool                `json:"experimental,omitempty"` // disabled upstream; off by default
	Options          []Option            `json:"options,omitempty"`
}

//go:embed rules.json
var rulesJSON []byte

//go:embed native-rules.json
var nativeRulesJSON []byte

var (
	loadOnce sync.Once
	rules    []Rule
	byID     map[string]*Rule
	loadErr  error
)

func load() {
	rules, byID, loadErr = decodeCatalogues(rulesJSON, nativeRulesJSON)
}

func decodeCatalogues(extracted, native []byte) ([]Rule, map[string]*Rule, error) {
	var combined []Rule
	for i, raw := range [][]byte{extracted, native} {
		var entries []Rule
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, nil, fmt.Errorf("meta: decode catalogue %d: %w", i, err)
		}
		for j := range entries {
			entries[j].Native = i == 1
		}
		combined = append(combined, entries...)
	}
	sort.Slice(combined, func(i, j int) bool { return combined[i].ID < combined[j].ID })
	lookup := make(map[string]*Rule, len(combined)*2)
	for i := range combined {
		r := &combined[i]
		if r.ID == "" {
			return nil, nil, fmt.Errorf("meta: empty rule ID")
		}
		for _, id := range []string{r.ID, r.LegacyID} {
			if id == "" {
				continue
			}
			if _, exists := lookup[id]; exists {
				return nil, nil, fmt.Errorf("meta: duplicate rule ID or alias %q", id)
			}
			lookup[id] = r
		}
	}
	return combined, lookup, nil
}

// All returns every rule, sorted by ID.
func All() ([]Rule, error) {
	loadOnce.Do(load)
	return rules, loadErr
}

// Lookup finds a rule by custos ID or legacy PhpStorm short name.
func Lookup(id string) (*Rule, bool) {
	loadOnce.Do(load)
	r, ok := byID[id]
	return r, ok
}
