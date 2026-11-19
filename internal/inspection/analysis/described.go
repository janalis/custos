package analysis

import "github.com/janalis/custos/internal/inspection/meta"

// Described is implemented by rules defined outside the embedded catalogue,
// such as the custom rules of a project: they carry their own facts and
// documentation instead of a catalogue entry. Their IDs must not name a
// catalogue rule.
type Described interface {
	Facts() meta.Rule
	Description() meta.Description
}

// Fingerprinted is implemented by rules whose behavior depends on data
// outside the binary (a project's custom rule definitions): Fingerprint
// changes whenever that data does, so cached findings are not reused.
type Fingerprinted interface {
	Fingerprint() string
}

// FactsOf returns the facts of r: its catalogue entry, else the facts it
// carries.
func FactsOf(r Rule) (*meta.Rule, bool) {
	if m, ok := meta.Lookup(r.ID()); ok {
		return m, true
	}
	if d, ok := r.(Described); ok {
		m := d.Facts()
		return &m, true
	}
	return nil, false
}

// Lookup finds the facts of the rule named id, by ID or legacy name: a
// catalogue rule, else one of rules that carries its own facts.
func Lookup(rules []Rule, id string) (*meta.Rule, bool) {
	if m, ok := meta.Lookup(id); ok {
		return m, true
	}
	for _, r := range rules {
		if d, ok := r.(Described); ok && r.ID() == id {
			m := d.Facts()
			return &m, true
		}
	}
	return nil, false
}

// Describe returns the documentation of the rule named id (see Lookup).
func Describe(rules []Rule, id string) (meta.Description, bool) {
	if m, ok := meta.Lookup(id); ok {
		return meta.Describe(m.ID)
	}
	for _, r := range rules {
		if d, ok := r.(Described); ok && r.ID() == id {
			return d.Description(), true
		}
	}
	return meta.Description{}, false
}
