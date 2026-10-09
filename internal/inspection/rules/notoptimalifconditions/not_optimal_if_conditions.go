package notoptimalifconditions

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// notOptimalIfConditions inspects if/elseif conditions for operand ordering,
// keyword logical operators and instanceof flaws.
type notOptimalIfConditions struct{}

const (
	notOptimalOrderMsg      = "Cheaper check placed after a costlier one; evaluate it first."
	notOptimalAndMsg        = "Use '&&' instead of 'and'."
	notOptimalOrMsg         = "Use '||' instead of 'or'."
	notOptimalEqualityMsg   = "Equality check on a value also tested with instanceof; verify the logic."
	notOptimalInstanceofMsg = "Redundant instanceof: another check on the same value already covers this type."
)

func (notOptimalIfConditions) ID() string               { return "NotOptimalIfConditions" }
func (notOptimalIfConditions) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (r notOptimalIfConditions) Check(ctx *analysis.Context, n syntax.Node) {
	ifs := n.(*syntax.If)
	conds := []syntax.Expr{ifs.Cond}
	for _, ei := range ifs.ElseIfs {
		conds = append(conds, ei.Cond)
	}
	optimize := ctx.Bool("SUGGEST_OPTIMIZING_CONDITIONS")
	literal := ctx.Bool("REPORT_LITERAL_OPERATORS")
	instanceofFlaws := ctx.Bool("REPORT_INSTANCE_OF_FLAWS")
	for _, c := range conds {
		if c == nil || c.Span().Len() == 0 {
			continue
		}
		ops, op := splitCondition(c)
		if optimize && len(ops) >= 2 {
			r.ordering(ctx, ops)
		}
		if literal {
			keywordOperators(ctx, c)
		}
		if instanceofFlaws && len(ops) >= 2 {
			if op == syntax.TBooleanAnd {
				equalityNextToInstanceof(ctx, ops)
			}
			redundantInstanceof(ctx, ops, op)
		}
	}
}

// splitCondition flattens a condition into its && / || operands.
func splitCondition(c syntax.Expr) ([]syntax.Expr, syntax.TokenKind) {
	e := syntax.UnwrapParens(c)
	if u, ok := e.(*syntax.Unary); ok {
		e = syntax.UnwrapParens(u.Expr)
	}
	b, ok := e.(*syntax.Binary)
	if !ok || (b.Op.Kind != syntax.TBooleanAnd && b.Op.Kind != syntax.TBooleanOr) {
		return []syntax.Expr{e}, 0
	}
	op := b.Op.Kind
	ops := []syntax.Expr{syntax.UnwrapParens(b.Right)}
	left := syntax.UnwrapParens(b.Left)
	for {
		lb, ok := left.(*syntax.Binary)
		if !ok || lb.Op.Kind != op {
			break
		}
		ops = append(ops, syntax.UnwrapParens(lb.Right))
		left = syntax.UnwrapParens(lb.Left)
	}
	ops = append(ops, left)
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}
	return ops, op
}

func (notOptimalIfConditions) ordering(ctx *analysis.Context, ops []syntax.Expr) {
	costs := make([]int, len(ops))
	known := make([]bool, len(ops))
	for i, o := range ops {
		costs[i] = conditionCost(ctx, o)
		known[i] = !costUnknown(ctx, o)
	}
	for i := 1; i < len(ops); i++ {
		if known[i] && known[i-1] && costs[i] < costs[i-1] && !operandsCoupled(ctx, ops[i-1], ops[i]) {
			ctx.ReportSeverity(ops[i].Span(), diagnostic.SeverityInfo, notOptimalOrderMsg)
		}
	}
}

var cheapFunctions = map[string]bool{
	"array_key_exists": true, "function_exists": true, "property_exists": true, "class_exists": true,
	"interface_exists": true, "trait_exists": true, "is_a": true, "is_subclass_of": true, "defined": true,
	"is_array": true, "is_bool": true, "is_callable": true, "is_countable": true, "is_float": true,
	"is_double": true, "is_real": true, "is_int": true, "is_integer": true, "is_long": true,
	"is_iterable": true, "is_null": true, "is_numeric": true, "is_object": true, "is_resource": true,
	"is_scalar": true, "is_string": true,
}

// filesystemFunctions ask the operating system about a path (custos): a
// stat() call costs more than any in-memory check, so they are never the
// cheaper operand of a pure computation.
var filesystemFunctions = map[string]bool{
	"file_exists": true, "is_file": true, "is_dir": true, "is_link": true, "is_readable": true,
	"is_writable": true, "is_writeable": true, "is_executable": true, "is_uploaded_file": true,
	"filesize": true, "filemtime": true, "fileatime": true, "filectime": true, "fileinode": true,
	"fileowner": true, "filegroup": true, "fileperms": true, "filetype": true, "stat": true,
	"lstat": true, "realpath": true, "glob": true, "disk_free_space": true, "disk_total_space": true,
}

// notOptimalArgsCost sums the argument costs (calls always carry an
// argument list).
func notOptimalArgsCost(ctx *analysis.Context, l *syntax.ArgList) int {
	s := 0
	for _, a := range l.Args {
		s += conditionCost(ctx, a)
	}
	return s
}

func conditionCost(ctx *analysis.Context, e syntax.Expr) int {
	if e == nil {
		return 0
	}
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.ConstFetch, *syntax.MagicConst, *syntax.Name, *syntax.ClassConstFetch, *syntax.Identifier,
		*syntax.StringPart, *syntax.VariadicPlaceholder:
		return 0
	case *syntax.Literal:
		return 0
	case *syntax.Variable:
		return conditionCost(ctx, x.NameExpr)
	case *syntax.InterpolatedString:
		s := 0
		for _, p := range x.Parts {
			s += conditionCost(ctx, p)
		}
		return s
	case *syntax.PropertyFetch:
		if classifyPropertyRead(ctx, x) == propComputed {
			return conditionCost(ctx, x.Var) + 5
		}
		return conditionCost(ctx, x.Var) + conditionCost(ctx, x.Name)
	case *syntax.StaticPropertyFetch:
		return conditionCost(ctx, x.Class) + conditionCost(ctx, x.Name)
	case *syntax.ArrayDimFetch:
		return conditionCost(ctx, x.Var) + conditionCost(ctx, x.Dim)
	case *syntax.Empty:
		return conditionCost(ctx, x.Expr)
	case *syntax.Isset:
		s := 0
		for _, v := range x.Vars {
			s += conditionCost(ctx, v)
		}
		return s
	case *syntax.Arg:
		return conditionCost(ctx, x.Value)
	case *syntax.FuncCall:
		c := notOptimalArgsCost(ctx, x.Args)
		switch name := ctx.GlobalFunctionName(x); { // resolved, lower-case: user functions cost like any call
		case filesystemFunctions[name]:
			c += 50
		case !cheapFunctions[name]:
			c += 5
		}
		return c
	case *syntax.MethodCall:
		return notOptimalArgsCost(ctx, x.Args) + conditionCost(ctx, x.Var) + 5
	case *syntax.StaticCall:
		return notOptimalArgsCost(ctx, x.Args) + conditionCost(ctx, x.Class) + 5
	case *syntax.Unary:
		return conditionCost(ctx, x.Expr)
	case *syntax.IncDec:
		return conditionCost(ctx, x.Var)
	case *syntax.Binary:
		return conditionCost(ctx, x.Left) + conditionCost(ctx, x.Right)
	case *syntax.Instanceof:
		return conditionCost(ctx, x.Expr) + conditionCost(ctx, x.Class)
	case *syntax.Array:
		s := 0
		for _, it := range x.Items {
			if it != nil {
				s += conditionCost(ctx, it.Key) + conditionCost(ctx, it.Value)
			}
		}
		return s
	case *syntax.Assign:
		return conditionCost(ctx, x.Value)
	case *syntax.Ternary:
		t, f := conditionCost(ctx, x.Then), conditionCost(ctx, x.Else)
		if f > t {
			t = f
		}
		return conditionCost(ctx, x.Cond) + t
	}
	return 10
}

// Property read classes (see the spec's *Property fetches*).
const (
	propStored = iota
	propComputed
	propUnknown
)

// classifyPropertyRead tells whether an instance property read reads a
// stored slot, runs code (get hook, virtual, abstract, __get()) or cannot
// be classified.
func classifyPropertyRead(ctx *analysis.Context, f *syntax.PropertyFetch) int {
	id, ok := f.Name.(*syntax.Identifier)
	if !ok {
		return propUnknown // dynamic name
	}
	atoms := ctx.TypeOf(f.Var).Without("null").Atoms()
	if len(atoms) == 0 {
		return propUnknown
	}
	out := propStored
	for _, a := range atoms {
		if !strings.HasPrefix(a, `\`) || strings.HasSuffix(a, "[]") || ctx.Index().Class(a, ctx.PHP) == nil {
			return propUnknown // scalar, mixed, object, template or unresolved class
		}
		p := ctx.Index().FindProperty(a, id.Value, ctx.PHP)
		switch {
		case p != nil && !p.Magic:
			if p.ReadsRunCode {
				out = propComputed
			}
		case ctx.Index().FindMethod(a, "__get", ctx.PHP) != nil:
			out = propComputed
		}
	}
	return out
}

// costUnknown reports whether e reads a property whose cost cannot be
// established (closures and nested classes are not evaluated).
func costUnknown(ctx *analysis.Context, e syntax.Expr) bool {
	unknown := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.PropertyFetch:
			if classifyPropertyRead(ctx, c) == propUnknown {
				unknown = true
			}
		}
		return !unknown
	})
	return unknown
}

// operandsCoupled implements scenarios S1–S5.
func operandsCoupled(ctx *analysis.Context, prev, cur syntax.Expr) bool {
	if operandsCoupledS123(ctx, prev, cur) || typeGuarded(ctx, prev, cur) {
		return true
	}
	// S4: side effects are never reordered.
	return notOptimalImpure(ctx, prev) || notOptimalImpure(ctx, cur)
}

func operandsCoupledS123(ctx *analysis.Context, prev, cur syntax.Expr) bool {
	// S1
	var mutated []syntax.Node
	addTargets := func(t syntax.Expr) {
		t = syntax.UnwrapParens(t)
		switch l := t.(type) {
		case *syntax.List, *syntax.Array:
			syntax.Inspect(l, func(x syntax.Node) bool {
				if v, ok := x.(*syntax.Variable); ok {
					mutated = append(mutated, v)
				}
				return true
			})
		default:
			if t != nil {
				mutated = append(mutated, t)
			}
		}
	}
	syntax.Inspect(prev, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.Assign:
			addTargets(c.Var)
		case *syntax.FuncCall:
			mutated = append(mutated, byRefVarArgs(ctx, c, c.Args)...)
		case *syntax.MethodCall:
			mutated = append(mutated, byRefVarArgs(ctx, c, c.Args)...)
		case *syntax.StaticCall:
			mutated = append(mutated, byRefVarArgs(ctx, c, c.Args)...)
		}
		return true
	})
	if len(mutated) > 0 {
		hit := false
		syntax.Inspect(cur, func(x syntax.Node) bool {
			if hit {
				return false
			}
			for _, m := range mutated {
				if astquery.EquivalentFoldNames(ctx.File, x, m) {
					hit = true
					break
				}
			}
			return true
		})
		if hit {
			return true
		}
	}
	// S2
	var chain []syntax.Node
	syntax.Inspect(cur, func(x syntax.Node) bool {
		d, ok := x.(*syntax.ArrayDimFetch)
		if !ok {
			return true
		}
		if p, ok := d.Parent().(*syntax.ArrayDimFetch); ok && p.Var == syntax.Expr(d) && x != syntax.Node(cur) {
			return true
		}
		for b := d.Var; b != nil; {
			chain = append(chain, b)
			bd, ok := b.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			b = bd.Var
		}
		return true
	})
	if len(chain) > 0 {
		hit := false
		syntax.Inspect(prev, func(x syntax.Node) bool {
			if hit {
				return false
			}
			if x == syntax.Node(prev) {
				return true
			}
			if p, ok := x.Parent().(*syntax.ArrayDimFetch); ok && p.Var == x {
				return true
			}
			for _, c := range chain {
				if astquery.EquivalentFoldNames(ctx.File, x, c) {
					hit = true
					break
				}
			}
			return true
		})
		if hit {
			return true
		}
	}
	// S3
	guarded := map[string]bool{}
	syntax.Inspect(prev, func(x syntax.Node) bool {
		is, ok := x.(*syntax.Isset)
		if !ok {
			return true
		}
		for _, a := range is.Vars {
			syntax.Inspect(a, func(y syntax.Node) bool {
				d, ok := y.(*syntax.ArrayDimFetch)
				if !ok {
					return true
				}
				var b syntax.Expr = d
				for {
					bd, ok := b.(*syntax.ArrayDimFetch)
					if !ok {
						break
					}
					b = bd.Var
				}
				if v, ok := astquery.PlainVariable(b); ok {
					guarded[v.Name] = true
				}
				return true
			})
		}
		return true
	})
	if len(guarded) > 0 {
		hit := false
		syntax.Inspect(cur, func(x syntax.Node) bool {
			if hit {
				return false
			}
			if v, ok := astquery.PlainVariable(astquery.AsExpr(x)); ok && guarded[v.Name] {
				hit = true
			}
			return true
		})
		if hit {
			return true
		}
	}
	return false
}

// typeGuarded implements S5 (custos): prev checks the type or the null /
// false value of a plain variable that cur uses (`false === $f`,
// `$f instanceof X`, `is_string($f)`), so cur may only be valid once prev
// has narrowed that variable. Such checks cost nothing by themselves: they
// matter inside a costlier operand (`(null !== $o && $o->ready()) && $o->id`).
func typeGuarded(ctx *analysis.Context, prev, cur syntax.Expr) bool {
	guarded := map[string]bool{}
	guard := func(e syntax.Expr) {
		if v, ok := astquery.PlainVariable(syntax.UnwrapParens(e)); ok {
			guarded[v.Name] = true
		}
	}
	syntax.Inspect(prev, func(x syntax.Node) bool {
		switch g := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Binary:
			switch g.Op.Kind {
			case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TIsEqual, syntax.TIsNotEqual:
				if notOptimalNullOrBool(g.Left) {
					guard(g.Right)
				} else if notOptimalNullOrBool(g.Right) {
					guard(g.Left)
				}
			}
		case *syntax.Instanceof:
			guard(g.Expr)
		case *syntax.FuncCall:
			name := ctx.GlobalFunctionName(g)
			if args, ok := astquery.CallArgValues(g); ok && len(args) > 0 && strings.HasPrefix(name, "is_") {
				guard(args[0])
			}
		}
		return true
	})
	hit := false
	syntax.Inspect(cur, func(x syntax.Node) bool {
		if v, ok := astquery.PlainVariable(astquery.AsExpr(x)); ok && guarded[v.Name] {
			hit = true
		}
		return !hit
	})
	return hit
}

// notOptimalNullOrBool reports whether e is the constant null, true or
// false (any case).
func notOptimalNullOrBool(e syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	if !ok || c.Name == nil {
		return false
	}
	switch strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)) {
	case "null", "true", "false":
		return true
	}
	return false
}

// notOptimalPure lists the built-in functions without observable side
// effects (S4), lower-case. ctype_* functions are pure as well.
var notOptimalPure = func() map[string]bool {
	m := map[string]bool{}
	for k := range cheapFunctions {
		m[k] = true
	}
	for _, f := range strings.Fields(`method_exists key_exists
		count sizeof strlen mb_strlen trim ltrim rtrim chop strtolower strtoupper mb_strtolower
		mb_strtoupper ucfirst lcfirst ucwords substr mb_substr substr_count strpos stripos strrpos
		strripos mb_strpos mb_stripos str_contains str_starts_with str_ends_with strstr stristr
		strrchr str_pad str_repeat strrev str_split mb_str_split sprintf vsprintf implode join
		explode strcmp strcasecmp strncmp strncasecmp strnatcmp strnatcasecmp str_replace
		str_ireplace nl2br number_format wordwrap htmlspecialchars htmlentities html_entity_decode
		strip_tags addslashes stripslashes quotemeta preg_quote preg_match preg_match_all
		preg_replace preg_split preg_grep
		in_array array_search array_keys array_values array_merge array_merge_recursive
		array_replace array_slice array_flip array_unique array_reverse array_combine array_fill
		array_fill_keys array_pad array_chunk array_diff array_diff_key array_diff_assoc
		array_intersect array_intersect_key array_column array_sum array_product
		array_count_values array_key_first array_key_last array_is_list range
		abs min max floor ceil round intdiv fmod sqrt pow intval floatval doubleval boolval
		strval gettype get_debug_type is_nan is_finite is_infinite
		get_class get_parent_class get_object_vars get_class_methods class_implements
		class_parents constant
		md5 sha1 crc32 hash base64_encode base64_decode bin2hex hex2bin json_encode json_decode
		urlencode urldecode rawurlencode rawurldecode http_build_query parse_url filter_var
		version_compare
		time microtime hrtime date gmdate mktime strtotime checkdate uniqid
		file_exists is_file is_dir is_link is_readable is_writable is_writeable is_executable
		filesize filemtime`) {
		m[f] = true
	}
	return m
}()

// notOptimalImpure reports whether e (itself or a sub-expression, not
// descending into closures) has a side effect that forbids reordering (S4).
func notOptimalImpure(ctx *analysis.Context, e syntax.Expr) bool {
	impure := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		if impure {
			return false
		}
		switch c := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.PropertyFetch:
			if classifyPropertyRead(ctx, c) == propComputed {
				impure = true // a get hook or __get() runs user code
				return false
			}
		case *syntax.Include, *syntax.Eval, *syntax.Exit, *syntax.Print,
			*syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			// Method, static and constructor calls run arbitrary code.
			impure = true
			return false
		case *syntax.FuncCall:
			if _, ok := c.Name.(*syntax.Name); !ok {
				impure = true
				return false
			}
			f := ctx.Types().ResolveFunction(c)
			if f == nil {
				impure = true
				return false
			}
			// A namespaced function never is a built-in; a global one in the
			// purity list cannot be redeclared by user code.
			name := strings.ToLower(strings.TrimPrefix(f.FQN, `\`))
			if strings.Contains(name, `\`) ||
				(!notOptimalPure[name] && !strings.HasPrefix(name, "ctype_")) {
				impure = true
				return false
			}
			if semanticquery.ArgBindsByRef(c.Args, f.Params, true) {
				impure = true
				return false
			}
		}
		return true
	})
	return impure
}

// byRefVarArgs returns the plain-variable arguments of call passed to
// by-reference parameters of the resolved callee; list is call's arguments.
func byRefVarArgs(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList) []syntax.Node {
	hasVar := false
	for _, a := range list.Args {
		if arg, ok := a.(*syntax.Arg); ok {
			if _, ok := astquery.PlainVariable(arg.Value); ok {
				hasVar = true
			}
		}
	}
	if !hasVar {
		return nil
	}
	params := calleeParams(ctx, call)
	if params == nil {
		return nil
	}
	var out []syntax.Node
	for i, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Name != nil {
			continue
		}
		v, ok := astquery.PlainVariable(arg.Value)
		if !ok {
			continue
		}
		var p *index.Param
		if i < len(params) {
			p = &params[i]
		} else if last := &params[len(params)-1]; last.Variadic {
			p = last
		}
		if p != nil && p.ByRef {
			out = append(out, v)
		}
	}
	return out
}

func calleeParams(ctx *analysis.Context, call syntax.Node) []index.Param {
	switch c := call.(type) {
	case *syntax.FuncCall:
		if f := ctx.Types().ResolveFunction(c); f != nil && len(f.Params) > 0 {
			return f.Params
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return nil
		}
		for _, cls := range ctx.TypeOf(c.Var).Classes() {
			if m := ctx.Index().FindMethod(cls, id.Value, ctx.PHP); m != nil && len(m.Params) > 0 {
				return m.Params
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		name, ok2 := c.Class.(*syntax.Name)
		if !ok || !ok2 {
			return nil
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		if m := ctx.Index().FindMethod(fqn, id.Value, ctx.PHP); m != nil && len(m.Params) > 0 {
			return m.Params
		}
	}
	return nil
}

func keywordOperators(ctx *analysis.Context, c syntax.Expr) {
	syntax.Inspect(c, func(x syntax.Node) bool {
		b, ok := x.(*syntax.Binary)
		if !ok {
			return true
		}
		switch b.Op.Kind {
		case syntax.TAnd:
			ctx.ReportSeverity(b.Op.Span, diagnostic.SeverityInfo, notOptimalAndMsg)
		case syntax.TOr:
			ctx.ReportSeverity(b.Op.Span, diagnostic.SeverityInfo, notOptimalOrMsg)
		}
		return true
	})
}

// notOptimalLiteral reports whether e (parentheses stripped) is a literal
// value: a scalar or string literal, null/true/false, an array literal, or a
// signed number.
func notOptimalLiteral(e syntax.Expr) bool {
	switch x := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal, *syntax.InterpolatedString, *syntax.Array:
		return true
	case *syntax.ConstFetch:
		switch strings.ToLower(strings.TrimPrefix(x.Name.Value, `\`)) {
		case "null", "true", "false":
			return true
		}
	case *syntax.Unary:
		return notOptimalLiteral(x.Expr)
	}
	return false
}

func equalityNextToInstanceof(ctx *analysis.Context, ops []syntax.Expr) {
	var subj syntax.Expr
	for _, o := range ops {
		if io, ok := o.(*syntax.Instanceof); ok {
			subj = io.Expr
			break
		}
	}
	if subj == nil {
		return
	}
	for _, o := range ops {
		b, ok := o.(*syntax.Binary)
		if !ok {
			continue
		}
		switch b.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
		default:
			continue
		}
		var other syntax.Expr
		switch {
		case astquery.EquivalentFoldNames(ctx.File, b.Left, subj):
			other = b.Right
		case astquery.EquivalentFoldNames(ctx.File, b.Right, subj):
			other = b.Left
		}
		// Comparing the object with another value (`$a instanceof X && $a !==
		// $b`) is an ordinary identity/equality check; only a comparison with
		// a literal (null, bool, number, string, array) is redundant or
		// suspicious next to instanceof.
		if other != nil && notOptimalLiteral(other) {
			ctx.ReportSeverity(b.Span(), diagnostic.SeverityInfo, notOptimalEqualityMsg)
		}
	}
}

type instanceofEntry struct {
	node *syntax.Instanceof
	fqn  string
}

func redundantInstanceof(ctx *analysis.Context, ops []syntax.Expr, op syntax.TokenKind) {
	var entries []instanceofEntry
	for _, o := range ops {
		io, ok := o.(*syntax.Instanceof)
		if !ok {
			continue
		}
		name, ok := io.Class.(*syntax.Name)
		if !ok {
			continue
		}
		fqn := ctx.Names().Class(name.Value, name.Span().Start)
		c := ctx.Index().Class(fqn, ctx.PHP)
		if c == nil {
			continue
		}
		entries = append(entries, instanceofEntry{io, c.FQN})
	}
	if len(entries) < 2 {
		return
	}
	used := make([]bool, len(entries))
	for i := range entries {
		if used[i] {
			continue
		}
		group := []int{i}
		used[i] = true
		for j := i + 1; j < len(entries); j++ {
			if !used[j] && astquery.EquivalentFoldNames(ctx.File, entries[i].node.Expr, entries[j].node.Expr) {
				group = append(group, j)
				used[j] = true
			}
		}
		if len(group) < 2 {
			continue
		}
		reported := map[int]bool{}
		for _, a := range group {
			anc := map[string]bool{}
			for _, c := range ctx.Index().Ancestors(entries[a].fqn, ctx.PHP) {
				anc[notOptimalClassKey(c.FQN)] = true
			}
			for _, b := range group {
				if b == a {
					continue
				}
				k := notOptimalClassKey(entries[b].fqn)
				if k == "datetimeinterface" && ctx.PHP < phpversion.PHP55 {
					continue
				}
				if anc[k] {
					// a's class is a subtype of b's: under || the specific
					// check a is subsumed, under && the broad check b is.
					redundant := a
					if op == syntax.TBooleanAnd {
						redundant = b
					}
					if !reported[redundant] {
						reported[redundant] = true
						ctx.Report(entries[redundant].node.Span(), notOptimalInstanceofMsg)
					}
				}
			}
		}
	}
}
func notOptimalClassKey(fqn string) string { return strings.ToLower(strings.TrimPrefix(fqn, `\`)) }
