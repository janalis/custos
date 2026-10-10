package semanticquery

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// ExpansionDMethodSucceeded proves a boolean method succeeded on the current path.
func ExpansionDMethodSucceeded(ctx *analysis.Context, c *syntax.MethodCall, at syntax.Node) bool {
	accepts := func(e syntax.Expr, truth bool) bool {
		e = syntax.UnwrapParens(e)
		if e == c {
			return truth
		}
		if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TExclaim {
			return syntax.UnwrapParens(u.Expr) == c && !truth
		}
		return false
	}
	for p := at; p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		if b, ok := p.(*syntax.If); ok && b.Body.Span().Contains(at.Span()) && accepts(b.Cond, true) {
			return true
		}
		if st, ok := p.(syntax.Stmt); ok {
			prev, exists := astquery.PrevStmt(ctx.File, st)
			if b, ok := prev.(*syntax.If); exists && ok && b.Else == nil && len(b.ElseIfs) == 0 && syntax.Terminates(b.Body) && accepts(b.Cond, false) {
				return true
			}
		}
	}
	return false
}

// ExpansionDMethods retains ordered same-object calls; ambiguous path effects
// discard the complete state history rather than inventing a state transition.
func ExpansionDMethods(ctx *analysis.Context, at syntax.Node, obj syntax.Expr) []*syntax.MethodCall {
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(at))
	if len(calls) > 256 {
		return nil
	}
	out := []*syntax.MethodCall{}
	for _, record := range calls {
		c, ok := record.Node.(*syntax.MethodCall)
		if !ok || c.Span().End >= at.Span().Start || !NativeSameValue(ctx, c.Var, obj) {
			continue
		}
		if !NativeDominates(c, at) && !ExpansionDMethodSucceeded(ctx, c, at) {
			return nil
		}
		out = append(out, c)
	}
	return out
}

// ExpansionDXML parses bounded literal XML without resolving external entities.
// Namespaces and nontrivial declarations remain unknown to these inspections.
func ExpansionDXML(s string) (children []string, text string, plain, known bool) {
	if len(s) > 65536 {
		return nil, "", false, false
	}
	d := xml.NewDecoder(strings.NewReader(s))
	depth := 0
	roots := 0
	plain = true
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			return children, text, plain, depth == 0 && roots == 1
		}
		if err != nil {
			return nil, "", false, false
		}
		switch x := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 1 {
				roots++
			}
			if x.Name.Space != "" {
				return nil, "", false, false
			}
			if depth == 2 {
				children = append(children, x.Name.Local)
				plain = false
			}
			if len(x.Attr) > 0 {
				plain = false
				for _, a := range x.Attr {
					if a.Name.Local == "xmlns" || a.Name.Space != "" {
						return nil, "", false, false
					}
				}
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 1 {
				text += string(x)
			}
		case xml.Directive:
			return nil, "", false, false
		}
	}
}

// ExpansionDFrames proves a lower bound for successfully appended local frames.
func ExpansionDFrames(ctx *analysis.Context, c *syntax.MethodCall) int {
	return expansionDFrames(ctx, c.Var, c, 0)
}

func expansionDFrames(ctx *analysis.Context, obj syntax.Expr, at syntax.Node, depth int) int {
	if depth > 8 {
		return 0
	}
	if !ExpansionDUnaliased(ctx, obj, at) {
		return 0
	}
	creation := NativeConstruction(ctx, obj, "Imagick")
	if creation == nil || CallArgument(creation.Args, 0, "files") != nil {
		return 0
	}
	count := 0
	for _, prior := range ExpansionDMethods(ctx, at, obj) {
		switch ExpansionDMethodName(prior) {
		case "newimage":
			if !ExpansionDMethodSucceeded(ctx, prior, at) {
				return 0
			}
			count++
		case "addimage":
			if !ExpansionDMethodSucceeded(ctx, prior, at) {
				return 0
			}
			frames := expansionDFrames(ctx, CallArgument(prior.Args, 0, "source"), prior, depth+1)
			if frames == 0 {
				return 0
			}
			count += frames
		case "getnumberimages", "getimageblob", "getimagesblob", "setiteratorindex":
		default:
			return 0
		}
	}
	return count
}

// ExpansionDMessage identifies one preceding successful local send selected by
// an exact positive type, with no intervening queue receive or send ambiguity.
func ExpansionDMessage(ctx *analysis.Context, c *syntax.FuncCall) *syntax.FuncCall {
	queue := CallArgument(c.Args, 0, "queue")
	if !ExpansionDUnaliased(ctx, queue, c) {
		return nil
	}
	wanted, known := NativeContractInt(ctx, CallArgument(c.Args, 1, "desired_message_type"))
	if !known || wanted <= 0 {
		return nil
	}
	flags := CallArgument(c.Args, 6, "flags")
	if flags != nil {
		v, k := NativeContractInt(ctx, flags)
		if !k || v&4 != 0 {
			return nil
		}
	}
	calls := NativeStreamCalls(ctx, c, queue, "msg_send", "msg_receive", "msg_remove_queue")
	if len(calls) != 1 {
		return nil
	}
	send := calls[0]
	if !NativeBuiltin(ctx, send, "msg_send") || !NativeCallSucceeded(ctx, send, c) {
		return nil
	}
	sent, k := NativeContractInt(ctx, CallArgument(send.Args, 1, "message_type"))
	if !k || sent != wanted {
		return nil
	}
	return send
}

// ExpansionDPristine excludes local writes through XML property/array handles.
func ExpansionDPristine(ctx *analysis.Context, obj syntax.Expr, at syntax.Node) bool {
	if !ExpansionDUnaliased(ctx, obj, at) {
		return false
	}
	good := true
	visited := 0
	scope := syntax.EnclosingVariableScope(at)
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		visited++
		if visited > 512 {
			good = false
			return false
		}
		if n.Span().Start >= at.Span().Start {
			return false
		}
		var targets []syntax.Expr
		switch x := n.(type) {
		case *syntax.Assign:
			targets = []syntax.Expr{x.Var}
		case *syntax.IncDec:
			targets = []syntax.Expr{x.Var}
		case *syntax.Unset:
			targets = x.Vars
		}
		for _, target := range targets {
			var base syntax.Expr
			switch x := target.(type) {
			case *syntax.PropertyFetch:
				base = x.Var
			case *syntax.ArrayDimFetch:
				base = x.Var
			}
			for {
				switch x := base.(type) {
				case *syntax.PropertyFetch:
					base = x.Var
				case *syntax.ArrayDimFetch:
					base = x.Var
				default:
					goto rooted
				}
			}
		rooted:
			if base != nil && NativeSameValue(ctx, base, obj) {
				good = false
			}
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return good
}

// ExpansionDMethodName returns a literal method name, excluding dynamic names.
func ExpansionDMethodName(c *syntax.MethodCall) string {
	if id, ok := c.Name.(*syntax.Identifier); ok {
		return strings.ToLower(id.Value)
	}
	return ""
}

type expansionDEscapeScopeKey struct{ scope syntax.Node }

type expansionDEscapeCandidates struct {
	nodes []syntax.Node
	known bool
}

// expansionDEscapes caches syntax-only escape candidates once per lexical scope.
// Receiver identity and types remain context-dependent and are never cached.
func expansionDEscapes(ctx *analysis.Context, scope syntax.Node) expansionDEscapeCandidates {
	return ctx.File.Memo(expansionDEscapeScopeKey{scope}, func() any {
		result := expansionDEscapeCandidates{known: true}
		visited, events := 0, 0
		visit := func(n syntax.Node) bool {
			if !result.known {
				return false
			}
			if n != scope && syntax.IsVariableScope(n) {
				if closure, ok := n.(*syntax.Closure); ok {
					visited += len(closure.Uses)
					if visited > 4096 {
						result.known = false
						result.nodes = nil
						return false
					}
					for _, use := range closure.Uses {
						result.nodes = append(result.nodes, use)
					}
				}
				return false
			}
			visited++
			switch n.(type) {
			case *syntax.Assign, *syntax.Unset, *syntax.FuncCall, *syntax.MethodCall:
				events++
			}
			if visited > 4096 || events > 512 {
				result.known = false
				result.nodes = nil
				return false
			}
			if a, ok := n.(*syntax.Assign); ok && a.ByRef {
				result.nodes = append(result.nodes, a)
			}
			return true
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
		return result
	}).(expansionDEscapeCandidates)
}

// ExpansionDUnaliased rejects escaping aliases using bounded cached syntax.
func ExpansionDUnaliased(ctx *analysis.Context, obj syntax.Expr, at syntax.Node) bool {
	candidates := expansionDEscapes(ctx, syntax.EnclosingVariableScope(at))
	if !candidates.known {
		return false
	}
	for _, candidate := range candidates.nodes {
		if candidate.Span().Start >= at.Span().Start {
			break
		}
		switch x := candidate.(type) {
		case *syntax.Assign:
			if astquery.Equivalent(ctx.File, x.Value, obj) || NativeSameValue(ctx, x.Value, obj) {
				return false
			}
		case *syntax.ClosureUse:
			if !astquery.Equivalent(ctx.File, x.Var, obj) {
				continue
			}
			if x.ByRef {
				return false
			}
			object := ctx.Types().Native().TypeOf(obj)
			if len(object.Classes()) > 0 || object.Has("resource") {
				return false
			}
		}
	}
	return true
}
