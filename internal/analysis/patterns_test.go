package analysis

import (
	"reflect"
	"testing"

	"custos/internal/syntax"
)

type patternRule struct{ id string }

func (r patternRule) ID() string                { return r.id }
func (patternRule) Kinds() []syntax.NodeKind    { return nil }
func (patternRule) Check(*Context, syntax.Node) {}
func (patternRule) CheckFile(*Context)          {}
func (patternRule) FilePatterns() []string      { return []string{"composer.json"} }

func TestFilePatterns(t *testing.T) {
	e, err := NewEngine([]Rule{patternRule{"SecurityAdvisories"}}, Config{EnableAll: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := e.FilePatterns(); !reflect.DeepEqual(got, []string{"composer.json"}) {
		t.Fatalf("got %v", got)
	}
	off := false
	e, _ = NewEngine([]Rule{patternRule{"SecurityAdvisories"}}, Config{Rules: map[string]RuleConfig{"SecurityAdvisories": {Enabled: &off}}})
	if got := e.FilePatterns(); len(got) != 0 {
		t.Fatalf("disabled rule must not add patterns, got %v", got)
	}
}
