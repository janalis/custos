package controlflow

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// pdoAPIUsage reports `$st = $pdo->prepare(...); $st->execute();` where no
// parameter is bound: PDO::query() does the same.
type pdoAPIUsage struct{}

func init() { register(pdoAPIUsage{}) }

func (pdoAPIUsage) ID() string { return "PdoApiUsage" }

func (pdoAPIUsage) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }

func (pdoAPIUsage) Semantic() {}

func (pdoAPIUsage) Check(ctx *analysis.Context, n syntax.Node) {
	exec := n.(*syntax.MethodCall)
	if id, ok := exec.Name.(*syntax.Identifier); !ok || !strings.EqualFold(id.Value, "execute") || (exec.Args != nil && len(exec.Args.Args) > 0) { // D1
		return
	}
	stmt, ok := exec.Parent().(*syntax.ExprStmt) // D2
	if !ok || stmt.Expr != syntax.Expr(exec) {
		return
	}
	prev, ok := util.PrevStmt(ctx.File, stmt) // D3
	if !ok {
		return
	}
	ps, ok := prev.(*syntax.ExprStmt) // D4
	if !ok {
		return
	}
	as, ok := ps.Expr.(*syntax.Assign)
	if !ok || as.Op.Kind != syntax.TEqual {
		return
	}
	prep, ok := as.Value.(*syntax.MethodCall)
	if !ok {
		return
	}
	pname, ok := prep.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(pname.Value, "prepare") {
		return
	}
	if !util.EquivalentFoldNames(ctx.File, as.Var, exec.Var) || !pdoPrepareResolves(ctx, prep) { // D6, D5
		return
	}
	var fixes []analysis.Fix
	if prep.Args != nil && len(prep.Args.Args) == 1 {
		fixes = append(fixes, analysis.Fix{
			Title: "Use query()",
			Edits: func() []analysis.TextEdit {
				del := stmt.Span()
				if ws, ok := util.TokenBefore(ctx.File, del.Start); ok && ws.Kind == syntax.TWhitespace {
					del.Start = ws.Start
				}
				return []analysis.TextEdit{
					{Span: pname.Span(), NewText: "query"},
					{Span: del, NewText: ""},
				}
			},
		})
	}
	ctx.ReportNode(exec, "No parameters are bound; call query() instead of prepare() + execute().", fixes...)
}

// pdoPrepareResolves reports whether the prepare() call resolves to a method
// of \PDO or of a class (not a trait) extending it.
func pdoPrepareResolves(ctx *analysis.Context, call *syntax.MethodCall) bool {
	ix := ctx.Index()
	for _, cls := range ctx.TypeOf(call.Var).Classes() {
		c := strings.TrimPrefix(cls, `\`)
		m := ix.FindMethod(c, "prepare", ctx.PHP)
		if m == nil {
			continue
		}
		decl := m.Class // always set by the index extractor
		dc := ix.Class(decl, ctx.PHP)
		if dc == nil || dc.Kind == syntax.KindTrait {
			continue
		}
		if ix.IsSubtype(decl, "PDO", ctx.PHP) {
			return true
		}
	}
	return false
}
