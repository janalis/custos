// Package ftppendingtransferacceptedascomplete implements FtpPendingTransferAcceptedAsComplete.
package ftppendingtransferacceptedascomplete

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Wait for FTP_FINISHED before reporting success."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FtpPendingTransferAcceptedAsComplete" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, c) {
	case "ftp_nb_get", "ftp_nb_put", "ftp_nb_fget", "ftp_nb_fput", "ftp_nb_continue":
	default:
		return
	}
	body := ftpPendingAcceptedBody(c)
	if body == nil {
		return
	}
	// A branch that mentions FTP_FINISHED checks or drives the transfer to
	// completion itself; whether it does so correctly is not provable here.
	handled := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		if k, ok := n.(*syntax.ConstFetch); ok && strings.EqualFold(semanticquery.GlobalConstName(ctx, k), "FTP_FINISHED") {
			handled = true
		}
		return !handled
	})
	if handled {
		return
	}
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if syntax.IsVariableScope(n) {
			return false
		}
		switch n.(type) {
		case *syntax.If, *syntax.While, *syntax.For, *syntax.Foreach, *syntax.DoWhile, *syntax.Switch, *syntax.Try:
			return false
		}
		if r, ok := n.(*syntax.Return); ok {
			v, k := astquery.BoolConst(r.Expr)
			found = found || k && v
		}
		if echo, ok := n.(*syntax.Echo); ok {
			for _, v := range echo.Exprs {
				text, k := semanticquery.NativeString(ctx, v)
				found = found || k && strings.EqualFold(strings.TrimSpace(text), "complete")
			}
		}
		return true
	})
	if found {
		ctx.ReportNode(c, message)
	}
}

// ftpPendingAcceptedBody returns the statement an `if` runs when the status
// returned by call is non-zero, so FTP_MOREDATA reaches it: the body of a
// truthy test, the `else` of a negated one. It returns nil when the call is
// not such a condition. Parentheses, assignments, `!` and comparisons with a
// boolean literal are followed; the status is an int, so a strict comparison
// of the status itself with a boolean has a fixed outcome and selects the
// branch that always runs.
func ftpPendingAcceptedBody(call *syntax.FuncCall) syntax.Stmt {
	// `ftp_nb_get(...)` builds a closure and starts no transfer.
	for _, arg := range call.Args.Args {
		if _, ok := arg.(*syntax.VariadicPlaceholder); ok {
			return nil
		}
	}
	var e syntax.Expr = call
	truthy, status := true, true
	for {
		switch p := e.Parent().(type) {
		case *syntax.Paren:
		case *syntax.Assign:
			if p.Value != e || p.Op.Kind != syntax.TEqual || p.ByRef {
				return nil
			}
		case *syntax.Unary:
			if p.Op.Kind != syntax.TExclaim {
				return nil
			}
			truthy, status = !truthy, false
		case *syntax.Binary:
			other := p.Right
			if other == e {
				other = p.Left
			}
			want, ok := astquery.BoolConst(other)
			if !ok {
				return nil
			}
			switch p.Op.Kind {
			case syntax.TIsEqual:
				truthy = truthy == want
			case syntax.TIsNotEqual:
				truthy = truthy != want
			case syntax.TIsIdentical:
				truthy = !status && truthy == want
			case syntax.TIsNotIdentical:
				truthy = status || truthy != want
			default:
				return nil
			}
			status = false
		case *syntax.If:
			switch {
			case truthy:
				return p.Body
			case p.Else != nil:
				return p.Else.Body
			}
			return nil
		default:
			return nil
		}
		e = e.Parent().(syntax.Expr)
	}
}
