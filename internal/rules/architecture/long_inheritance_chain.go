package architecture

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// longInheritanceChain reports classes with too many ancestor classes.
type longInheritanceChain struct{}

func init() { register(longInheritanceChain{}) }

func (longInheritanceChain) ID() string { return "LongInheritanceChain" }

func (longInheritanceChain) Semantic() {}

func (longInheritanceChain) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KClassLike} }

// licStopClasses are framework base classes ending the walk (S).
var licStopClasses = map[string]bool{
	`PHPUnit_Framework_TestCase`: true, `PHPUnit\Framework\TestCase`: true, `yii\base\Component`: true,
	`yii\base\Behavior`: true, `CComponent`: true, `Zend\Form\Form`: true, `Phalcon\Di\Injectable`: true,
}

func (longInheritanceChain) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.Name == nil || cl.ClassKind != syntax.KindClass || cl.Name.Span().Len() == 0 {
		return
	}
	fqn := util.ClassDeclFQN(ctx.Names(), cl)
	if strings.HasSuffix(cl.Name.Value, "Exception") || util.IsTestPath(ctx.File.Path) || util.IsTestClassFQN(fqn) { // D1
		return
	}
	ix := ctx.Index()
	parent := util.ParentFQN(ctx.Names(), cl)
	if parent == "" {
		return
	}
	p1 := ix.Class(parent, ctx.PHP)
	if p1 == nil {
		return
	}
	if !cl.Modifiers.Has(syntax.TAbstract) && p1.Abstract { // D2
		return
	}
	seen := map[string]bool{strings.ToLower(fqn): true}
	count := 0
	for cur := p1; cur != nil; { // D3
		k := strings.ToLower(strings.TrimPrefix(cur.FQN, `\`))
		if seen[k] { // back at the class itself (E6) or any other cycle
			break
		}
		seen[k] = true
		var next *index.Class
		if cur.Parent != "" {
			next = ix.Class(cur.Parent, ctx.PHP)
		}
		count++
		if next != nil {
			nfqn := strings.TrimPrefix(next.FQN, `\`)
			if licStopClasses[nfqn] {
				count++
				break
			}
			short := nfqn[strings.LastIndexByte(nfqn, '\\')+1:]
			if strings.HasSuffix(short, "Exception") {
				return
			}
		}
		cur = next
	}
	if count >= ctx.Int("COMPLAIN_THRESHOLD") && !util.DocHasTag(ctx.File, cl, "deprecated") { // D4
		ctx.ReportSeverity(cl.Name.Span(), meta.SeverityInfo, strconv.Itoa(count)+" levels of parent classes; prefer composition over deep inheritance.")
	}
}
