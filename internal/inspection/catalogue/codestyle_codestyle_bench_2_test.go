package catalogue

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/accessmodifierpresented"
	"custos/internal/inspection/rules/comparisonoperandsorder"
	"custos/internal/inspection/rules/disallowwritingintostaticproperties"
	"custos/internal/inspection/rules/dynamicinvocationviascoperesolution"
	"custos/internal/inspection/rules/implodeargumentsorder"
	"custos/internal/inspection/rules/incrementdecrementoperationequivalent"
	"custos/internal/inspection/rules/isemptyfunctionusage"
	"custos/internal/inspection/rules/isnullfunctionusage"
	"custos/internal/inspection/rules/misorderedmodifiers"
	"custos/internal/inspection/rules/missingoremptygroupstatement"
	"custos/internal/inspection/rules/missusingparentkeyword"
	"custos/internal/inspection/rules/nestedassignmentsusage"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/inspection/rules/nestedpositiveifstatements"
	"custos/internal/inspection/rules/opassignshortsyntax"
	"custos/internal/inspection/rules/parameterdefaultvalueisnotnull"
	"custos/internal/inspection/rules/propernullcoalescingoperatorusage"
	"custos/internal/inspection/rules/selfclassreferencing"
	"custos/internal/inspection/rules/shortechotagcanbeused"
	"custos/internal/inspection/rules/shortopentagusage"
	"custos/internal/inspection/rules/staticclosurecanbeused"
	"custos/internal/inspection/rules/staticinvocationviathis"
	"custos/internal/inspection/rules/unknowninspection"
	"custos/internal/inspection/rules/unnecessarycasting"
	"custos/internal/inspection/rules/unnecessarydoublequotes"
	"custos/internal/inspection/rules/unnecessaryfinalmodifier"
	"custos/internal/inspection/rules/unnecessarysemicolon"
	"custos/internal/inspection/rules/unnecessaryusealias"
	"custos/internal/inspection/rules/unsetconstructscanbemerged"
	"custos/internal/inspection/rules/usinginclusionreturnvalue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func BenchmarkCodeStyle(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{accessmodifierpresented.New(), comparisonoperandsorder.New(), disallowwritingintostaticproperties.New(), dynamicinvocationviascoperesolution.New(), implodeargumentsorder.New(), incrementdecrementoperationequivalent.New(), isemptyfunctionusage.New(), isnullfunctionusage.New(), misorderedmodifiers.New(), missusingparentkeyword.New(), missingoremptygroupstatement.New(), nestedassignmentsusage.New(), nestednotoperators.New(), nestedpositiveifstatements.New(), opassignshortsyntax.New(), parameterdefaultvalueisnotnull.New(), propernullcoalescingoperatorusage.New(), selfclassreferencing.New(), shortechotagcanbeused.New(), shortopentagusage.New(), staticclosurecanbeused.New(), staticinvocationviathis.New(), unnecessarydoublequotes.New(), unknowninspection.New(), unnecessarycasting.New(), unnecessaryfinalmodifier.New(), unnecessarysemicolon.New(), unnecessaryusealias.New(), unsetconstructscanbemerged.New(), usinginclusionreturnvalue.New()}, analysis.Config{EnableAll: true})
	if err != nil {
		b.Fatal(err)
	}
	f := syntax.Parse("bench.php", []byte(benchSrc), syntax.Options{Version: phpversion.Max})
	if len(f.Errors) > 0 {
		b.Fatal(f.Errors[0])
	}
	e.Analyze(f) // warm file queries
	b.ReportAllocs()
	for b.Loop() {
		if fs := e.Analyze(f); len(fs) != 0 {
			b.Fatalf("unexpected findings: %v", fs)
		}
	}
}
