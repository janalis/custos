package analysis_test

import (
	"slices"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/dommissingattributecheckedasnull"
	"custos/internal/inspection/rules/impossibletypecomparison"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/inspection/rules/redundanttypecheck"
	"custos/internal/inspection/rules/xmlparsezerocomparedwithfalse"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestSharedDataFailureWithRealRules(t *testing.T) {
	restore := analysis.FailNamesForTest()
	t.Cleanup(restore)
	engine, err := analysis.NewEngine([]analysis.Rule{redundanttypecheck.New(), nestednotoperators.New()}, analysis.Config{EnableAll: true, PHP: phpversion.Max})
	if err != nil {
		t.Fatal(err)
	}
	file := syntax.Parse("failure.php", []byte(`<?php function f(int $x) { is_string($x); return !!!$x; }`), syntax.Options{Version: phpversion.Max})
	result := engine.AnalyzeResult(file)
	found := false
	for _, finding := range result.Findings {
		if finding.Rule == "NestedNotOperators" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unrelated real rule lost: %+v", result)
	}
	if len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "internal-failure" {
		t.Fatalf("missing shared failure status: %+v", result)
	}
}

// TestPerNodeDeferralWithRealRules: when the flow data fails to build,
// XmlParseZeroComparedWithFalse and DomMissingAttributeCheckedAsNull skip
// the comparisons that need it (an xml_parse() call, a local), and the
// type-check rules that leave those to them report them instead; a
// comparison the owner still checks (a direct getAttribute() call) stays
// the owner's. Both rule orders.
func TestPerNodeDeferralWithRealRules(t *testing.T) {
	src := `<?php
function f($parser, string $data, DOMElement $el): array
{
    $id = $el->getAttribute('id');
    return [xml_parse($parser, $data) === false, $id === null, $el->getAttribute('id') === null];
}
`
	pairs := [][]analysis.Rule{
		{impossibletypecomparison.New(), xmlparsezerocomparedwithfalse.New()},
		{redundanttypecheck.New(), dommissingattributecheckedasnull.New()},
	}
	run := func(rules []analysis.Rule, fail bool) []string {
		t.Helper()
		e, err := analysis.NewEngine(rules, analysis.Config{EnableAll: true, PHP: phpversion.PHP84})
		if err != nil {
			t.Fatal(err)
		}
		if fail {
			defer analysis.FailFlowForTest()()
		}
		f := syntax.Parse("x.php", []byte(src), syntax.Options{Version: phpversion.PHP84})
		var got []string
		for _, fd := range e.Analyze(f) {
			got = append(got, fd.Rule+" "+string(f.Src[fd.Span.Start:fd.Span.End]))
		}
		slices.Sort(got)
		return got
	}
	for _, tc := range []struct {
		pair int
		fail bool
		want []string
	}{
		{0, false, []string{"XmlParseZeroComparedWithFalse xml_parse($parser, $data) === false"}},
		{0, true, []string{"ImpossibleTypeComparison xml_parse($parser, $data) === false", "internal "}},
		{1, false, []string{"DomMissingAttributeCheckedAsNull $el->getAttribute('id') === null", "DomMissingAttributeCheckedAsNull $id === null"}},
		{1, true, []string{"DomMissingAttributeCheckedAsNull $el->getAttribute('id') === null", "RedundantTypeCheck $id === null", "internal "}},
	} {
		for _, rules := range [][]analysis.Rule{pairs[tc.pair], {pairs[tc.pair][1], pairs[tc.pair][0]}} {
			if got := run(rules, tc.fail); !slices.Equal(got, tc.want) {
				t.Errorf("%s with %s, flow failing %v: got %q, want %q", rules[0].ID(), rules[1].ID(), tc.fail, got, tc.want)
			}
		}
	}
}
