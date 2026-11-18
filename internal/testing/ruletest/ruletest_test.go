package ruletest

import (
	"fmt"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// recorder is a testing.TB that keeps failures instead of stopping.
type recorder struct {
	testing.TB
	failures []string
}

func (r *recorder) Helper()           {}
func (r *recorder) Fatal(args ...any) { r.failures = append(r.failures, fmt.Sprint(args...)) }

func (r *recorder) Fatalf(format string, a ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, a...))
}

// rule reports every unary operation with message "m".
type rule struct{ id string }

func (r rule) ID() string                               { return r.id }
func (rule) Kinds() []syntax.NodeKind                   { return []syntax.NodeKind{syntax.KUnary} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) { ctx.ReportNode(n, "m") }

func TestCount(t *testing.T) {
	Count(t, rule{"NestedNotOperators"}, "m", []Case{{Src: "$a = !$b;", Want: 1}, {Src: "$a = $b;", Want: 0}})
	r := &recorder{TB: t}
	Count(r, rule{"NestedNotOperators"}, "other", []Case{{Src: "$a = !$b;", Want: 2}, {Src: "$a = !$b;", Want: 1}})
	Count(r, rule{"NoSuchRule"}, "m", nil)
	if len(r.failures) != 3 {
		t.Fatalf("failures: %q", r.failures)
	}
}
