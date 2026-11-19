package catalogue

import (
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/rules/accessmodifierpresented"
	"github.com/janalis/custos/internal/inspection/rules/comparisonoperandsorder"
	"github.com/janalis/custos/internal/inspection/rules/disallowwritingintostaticproperties"
	"github.com/janalis/custos/internal/inspection/rules/dynamicinvocationviascoperesolution"
	"github.com/janalis/custos/internal/inspection/rules/implodeargumentsorder"
	"github.com/janalis/custos/internal/inspection/rules/incrementdecrementoperationequivalent"
	"github.com/janalis/custos/internal/inspection/rules/isemptyfunctionusage"
	"github.com/janalis/custos/internal/inspection/rules/isnullfunctionusage"
	"github.com/janalis/custos/internal/inspection/rules/misorderedmodifiers"
	"github.com/janalis/custos/internal/inspection/rules/missingoremptygroupstatement"
	"github.com/janalis/custos/internal/inspection/rules/missusingparentkeyword"
	"github.com/janalis/custos/internal/inspection/rules/nestedassignmentsusage"
	"github.com/janalis/custos/internal/inspection/rules/nestednotoperators"
	"github.com/janalis/custos/internal/inspection/rules/nestedpositiveifstatements"
	"github.com/janalis/custos/internal/inspection/rules/opassignshortsyntax"
	"github.com/janalis/custos/internal/inspection/rules/parameterdefaultvalueisnotnull"
	"github.com/janalis/custos/internal/inspection/rules/propernullcoalescingoperatorusage"
	"github.com/janalis/custos/internal/inspection/rules/selfclassreferencing"
	"github.com/janalis/custos/internal/inspection/rules/shortechotagcanbeused"
	"github.com/janalis/custos/internal/inspection/rules/shortopentagusage"
	"github.com/janalis/custos/internal/inspection/rules/staticclosurecanbeused"
	"github.com/janalis/custos/internal/inspection/rules/staticinvocationviathis"
	"github.com/janalis/custos/internal/inspection/rules/unknowninspection"
	"github.com/janalis/custos/internal/inspection/rules/unnecessarycasting"
	"github.com/janalis/custos/internal/inspection/rules/unnecessarydoublequotes"
	"github.com/janalis/custos/internal/inspection/rules/unnecessaryfinalmodifier"
	"github.com/janalis/custos/internal/inspection/rules/unnecessarysemicolon"
	"github.com/janalis/custos/internal/inspection/rules/unnecessaryusealias"
	"github.com/janalis/custos/internal/inspection/rules/unsetconstructscanbemerged"
	"github.com/janalis/custos/internal/inspection/rules/usinginclusionreturnvalue"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
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
