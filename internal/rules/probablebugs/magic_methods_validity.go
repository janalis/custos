package probablebugs

import (
	"strconv"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// magicMethodsValidity validates the contracts of magic methods and the use
// of the `__` prefix.
type magicMethodsValidity struct{}

func init() { register(magicMethodsValidity{}) }

func (magicMethodsValidity) ID() string { return "MagicMethodsValidity" }

func (magicMethodsValidity) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

// Semantic marks the rule as needing the project index.
func (magicMethodsValidity) Semantic() {}

// magicCanonical maps the lower-cased magic method names to their canonical
// spelling: PHP method names are case-insensitive, so dispatch is too.
var magicCanonical = func() map[string]string {
	m := map[string]string{}
	for _, n := range []string{"__construct", "__destruct", "__clone", "__get", "__isset", "__unset", "__set",
		"__call", "__callStatic", "__toString", "__debugInfo", "__set_state", "__invoke", "__wakeup",
		"__unserialize", "__sleep", "__serialize", "__autoload"} {
		m[strings.ToLower(n)] = n
	}
	return m
}()

// magicLookup returns set[name] with a case-insensitive key match.
func magicLookup(set map[string]bool, name string) bool {
	for k := range set {
		if strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}

var magicKnownNonMagic = map[string]bool{
	"__": true, "__inject": true, "__prepare": true, "__toArray": true,
	"__doRequest": true, "__getCookies": true, "__getFunctions": true, "__getLastRequest": true,
	"__getLastRequestHeaders": true, "__getLastResponse": true, "__getLastResponseHeaders": true,
	"__getTypes": true, "__setCookie": true, "__setLocation": true, "__setSoapHeaders": true, "__soapCall": true,
}

var magicMissingUnderscore = map[string]bool{
	"_construct": true, "_destruct": true, "_call": true, "_callStatic": true, "_get": true, "_set": true,
	"_isset": true, "_unset": true, "_sleep": true, "_wakeup": true, "_toString": true, "_invoke": true,
	"_set_state": true, "_clone": true, "_debugInfo": true,
}

type magicCheck struct {
	ctx  *analysis.Context
	m    *syntax.Method
	cl   *syntax.ClassLike
	fqn  string // containing class FQN without leading backslash ("" for anonymous)
	name string
}

func (c *magicCheck) report(msg string, fixes ...analysis.Fix) {
	c.ctx.ReportNode(c.m.Name, msg, fixes...)
}

func (magicMethodsValidity) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	cl := m.Parent().(*syntax.ClassLike)       // methods only appear in class bodies
	if !strings.HasPrefix(m.Name.Value, "_") { // also skips recovery (empty) names
		return
	}
	if cl.ClassKind == syntax.KindInterface || m.Modifiers.Has(syntax.TAbstract) {
		return
	}
	c := &magicCheck{ctx: ctx, m: m, cl: cl, fqn: ctx.Names().DeclFQN(cl), name: m.Name.Value}
	canonical := magicCanonical[strings.ToLower(c.name)]
	switch canonical {
	case "__construct": // D1
		c.notStatic()
		c.noReturn()
		if !ctx.IsTestFile() && !util.IsTestClassFQN(c.fqn) {
			c.callsParent()
		}
	case "__destruct", "__clone": // D2
		c.notStatic()
		c.noReturn()
		c.noArgs()
		c.callsParent()
	case "__get", "__isset", "__unset": // D3
		c.argc(1)
		c.notStatic()
		c.public()
		c.noByRef()
		c.pair("__set")
	case "__set": // D4
		c.argc(2)
		c.notStatic()
		c.public()
		c.noByRef()
		c.pair("__isset")
		c.pair("__get")
	case "__call": // D5
		c.argc(2)
		c.notStatic()
		c.public()
		c.noByRef()
	case "__callStatic": // D6
		c.argc(2)
		c.mustStatic()
		c.public()
		c.noByRef()
	case "__toString": // D7
		c.notStatic()
		c.noArgs()
		c.public()
		c.returns("string")
	case "__debugInfo": // D8
		c.notStatic()
		c.noArgs()
		c.public()
		c.returns("array", "null")
		if ctx.PHP < phpver.PHP56 {
			c.report(c.name + " only exists from PHP 5.6; it is never called here.")
		}
	case "__set_state": // D9
		c.argc(1)
		c.mustStatic()
		c.public()
		if c.fqn != "" {
			c.returns(`\`+c.fqn, "static")
		}
	case "__invoke": // D10
		c.notStatic()
		c.public()
	case "__wakeup": // D11
		c.notStatic()
		c.noArgs()
		c.noReturn()
	case "__unserialize": // D12
		c.notStatic()
		c.public()
		c.argc(1)
		c.noReturn()
	case "__sleep", "__serialize": // D13
		c.notStatic()
		c.public()
		c.noArgs()
		c.returns("array")
	case "__autoload": // D14
		c.argc(1)
		c.noReturn()
		c.report("__autoload is deprecated since PHP 7.2; use spl_autoload_register().")
	default: // D15
		if strings.HasPrefix(c.name, "__") && !magicLookup(magicKnownNonMagic, c.name) {
			// custos: a name imposed by an interface or parent class
			// (Restler's iFilter::__isAllowed()) cannot be renamed here;
			// the declaring ancestor is reported instead.
			if c.inherited() {
				return
			}
			c.report("The '__' prefix is reserved for magic methods.")
			return
		}
		// custos: no rename fix — it would break every caller of the
		// method, and the signature may not fit the magic contract.
		if magicLookup(magicMissingUnderscore, c.name) && !c.inherited() && !c.deliberateHook() {
			c.report("'" + c.name + "' is not magic; did you mean '_" + c.name + "'?")
		}
	}
}

// inherited reports whether a parent, interface or trait of the class
// declares the method too: the name is then imposed by the hierarchy (a
// `_get()` hook of an abstract cache class), not a misspelt magic method.
func (c *magicCheck) inherited() bool {
	if c.fqn == "" {
		return false
	}
	lname := strings.ToLower(c.name)
	for i, a := range c.ctx.Index().Ancestors(c.fqn, c.ctx.PHP) {
		if _, ok := a.Methods[lname]; ok && i > 0 {
			return true
		}
	}
	return false
}

// deliberateHook reports whether a single-underscore name is an explicit
// hook rather than a misspelt magic method (custos): the class hierarchy
// also declares the real magic method (`__construct()` calling
// `$this->_construct()`), or the file calls the method by name.
func (c *magicCheck) deliberateHook() bool {
	lname := strings.ToLower(c.name)
	if c.fqn != "" {
		for _, a := range c.ctx.Index().Ancestors(c.fqn, c.ctx.PHP) {
			if _, ok := a.Methods["_"+lname]; ok {
				return true
			}
		}
	}
	called := c.ctx.Memo("called-methods", func() any {
		names := map[string]bool{}
		visit := func(n syntax.Node) bool {
			var name syntax.Expr
			switch x := n.(type) {
			case *syntax.MethodCall:
				name = x.Name
			case *syntax.StaticCall:
				name = x.Name
			}
			if id, ok := name.(*syntax.Identifier); ok {
				names[strings.ToLower(id.Value)] = true
			}
			return true
		}
		for _, st := range c.ctx.File.Stmts {
			syntax.Inspect(st, visit)
		}
		return names
	}).(map[string]bool)
	return called[lname]
}

func (c *magicCheck) notStatic() {
	if c.m.Modifiers.Has(syntax.TStatic) {
		c.report(c.name + " must not be static.")
	}
}

func (c *magicCheck) mustStatic() {
	if !c.m.Modifiers.Has(syntax.TStatic) {
		c.report(c.name + " must be declared static.")
	}
}

func (c *magicCheck) public() {
	if c.m.Modifiers.Has(syntax.TProtected) || c.m.Modifiers.Has(syntax.TPrivate) {
		c.report(c.name + " must be declared public.")
	}
}

func (c *magicCheck) noArgs() {
	if len(c.m.Params) > 0 {
		c.report(c.name + " must not declare parameters.")
	}
}

func (c *magicCheck) argc(n int) {
	if len(c.m.Params) != n {
		c.report(c.name + " must declare exactly " + strconv.Itoa(n) + " parameter(s).")
	}
}

func (c *magicCheck) noByRef() {
	for _, p := range c.m.Params {
		if p.ByRef {
			c.report(c.name + " must not take parameters by reference.")
			return
		}
	}
}

// returnsOf lists the return statements of the method body at any depth;
// own reports whether the statement belongs to the method itself.
func (c *magicCheck) returnsOf(fn func(r *syntax.Return, own bool)) {
	if c.m.Body == nil {
		return
	}
	syntax.Inspect(c.m.Body, func(n syntax.Node) bool {
		if r, ok := n.(*syntax.Return); ok {
			fn(r, syntax.EnclosingFuncLike(r) == syntax.Node(c.m))
		}
		return true
	})
}

func (c *magicCheck) noReturn() {
	c.returnsOf(func(r *syntax.Return, own bool) {
		if own && r.Expr != nil {
			c.ctx.ReportNode(r, c.name+" must not return a value.")
		}
	})
}

func (c *magicCheck) pair(companion string) {
	lc := strings.ToLower(companion)
	for _, mem := range c.cl.Members {
		if m, ok := mem.(*syntax.Method); ok && m.Name != nil && strings.ToLower(m.Name.Value) == lc {
			return
		}
	}
	ix := c.ctx.Index()
	var refs []string
	if p := c.ctx.Names().ParentFQN(c.cl); p != "" {
		refs = append(refs, p)
	}
	for _, mem := range c.cl.Members {
		if tu, ok := mem.(*syntax.TraitUse); ok {
			for _, t := range tu.Traits {
				refs = append(refs, c.ctx.Names().Class(t.Value, t.Span().Start))
			}
		}
	}
	for _, r := range refs {
		if ix.FindMethod(r, companion, c.ctx.PHP) != nil {
			return
		}
	}
	c.report(c.name + " needs a companion " + companion + " method.")
}

func (c *magicCheck) hasOverride() bool {
	for _, g := range c.m.Attrs {
		for _, a := range g.Attrs {
			if a.Name != nil && strings.EqualFold(c.ctx.Names().Class(a.Name.Value, a.Name.Span().Start), "Override") {
				return true
			}
		}
	}
	return false
}

func (c *magicCheck) callsParent() {
	if c.hasOverride() {
		return
	}
	parent := c.ctx.Names().ParentFQN(c.cl)
	if parent == "" {
		return
	}
	ix := c.ctx.Index()
	if ix.Class(parent, c.ctx.PHP) == nil {
		return
	}
	pm := ix.FindMethod(parent, c.name, c.ctx.PHP)
	if pm == nil || pm.Abstract || pm.Visibility == index.Private || pm.EmptyBody {
		return
	}
	called := false
	if c.m.Body != nil {
		syntax.Inspect(c.m.Body, func(n syntax.Node) bool {
			if called {
				return false
			}
			var name syntax.Expr
			switch x := n.(type) {
			case *syntax.MethodCall:
				name = x.Name
			case *syntax.StaticCall:
				name = x.Name
			case *syntax.FuncCall:
				// custos: call_user_func_array('Base::__construct', $args),
				// [$this, 'parent::__construct'], parent::class . '::__construct'
				called = c.callableNamesParent(x)
			}
			if id, ok := name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, c.name) {
				called = true
			}
			return true
		})
	}
	if called || (pm.StoresParams && c.setsAll(pm.Stores)) {
		return
	}
	declaring := strings.TrimPrefix(pm.Class, `\`)
	declaring = util.LastNamePart(declaring)
	c.report(c.name + " does not call " + declaring + "::" + c.name + "().")
}

// callableNamesParent reports whether call is call_user_func[_array]() with
// a callable whose string part ends in `::<method name>`.
func (c *magicCheck) callableNamesParent(call *syntax.FuncCall) bool {
	if fn := c.ctx.GlobalFunctionName(call); fn != "call_user_func" && fn != "call_user_func_array" {
		return false
	}
	args, ok := util.CallArgValues(call)
	if !ok || len(args) == 0 {
		return false
	}
	found := false
	syntax.Inspect(args[0], func(n syntax.Node) bool {
		if e, ok := n.(syntax.Expr); ok {
			v, ok := util.QuotedStringValue(e)
			found = ok && strings.HasSuffix(strings.ToLower(v), "::"+strings.ToLower(c.name))
		}
		return !found
	})
	return found
}

// setsAll reports whether the method sets every property of props itself
// (`$this->p = …` anywhere in its body, or a promoted parameter p): a
// parent constructor that only stores its parameters into those
// properties has nothing left to do (custos, see Divergences).
func (c *magicCheck) setsAll(props []string) bool {
	set := map[string]bool{}
	for _, p := range c.m.Params {
		if len(p.Modifiers) > 0 {
			set[p.Var.Name] = true
		}
	}
	if c.m.Body != nil {
		syntax.Inspect(c.m.Body, func(n syntax.Node) bool {
			if a, ok := n.(*syntax.Assign); ok {
				if pf, ok := a.Var.(*syntax.PropertyFetch); ok {
					if v, ok := pf.Var.(*syntax.Variable); ok && v.Name == "this" {
						if id, ok := pf.Name.(*syntax.Identifier); ok {
							set[id.Value] = true
						}
					}
				}
			}
			return true
		})
	}
	for _, p := range props {
		if !set[p] {
			return false
		}
	}
	return true
}

// normaliseMagicType maps self to the containing class and typed arrays
// (`string[]`) to array.
func (c *magicCheck) normaliseMagicType(t types.Type) []string {
	var out []string
	for _, a := range t.Atoms() {
		if a == "self" && c.fqn != "" {
			a = `\` + c.fqn
		}
		if strings.HasSuffix(a, "[]") {
			a = "array"
		}
		out = append(out, a)
	}
	return out
}

func (c *magicCheck) returns(allowed ...string) {
	ok := map[string]bool{}
	var shown []string
	for _, a := range allowed {
		ok[a] = true
		if a != "static" {
			shown = append(shown, a)
		}
	}
	msg := func(got []string) string {
		return c.name + " must return " + strings.Join(shown, "|") + "; got '" + strings.Join(got, "|") + "'."
	}
	offending := func(atoms []string) []string {
		var bad []string
		for _, a := range atoms {
			// custos: a declared `never` (always throws) satisfies any
			// return contract.
			if !ok[a] && a != "never" && !c.subtypeOfAllowed(a, allowed) {
				bad = append(bad, a)
			}
		}
		return bad
	}
	if c.m.ReturnType != nil {
		at := c.m.ReturnType.Span().Start
		t := types.FromNode(c.m.ReturnType, func(w string) string { return c.ctx.Names().Class(w, at) })
		if bad := offending(c.normaliseMagicType(t)); len(bad) > 0 {
			c.report(msg(bad))
		}
		return
	}
	any := false
	c.returnsOf(func(*syntax.Return, bool) { any = true })
	if !any {
		// A body that always throws (or exits) returns nothing either way.
		if c.m.Body == nil || !syntax.Terminates(c.m.Body) {
			c.report(msg(nil))
		}
		return
	}
	c.returnsOf(func(r *syntax.Return, own bool) {
		if !own {
			return // see spec divergences: nested function-likes are ignored
		}
		if r.Expr == nil {
			c.ctx.ReportNode(r, msg(nil))
			return
		}
		t := c.ctx.TypeOf(syntax.UnwrapParens(r.Expr))
		if t.IsUnknown() || t.Has("mixed") { // mixed may well be the right type
			return
		}
		if bad := offending(c.normaliseMagicType(t)); len(bad) > 0 {
			c.ctx.ReportNode(r, msg(bad))
		}
	})
}

// subtypeOfAllowed reports whether the class type atom a extends or
// implements one of the allowed class types (a subclass instance satisfies
// a class return contract).
func (c *magicCheck) subtypeOfAllowed(a string, allowed []string) bool {
	if !strings.HasPrefix(a, `\`) || strings.HasSuffix(a, "]") {
		return false
	}
	for _, t := range allowed {
		if strings.HasPrefix(t, `\`) && c.ctx.Index().IsSubtype(a, t, c.ctx.PHP) {
			return true
		}
	}
	return false
}
