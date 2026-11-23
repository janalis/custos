package deprecatedfeature

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

const since = " is deprecated as of PHP 8.6"

// hints are the replacements PHP names for what it deprecates.
var hints = map[string]string{
	"spl_classes":        "use ReflectionExtension::getClassNames()",
	"spl_object_hash":    "use spl_object_id()",
	"strcoll":            "use Collator::compare()",
	"mysqli_stmt_init":   "pass the query to mysqli_prepare()",
	"SORT_LOCALE_STRING": "sort with Collator",
}

// deprecatedMethod is a class declaring a method PHP deprecates, with the
// version that does.
type deprecatedMethod struct {
	class string
	since phpversion.Version
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP.Below(phpversion.PHP80) {
		return
	}
	// What PHP's own functions, methods and constants deprecate, by version.
	switch x := n.(type) {
	case *syntax.FuncCall:
		checkCall(ctx, x)
		return
	case *syntax.MethodCall:
		checkMethod(ctx, x)
		return
	case *syntax.StaticCall:
		checkStaticMethod(ctx, x)
		return
	case *syntax.ClassConstFetch:
		checkClassConstant(ctx, x)
		return
	case *syntax.ConstFetch:
		checkConstant(ctx, x)
		return
	}
	// The language deprecations of PHP 8.6.
	if ctx.PHP.Below(phpversion.PHP86) {
		return
	}
	switch x := n.(type) {
	case *syntax.New:
		checkNew(ctx, x)
	case *syntax.Return:
		checkReturn(ctx, x)
	case *syntax.Yield, *syntax.YieldFrom:
		if name := structor(x); name != "" {
			ctx.ReportNode(x, "Remove this yield from "+name+"(): making it a generator"+since+".")
		}
	case *syntax.Method:
		checkDebugInfo(ctx, x)
	case *syntax.ClassConst:
		for _, item := range x.Consts {
			if strings.EqualFold(item.Name.Value, "namespace") {
				ctx.ReportNode(item.Name, "Rename the class constant 'namespace': this name"+since+".")
			}
		}
	case *syntax.ClassLike:
		if x.Name != nil && reservedName(x.Name.Value) {
			ctx.ReportNode(x.Name, nameMessage(x.Name.Value))
		}
	case *syntax.Function:
		if reservedName(x.Name.Value) || strings.EqualFold(x.Name.Value, "readonly") {
			ctx.ReportNode(x.Name, nameMessage(x.Name.Value))
		}
	case *syntax.ConstStmt:
		for _, item := range x.Consts {
			if reservedName(item.Name.Value) || item.Name.Value == "_" {
				ctx.ReportNode(item.Name, nameMessage(item.Name.Value))
			}
		}
	case *syntax.Use:
		checkUse(ctx, x)
	}
}

// checkCall covers PHP's deprecated functions and the call shapes 8.6
// deprecates: define()'s third argument, a class name given to is_a() or
// is_subclass_of() without allow_string, an object walked as an array.
func checkCall(ctx *analysis.Context, call *syntax.FuncCall) {
	// A partial application (`spl_object_hash(?)`) refers to the function
	// too: calling the closure raises the deprecation.
	name := semanticquery.NativeBuiltinName(ctx, call)
	if v, ok := deprecatedFunctions[name]; ok {
		if ctx.PHP.AtLeast(v) && !semanticquery.ExistenceChecked(ctx, call, name) {
			ctx.ReportNode(call.Name, replace(name+"()", hints[name], v))
		}
		return
	}
	if ctx.PHP.Below(phpversion.PHP86) {
		return // the call shapes below are PHP 8.6's
	}
	args := call.Args
	switch name {
	case "define":
		if third := semanticquery.CallArgument(args, 2, "case_insensitive"); third != nil {
			ctx.ReportNode(third, "Remove define()'s third argument: passing it"+since+".")
		}
		if declared, ok := literal(semanticquery.CallArgument(args, 0, "constant_name")); ok {
			short := declared[strings.LastIndexByte(declared, '\\')+1:]
			if reservedName(short) || short == "_" {
				ctx.ReportNode(semanticquery.CallArgument(args, 0, "constant_name"), nameMessage(short))
			}
		}
	case "class_alias":
		if alias, ok := literal(semanticquery.CallArgument(args, 1, "alias")); ok {
			if short := alias[strings.LastIndexByte(alias, '\\')+1:]; reservedName(short) {
				ctx.ReportNode(semanticquery.CallArgument(args, 1, "alias"), nameMessage(short))
			}
		}
	case "is_a", "is_subclass_of":
		first := semanticquery.CallArgument(args, 0, "object_or_class")
		if first == nil || !ctx.TypeOf(first).OnlyOf("string") {
			return
		}
		value, state := semanticquery.CallArgumentBinding(args, 2, "allow_string")
		omitted := state == semanticquery.ArgumentMissing && name == "is_a"
		if omitted || state == semanticquery.ArgumentBound && falseConstant(value) {
			ctx.ReportNode(first, "Pass allow_string: true to "+name+"(): giving it a class name otherwise"+since+".")
		}
	case "array_walk", "array_walk_recursive", "http_build_query":
		param := "array"
		if name == "http_build_query" {
			param = "data"
		}
		if first := semanticquery.CallArgument(args, 0, param); first != nil && object(ctx, first) {
			ctx.ReportNode(first, "Pass an array to "+name+"(): giving it an object"+since+"; use get_object_vars().")
		}
	}
}

// checkMethod covers PHP's deprecated methods, on a receiver that is known
// to run PHP's own declaration.
func checkMethod(ctx *analysis.Context, call *syntax.MethodCall) {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	for _, d := range deprecatedMethods[strings.ToLower(id.Value)] {
		if ctx.PHP.AtLeast(d.since) && semanticquery.NativeMethod(ctx, call, d.class, id.Value) {
			if !semanticquery.ExistenceChecked(ctx, call, d.class) {
				ctx.ReportNode(id, replace(d.class+"::"+id.Value+"()", "", d.since))
			}
			return
		}
	}
}

// checkStaticMethod covers a deprecated method called through a class
// name, self or parent (`parent::__wakeup()`), when PHP's own declaration
// runs.
func checkStaticMethod(ctx *analysis.Context, call *syntax.StaticCall) {
	id, ok := call.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	class := semanticquery.StaticCallClass(ctx, call.Class)
	for _, d := range deprecatedMethods[strings.ToLower(id.Value)] {
		m := ctx.Index().FindMethod(class, id.Value, ctx.PHP)
		if ctx.PHP.AtLeast(d.since) && m != nil && m.Builtin && strings.EqualFold(strings.TrimPrefix(m.Class, `\`), d.class) {
			if !semanticquery.ExistenceChecked(ctx, call, d.class) {
				ctx.ReportNode(id, replace(d.class+"::"+id.Value+"()", "", d.since))
			}
			return
		}
	}
}

// checkClassConstant covers PHP's deprecated class constants, declared by
// PHP's own class (`PDO::MYSQL_ATTR_COMPRESS`).
func checkClassConstant(ctx *analysis.Context, c *syntax.ClassConstFetch) {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	class := semanticquery.StaticCallClass(ctx, c.Class)
	k := ctx.Index().FindConst(class, id.Value, ctx.PHP)
	if k == nil {
		return
	}
	owner := strings.TrimPrefix(k.Class, `\`)
	if v, ok := deprecatedClassConstants[strings.ToLower(owner)+"::"+id.Value]; ok && ctx.PHP.AtLeast(v) && builtinClass(ctx, owner) && !semanticquery.ExistenceChecked(ctx, c, owner) {
		ctx.ReportNode(c, replace(owner+"::"+id.Value, "", v))
	}
}

func builtinClass(ctx *analysis.Context, fqn string) bool {
	c := ctx.Index().Class(fqn, ctx.PHP)
	return c != nil && strings.HasPrefix(c.File, "stubs/")
}

// checkNew reports a mysqli_stmt built without its query.
func checkNew(ctx *analysis.Context, n *syntax.New) {
	name, ok := n.Class.(*syntax.Name)
	if !ok || !strings.EqualFold(strings.TrimPrefix(ctx.Names().Class(name.Value, name.Span().Start), `\`), "mysqli_stmt") {
		return
	}
	if !builtinClass(ctx, "mysqli_stmt") {
		return
	}
	if semanticquery.CallArgumentOmitted(n.Args, 1, "query") {
		ctx.ReportNode(n, "Pass the query to new mysqli_stmt(): building it without one"+since+".")
	}
}

// checkConstant reports PHP's deprecated constants, unless the project
// declares one of that name in the file's namespace.
func checkConstant(ctx *analysis.Context, c *syntax.ConstFetch) {
	short := strings.TrimPrefix(c.Name.Value, `\`)
	v, ok := deprecatedConstants[short]
	if !ok || ctx.PHP.Below(v) {
		return
	}
	fqn, fallback := ctx.Names().Const(c.Name.Value, c.Name.Span().Start)
	known := ctx.Index().Constant(fqn, ctx.PHP)
	if known == nil && fallback != "" {
		known = ctx.Index().Constant(fallback, ctx.PHP)
	}
	if known != nil && known.Builtin && !semanticquery.ExistenceChecked(ctx, c, short) {
		ctx.ReportNode(c, replace(short, hints[short], v))
	}
}

// checkReturn reports a return in a finally block of the same function, and
// a returned value in a constructor or destructor.
func checkReturn(ctx *analysis.Context, r *syntax.Return) {
	for p := r.Parent(); p != nil && !syntax.IsFuncLike(p); p = p.Parent() {
		if _, ok := p.(*syntax.Finally); ok {
			ctx.ReportNode(r, "Move this return out of the finally block: returning there"+since+".")
			return
		}
	}
	if name := structor(r); name != "" && r.Expr != nil {
		ctx.ReportNode(r, "Remove the value returned from "+name+"(): returning one"+since+".")
	}
}

// structor names the constructor or destructor directly enclosing n, if
// any.
func structor(n syntax.Node) string {
	m, ok := syntax.EnclosingFuncLike(n).(*syntax.Method)
	if !ok || m.Name == nil {
		return ""
	}
	if name := strings.ToLower(m.Name.Value); name == "__construct" || name == "__destruct" {
		return name
	}
	return ""
}

// checkDebugInfo reports a nullable return type on __debugInfo().
func checkDebugInfo(ctx *analysis.Context, m *syntax.Method) {
	if m.Name == nil || !strings.EqualFold(m.Name.Value, "__debugInfo") || !nullable(m.ReturnType) {
		return
	}
	ctx.ReportNode(m.ReturnType, "Declare __debugInfo() as returning array: allowing null"+since+".")
}

func nullable(t syntax.Expr) bool {
	switch x := t.(type) {
	case *syntax.NullableType:
		return true
	case *syntax.Name:
		return strings.EqualFold(x.Value, "null")
	case *syntax.UnionType:
		for _, member := range x.Types {
			if name, ok := member.(*syntax.Name); ok && strings.EqualFold(name.Value, "null") {
				return true
			}
		}
	}
	return false
}

// checkUse reports an import named let or is, or a class or constant
// imported as _.
func checkUse(ctx *analysis.Context, u *syntax.Use) {
	for _, item := range u.Items {
		node, name := syntax.Node(item.Alias), ""
		if item.Alias != nil {
			name = item.Alias.Value
		} else {
			node, name = item.Name, item.Name.Value[strings.LastIndexByte(item.Name.Value, '\\')+1:]
		}
		kind := u.Type
		if item.Type != syntax.UseNormal {
			kind = item.Type
		}
		if reservedName(name) || name == "_" && kind != syntax.UseFunction {
			ctx.ReportNode(node, nameMessage(name))
		}
	}
}

// reservedName reports the names PHP 8.6 deprecates for declarations, in
// any case.
func reservedName(name string) bool {
	return strings.EqualFold(name, "let") || strings.EqualFold(name, "is")
}

func nameMessage(name string) string {
	return "Rename '" + name + "': this name" + since + "."
}

// replace asks to replace what PHP deprecates as of version v, with PHP's
// replacement when it names one.
func replace(what, hint string, v phpversion.Version) string {
	msg := "Replace " + what + ", deprecated as of PHP " + v.String()
	if hint == "" {
		return msg + "."
	}
	return msg + ": " + hint + "."
}

func literal(e syntax.Expr) (string, bool) {
	lit, ok := syntax.UnwrapParens(e).(*syntax.Literal)
	if !ok {
		return "", false
	}
	return syntax.LiteralStringValue(lit)
}

// object reports a value known to be an object.
func object(ctx *analysis.Context, e syntax.Expr) bool {
	t := ctx.TypeOf(e)
	return t.OnlyOf(append(t.Classes(), "object")...)
}

// falseConstant reports the constant false, in any case.
func falseConstant(e syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	return ok && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false")
}
