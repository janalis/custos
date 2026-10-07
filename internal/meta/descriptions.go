package meta

import (
	_ "embed"
	"encoding/json"
	"sync"
)

// Description is the human documentation of a rule (from its spec).
type Description struct {
	Summary string `json:"summary"`
	Options string `json:"options,omitempty"`
	Fix     bool   `json:"fix,omitempty"` // custos offers a quick-fix
}

//go:embed descriptions.json
var descriptionsJSON []byte

var (
	descOnce sync.Once
	descs    map[string]Description
)

// Describe returns the documentation of a rule by ID.
func Describe(id string) (Description, bool) {
	descOnce.Do(func() { _ = json.Unmarshal(descriptionsJSON, &descs) })
	d, ok := descs[id]
	return d, ok
}
