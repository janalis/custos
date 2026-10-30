package curlsslserverspoofing

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// curlSslServerSpoofing reports CURLOPT_SSL_VERIFYHOST / CURLOPT_SSL_VERIFYPEER
// set to a value that disables TLS verification.
type curlSslServerSpoofing struct{}

const (
	curlHostMsg = "Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2."
	curlPeerMsg = "Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled."
)

func (curlSslServerSpoofing) ID() string { return "CurlSslServerSpoofing" }
func (curlSslServerSpoofing) Semantic()  {}
func (curlSslServerSpoofing) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KConstFetch}
}

func (curlSslServerSpoofing) Check(ctx *analysis.Context, n syntax.Node) {
	k := n.(*syntax.ConstFetch) // the parser always gives it a name token
	var host bool
	switch astquery.LastNamePart(k.Name.Value) {
	case "CURLOPT_SSL_VERIFYHOST":
		host = true
	case "CURLOPT_SSL_VERIFYPEER":
	default:
		return // E5
	}
	site, value := curlSettingSite(ctx, k)
	if site == nil || value == nil {
		return
	}
	w := &curlValues{ValueWalk: flowquery.NewValueWalk(ctx.File), ctx: ctx, host: host, busy: map[syntax.Node]bool{}}
	if enable, disable := w.verdict(value); !disable || enable { // D6, E3
		return
	}
	if host {
		ctx.ReportNode(site, curlHostMsg)
	} else {
		ctx.ReportNode(site, curlPeerMsg)
	}
}

// curlValues is the rule's value discovery (D4–D6): shared discovery, except
// that a local variable only takes the assignments reaching its use and a
// property overwritten with an undecidable value makes the result unknown.
type curlValues struct {
	flowquery.ValueWalk
	ctx  *analysis.Context
	host bool
	busy map[syntax.Node]bool // property writes being judged (cycles)
}

// verdict classifies every value discovered from e; both results are false
// when the discovery is unknown.
func (w *curlValues) verdict(e syntax.Expr) (enable, disable bool) {
	w.ElvisCond, w.SkipEmpty = true, true
	w.Resolve = w.resolve
	w.Collect(e)
	vals, _ := w.Result()
	for _, v := range vals {
		switch curlClassify(w.ctx, v, w.host) {
		case 1:
			enable = true
		case -1:
			disable = true
		}
	}
	return enable, disable
}

func (w *curlValues) resolve(e syntax.Expr) bool {
	env := w.ctx.Types()
	switch x := e.(type) {
	case *syntax.Variable:
		w.variable(x)
		return true
	case *syntax.PropertyFetch:
		if id, ok := x.Name.(*syntax.Identifier); ok {
			w.overwrites(x, env.TypeOf(x.Var).Classes(), id.Value)
		}
	case *syntax.StaticPropertyFetch:
		if v, ok := x.Name.(*syntax.Variable); ok {
			w.overwrites(x, []string{env.ClassRef(x.Class)}, v.Name)
		}
	}
	vals, known := semanticquery.DiscoverValuesKnown(env, e)
	if !known {
		w.Stop = true
	}
	w.Out = append(w.Out, vals...)
	return true
}

// variable collects the parameter default (when the entry value reaches v)
// and the plain assignments reaching v; an unstable variable is unknown.
func (w *curlValues) variable(v *syntax.Variable) {
	scope := syntax.EnclosingFuncLike(v)
	if scope == nil || v.NameExpr != nil {
		return
	}
	params, body := astquery.ScopeParts(scope)
	if flowquery.UnstableVariableIn(w.F, body, v.Name) {
		w.Stop = true
		return
	}
	defs, entry := flowquery.ReachingAssignmentsIn(w.F, scope, v, v.Name)
	if entry {
		for _, p := range params {
			if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
				w.Collect(p.Default)
			}
		}
	}
	for _, d := range defs {
		w.Collect(astquery.AssignedValue(d))
	}
}

// overwrites makes the result unknown when the last word on the property is
// an assignment storing a value in which nothing can be classified: the
// declared default or an earlier assignment then says nothing about what the
// option receives. Two places are read in source order: the current scope up
// to the fetch, and the declaring class's constructor. An unclassifiable
// value only counts when its statement always runs (curlAlways); any later
// classifiable assignment makes the property decidable again.
func (w *curlValues) overwrites(fetch syntax.Expr, classes []string, name string) {
	env := w.ctx.Types()
	var own syntax.Node
	var roots []syntax.Node
	if scope := syntax.EnclosingFuncLike(fetch); scope != nil {
		_, own = astquery.ScopeParts(scope)
		roots = append(roots, own)
	}
	for _, cls := range classes {
		p := env.Index.FindProperty(strings.TrimPrefix(cls, `\`), name, env.PHP)
		if p == nil {
			continue
		}
		if decl := semanticquery.ClassDecl(w.F, env.Index.Class(p.Class, env.PHP)); decl != nil {
			for _, m := range decl.Members {
				if meth, ok := m.(*syntax.Method); ok && meth.Name != nil && meth.Body != nil &&
					strings.EqualFold(meth.Name.Value, "__construct") && syntax.Node(meth.Body) != own {
					roots = append(roots, meth.Body)
				}
			}
		}
	}
	for _, root := range roots {
		local := root == own
		unknown := false
		for _, a := range flowquery.AssignmentsOfKind(w.F, root, fetch.Kind()) {
			rhs := astquery.AssignedValue(a)
			// In the constructor the property is written through $this
			// (or any static reference), whatever the receiver read here.
			if w.busy[rhs] || !astquery.Equivalent(w.F, a.Var, fetch) && (local || !curlOwnProperty(a.Var, name)) {
				continue
			}
			if local && a.Span().Start > fetch.Span().Start {
				break // written after the value is read
			}
			w.busy[rhs] = true
			sub := &curlValues{ValueWalk: flowquery.NewValueWalk(w.F), ctx: w.ctx, host: w.host, busy: w.busy}
			enable, disable := sub.verdict(rhs)
			delete(w.busy, rhs)
			if enable || disable {
				unknown = false
			} else if curlAlways(a, root, fetch, local) {
				unknown = true
			}
		}
		if unknown {
			w.Stop = true
			return
		}
	}
}

// curlAlways reports whether assignment a is a statement of its own that
// runs whenever the value is read: in the fetch's own scope, a statement of
// a block enclosing the fetch; in a constructor, a statement of its body.
func curlAlways(a *syntax.Assign, root syntax.Node, fetch syntax.Expr, local bool) bool {
	stmt, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	block := stmt.Parent()
	if !local {
		return block == root
	}
	for p := fetch.Parent(); p != nil; p = p.Parent() {
		if p == block {
			return true
		}
	}
	return false
}

// curlOwnProperty reports whether target, a property fetch, is `$this->name`
// or a static `…::$name`.
func curlOwnProperty(target syntax.Expr, name string) bool {
	if p, ok := target.(*syntax.PropertyFetch); ok {
		v, isVar := p.Var.(*syntax.Variable)
		id, isID := p.Name.(*syntax.Identifier)
		return isVar && isID && v.Name == "this" && id.Value == name
	}
	v, ok := target.(*syntax.StaticPropertyFetch).Name.(*syntax.Variable)
	return ok && v.Name == name
}

// curlSettingSite implements D1–D3: the reported node and the value set.
func curlSettingSite(ctx *analysis.Context, k *syntax.ConstFetch) (site syntax.Node, value syntax.Expr) {
	switch p := k.Parent().(type) {
	case *syntax.Arg: // D1 (an Arg's only expression child is its value)
		call := astquery.ParentFuncCall(k)
		if call == nil || !ctx.IsGlobalFunctionCall(call, "curl_setopt") || astquery.ArgCount(call) != 3 {
			return nil, nil
		}
		third, ok := call.Args.Args[2].(*syntax.Arg)
		if !ok || third.Value == nil {
			return nil, nil
		}
		return call, third.Value
	case *syntax.ArrayItem: // D2
		if p.Key != syntax.Expr(k) || p.Value == nil {
			return nil, nil
		}
		arr, ok := p.Parent().(*syntax.Array)
		if !ok || curlDestructuring(arr) || curlNameMap(arr) { // list() / destructuring read the option; E7
			return nil, nil
		}
		return p, p.Value
	case *syntax.ArrayDimFetch: // D3
		if p.Dim != syntax.Expr(k) {
			return nil, nil
		}
		var cur syntax.Expr = p
		for {
			up, ok := cur.Parent().(*syntax.ArrayDimFetch)
			if !ok || up.Var != cur {
				break
			}
			cur = up
		}
		a, ok := cur.Parent().(*syntax.Assign)
		if !ok || a.Var != cur || a.Op.Kind != syntax.TEqual || a.Value == nil {
			return nil, nil
		}
		return a, a.Value
	}
	return nil, nil
}

// curlDestructuring reports whether arr is (part of) a destructuring
// pattern: an assignment target or a foreach value, possibly nested in
// other arrays or list().
// curlNameMap reports whether every element of arr has a non-numeric string
// literal value (`[CURLOPT_SSL_VERIFYPEER => 'verify_peer', …]`): the array
// maps options to names, it does not set them (E7).
func curlNameMap(arr *syntax.Array) bool {
	for _, it := range arr.Items {
		var v syntax.Expr
		if it != nil { // an empty slot only parses in a pattern
			v = it.Value
		}
		lit, ok := v.(*syntax.Literal)
		if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 {
			return false
		}
		text := strings.TrimSpace(lit.Raw[1 : len(lit.Raw)-1])
		if _, err := strconv.ParseFloat(text, 64); err == nil || text == "" {
			return false
		}
	}
	return true
}

func curlDestructuring(arr *syntax.Array) bool {
	var cur syntax.Expr = arr
	for {
		it, ok := cur.Parent().(*syntax.ArrayItem)
		if !ok || it.Value != cur {
			break
		}
		cur = it.Parent().(syntax.Expr) // an Array or a list()
	}
	switch p := cur.Parent().(type) {
	case *syntax.Assign:
		return p.Var == cur
	case *syntax.Foreach:
		return p.Value == cur
	}
	return false
}

// curlClassify implements D4/D5: 1 = enable, -1 = disable, 0 = ignored.
func curlClassify(ctx *analysis.Context, v syntax.Expr, host bool) int {
	want := "1"
	if host {
		want = "2"
	}
	if content, _, ok := astquery.QuotedStringRaw(v); ok {
		return curlBool(content == want)
	}
	if _, ok := v.(*syntax.InterpolatedString); ok {
		return -1
	}
	if c, ok := v.(*syntax.ConstFetch); ok {
		name := astquery.LastNamePart(c.Name.Value)
		switch strings.ToLower(name) {
		case "true":
			return curlBool(!host)
		case "false", "null":
			return -1
		}
		// A constant declared outside this file: classify its value text.
		if k := semanticquery.ResolveConstant(ctx.Types(), c); k != nil {
			return curlClassifyText(strings.TrimSpace(k.Value), want)
		}
		return -1
	}
	return curlClassifyText(ctx.Text(v), want)
}

func curlClassifyText(text, want string) int {
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') && text[len(text)-1] == text[0] {
		return curlBool(text[1:len(text)-1] == want)
	}
	if len(text) == 1 {
		return curlBool(text == want)
	}
	switch strings.ToLower(text) {
	case "true":
		return curlBool(want == "1")
	case "false", "null":
		return -1
	}
	return 0
}

func curlBool(enable bool) int {
	if enable {
		return 1
	}
	return -1
}
