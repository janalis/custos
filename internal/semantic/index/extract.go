package index

import (
	"bytes"
	"slices"
	"sort"
	"strings"

	"github.com/janalis/custos/internal/php/phpdoc"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/names"
	"github.com/janalis/custos/internal/semantic/types"
)

type extractor struct {
	docClasses map[string]*Class
	stub       bool
	f          *syntax.File
	r          *names.Resolver
	out        *FileSymbols
	templates  map[string]bool   // @template names in scope (class + member)
	classTpl   map[string]bool   // class-level @template names of the class being extracted
	methodTpl  map[string]bool   // @template names of the method being extracted
	aliases    map[string]string // @phpstan-type / import-type names in scope (imported: importedAlias)
}

// importedAlias marks in extractor.aliases a type alias the class imports:
// its types name it AliasMark+name until the index links it.
const importedAlias = "\x00import"

// AliasMark prefixes the class name standing for a type alias a class
// imports in the types the index records for it (`\__custos_alias__Name`),
// until the index links it (see Index.LinkTypeAliases).
const AliasMark = "__custos_alias__"

// Extract collects the symbols declared in f.
func Extract(f *syntax.File) *FileSymbols {
	return extractFile(f, false)
}

// ExtractStubs collects symbols with PHP stub availability annotations enabled.
func ExtractStubs(f *syntax.File) *FileSymbols {
	return extractFile(f, true)
}

// bindsGlobal records a global variable a function-like of the file binds;
// an unnamed one stands for any of them.
func (x *extractor) bindsGlobal(name string, named bool) {
	if !named {
		name = ""
	}
	if !slices.Contains(x.out.Globals, name) {
		x.out.Globals = append(x.out.Globals, name)
	}
}

// globalsEntry names the global a literal key of the globals table reads.
func globalsEntry(table *syntax.Variable) (string, bool) {
	entry, ok := table.Parent().(*syntax.ArrayDimFetch)
	if !ok || entry.Var != syntax.Expr(table) {
		return "", false
	}
	key, ok := syntax.UnwrapParens(entry.Dim).(*syntax.Literal)
	if !ok || key.LitKind != syntax.LitString {
		return "", false
	}
	return syntax.LiteralStringValue(key)
}

func extractFile(f *syntax.File, stub bool) *FileSymbols {
	x := &extractor{stub: stub, f: f, r: names.New(f), out: &FileSymbols{Path: f.Path}}
	var calls []*syntax.FuncCall
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ClassLike:
			x.class(n)
		case *syntax.Function:
			x.function(n)
		case *syntax.ConstStmt:
			for _, c := range n.Consts {
				fqn := c.Name.Value
				if ns := x.r.Namespace(c.Span().Start); ns != "" {
					fqn = ns + `\` + c.Name.Value
				}
				x.out.Constants = append(x.out.Constants, &Constant{FQN: fqn, Value: x.text(c.Value), ValueType: constArrayDoc(c.Value, nil), File: f.Path, Span: c.Span()})
			}
		case *syntax.FuncCall:
			switch x.declarationBuiltin(n, nil) {
			case "":
			case prefixTest:
				x.loaderTest(n)
			default:
				calls = append(calls, n)
			}

		case *syntax.Global:
			if syntax.EnclosingVariableScope(n) != nil {
				// A name chosen at run time is not counted, as within a file:
				// only a use of the globals table may reach any of them.
				for _, v := range n.Vars {
					if variable, ok := v.(*syntax.Variable); ok && variable.NameExpr == nil {
						x.bindsGlobal(variable.Name, true)
					}
				}
			}
		case *syntax.Variable:
			if n.Name == "GLOBALS" && n.NameExpr == nil && syntax.EnclosingVariableScope(n) != nil {
				x.bindsGlobal(globalsEntry(n))
			}
		}
		return true
	})
	// Rare: mark the loaders naming a phar.
	if bytes.Contains(f.Src, pharScheme) {
		x.pharLoaders()
	}
	if len(calls) > 0 {
		declared := make(map[string]bool, len(x.out.Functions))
		for _, fn := range x.out.Functions {
			declared[key(fn.FQN)] = true
		}
		var loaders []*syntax.FuncCall
		for _, call := range calls {
			switch x.declarationBuiltin(call, declared) {
			case "define":
				x.define(call)
			case "class_alias":
				x.classAlias(call)
			case "spl_autoload_register":
				ns := strings.ToLower(x.r.Namespace(call.Span().Start))
				if ns != "" {
					ns += `\`
				}
				if !slices.Contains(x.out.Autoloaders, ns) {
					x.out.Autoloaders = append(x.out.Autoloaders, ns)
				}
				loaders = append(loaders, call)
			}
		}
		// Rare: read the autoloaders' callbacks.
		if len(loaders) > 0 {
			x.autoloaders(loaders)
		}
		// Deferred define calls must retain source order relative to const
		// declarations: duplicate declarations use the first indexed symbol.
		sort.SliceStable(x.out.Constants, func(i, j int) bool {
			return x.out.Constants[i].Span.Start < x.out.Constants[j].Span.Start
		})
	}
	return x.out
}

// maxAutoloadPrefixes caps the prefixes kept for one file, or for one of
// its loaders, and the callbacks kept for one file.
const maxAutoloadPrefixes = 16

// pharScheme starts a path inside a phar.
var pharScheme = []byte("phar://")

// autoloaders reads the callbacks the spl_autoload_register() calls of the
// file register: for closures and arrow functions, the class name
// prefixes they test and whether they name a phar; for functions and
// methods named by a string, an array or a first-class callable, which
// one (see Index.LoaderFacts).
func (x *extractor) autoloaders(calls []*syntax.FuncCall) {
	for _, call := range calls {
		body, param, k := x.callback(call)
		if k != "" && len(x.out.AutoloadCallbacks) < maxAutoloadPrefixes && !slices.Contains(x.out.AutoloadCallbacks, k) {
			x.out.AutoloadCallbacks = append(x.out.AutoloadCallbacks, k)
		}
		if body == nil {
			continue
		}
		syntax.Inspect(body, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.FuncCall:
				if p, ok := prefixTested(n, param); ok && len(x.out.AutoloadPrefixes) < maxAutoloadPrefixes && !slices.Contains(x.out.AutoloadPrefixes, p) {
					x.out.AutoloadPrefixes = append(x.out.AutoloadPrefixes, p)
				}
			case *syntax.Literal, *syntax.InterpolatedString:
				sp := n.Span()
				x.out.Phar = x.out.Phar || bytes.Contains(x.f.Src[sp.Start:sp.End], pharScheme)
			}
			return true
		})
	}
}

// prefixTest is what declarationBuiltin returns for strpos() and
// str_starts_with(), which a loader tests its class name with.
const prefixTest = "prefix test"

// callableKey returns the Loaders key of a named function or method
// declaration ("" for another node, or a method of an anonymous class).
func (x *extractor) callableKey(n syntax.Node) string {
	switch d := n.(type) {
	case *syntax.Function:
		return key(join(x.r.Namespace(d.Span().Start), d.Name.Value))
	case *syntax.Method:
		if c := x.classFQN(d); c != "" {
			return c + "::" + key(d.Name.Value)
		}
	}
	return ""
}

// loaderTest records the prefix call tests, when it tests the first
// parameter of the named function or method it lies in for a literal one.
func (x *extractor) loaderTest(call *syntax.FuncCall) {
	fn := syntax.EnclosingFuncLike(call)
	params := syntax.FuncLikeParams(fn)
	if len(params) == 0 || params[0].Var == nil {
		return
	}
	if p, ok := prefixTested(call, params[0].Var.Name); ok {
		if l := x.loader(fn); l != nil && len(l.Prefixes) < maxAutoloadPrefixes && !slices.Contains(l.Prefixes, p) {
			l.Prefixes = append(l.Prefixes, p)
		}
	}
}

// pharLoaders marks the named functions and methods with a string naming
// a phar in their own body (not in a closure they define).
func (x *extractor) pharLoaders() {
	syntax.InspectFile(x.f, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.Literal, *syntax.InterpolatedString:
			sp := n.Span()
			if bytes.Contains(x.f.Src[sp.Start:sp.End], pharScheme) {
				if l := x.loader(syntax.EnclosingFuncLike(n)); l != nil {
					l.Phar = true
				}
			}
			return false
		}
		return true
	})
}

// loader returns the Loader facts of the named function or method fn,
// created on first use (nil for another function-like).
func (x *extractor) loader(fn syntax.Node) *Loader {
	k := x.callableKey(fn)
	if k == "" {
		return nil
	}
	if x.out.Loaders == nil {
		x.out.Loaders = map[string]*Loader{}
	}
	l := x.out.Loaders[k]
	if l == nil {
		l = &Loader{}
		x.out.Loaders[k] = l
	}
	return l
}

// join qualifies name with namespace ns.
func join(ns, name string) string {
	if ns == "" {
		return name
	}
	return ns + `\` + name
}

// classFQN returns the key of the named class n lies in ("" outside one
// or in an anonymous class).
func (x *extractor) classFQN(n syntax.Node) string {
	c := syntax.EnclosingClass(n)
	if c == nil || c.Name == nil {
		return ""
	}
	return key(join(x.r.Namespace(c.Span().Start), c.Name.Value))
}

// classOf returns the key of the class e names in a callable written at
// call: `$this`, `self`, `static`, `__CLASS__`, a name, a `::class`
// fetch, a `new` or a string ("" when unknown).
func (x *extractor) classOf(e syntax.Expr, call syntax.Node) string {
	switch v := e.(type) {
	case *syntax.Variable:
		if v.Name == "this" && v.NameExpr == nil {
			return x.classFQN(call)
		}
	case *syntax.MagicConst:
		if v.Token.Kind == syntax.TClassC {
			return x.classFQN(call)
		}
	case *syntax.Name:
		if c, ok := x.relative(v.Value, call); ok {
			return c
		}
		return key(x.r.Class(v.Value, v.Span().Start))
	case *syntax.ClassConstFetch:
		if id, ok := v.Name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, "class") {
			return x.classOf(v.Class, call)
		}
	case *syntax.New:
		return x.classOf(v.Class, call)
	case *syntax.Literal:
		s, _ := syntax.LiteralStringValue(v)
		return x.stringClass(s, call)
	}
	return ""
}

// stringClass returns the key of the class a string names in a callable
// written at call: fully qualified, or `self`, `static` or `parent`.
func (x *extractor) stringClass(s string, call syntax.Node) string {
	if c, ok := x.relative(s, call); ok {
		return c
	}
	return key(s)
}

// relative resolves `self`, `static` and `parent` written at call to the
// key of the class they name ("" outside a class, or for the parent of
// one extending none); ok is false for another name. Inside a trait,
// they name the trait.
func (x *extractor) relative(name string, call syntax.Node) (class string, ok bool) {
	switch strings.ToLower(name) {
	case "self", "static":
		return x.classFQN(call), true
	case "parent":
		c := syntax.EnclosingClass(call)
		if c == nil || c.ClassKind != syntax.KindClass || len(c.Extends) == 0 {
			return "", true
		}
		return key(x.r.Class(c.Extends[0].Value, c.Extends[0].Span().Start)), true
	}
	return "", false
}

// callback returns what the spl_autoload_register() call registers: a
// closure or arrow function with the name of its parameter receiving the
// class name, or the key of the function or method it names exactly (an
// AutoloadCallbacks entry): a function by its name, a method by its class
// (`$this`, `self::`, a class name) and its name.
func (x *extractor) callback(call *syntax.FuncCall) (body syntax.Node, param, k string) {
	var callable syntax.Expr
	for i, a := range call.Args.Args {
		arg, ok := a.(*syntax.Arg)
		if ok && (i == 0 && arg.Name == nil || arg.Name != nil && strings.EqualFold(arg.Name.Value, "callback")) {
			callable = arg.Value
		}
	}
	method := func(class, name string) string {
		if class == "" {
			return ""
		}
		return class + "::" + key(name)
	}
	switch v := callable.(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		if params := syntax.FuncLikeParams(v); len(params) > 0 && params[0].Var != nil {
			param = params[0].Var.Name
		}
		return v, param, ""
	case *syntax.Literal: // 'fn' or 'Cls::method'
		s, _ := syntax.LiteralStringValue(v)
		if class, name, ok := strings.Cut(s, "::"); ok {
			return nil, "", method(x.stringClass(class, call), name)
		}
		return nil, "", key(s)
	case *syntax.Binary: // __NAMESPACE__ . '\\fn'
		ns, ok := v.Left.(*syntax.MagicConst)
		lit, isLit := v.Right.(*syntax.Literal)
		if ok && isLit && v.Op.Kind == syntax.TDot && ns.Token.Kind == syntax.TNsC {
			name, _ := syntax.LiteralStringValue(lit)
			return nil, "", key(x.r.Namespace(v.Span().Start) + name)
		}
	case *syntax.Array: // [$this, 'method'], [Cls::class, 'method']
		if len(v.Items) == 2 && v.Items[0] != nil && v.Items[1] != nil {
			if lit, ok := v.Items[1].Value.(*syntax.Literal); ok {
				name, _ := syntax.LiteralStringValue(lit)
				return nil, "", method(x.classOf(v.Items[0].Value, call), name)
			}
		}
	case *syntax.FuncCall: // load(...): the namespaced function, else the global one
		if n, ok := v.Name.(*syntax.Name); ok && v.Args.FirstClassCallable() {
			fqn, fallback := x.r.Function(n.Value, n.Span().Start)
			if k = key(fqn); fallback != "" {
				k += "|" + key(fallback)
			}
			return nil, "", k
		}
	case *syntax.MethodCall: // $this->load(...)
		if id, ok := v.Name.(*syntax.Identifier); ok && !v.NullSafe && v.Args.FirstClassCallable() {
			if self, ok := v.Var.(*syntax.Variable); ok && self.Name == "this" && self.NameExpr == nil {
				return nil, "", method(x.classFQN(call), id.Value)
			}
		}
	case *syntax.StaticCall: // self::load(...), Cls::load(...)
		if id, ok := v.Name.(*syntax.Identifier); ok && v.Args.FirstClassCallable() {
			return nil, "", method(x.classOf(v.Class, call), id.Value)
		}
	}
	return nil, "", ""
}

// minAutoloadPrefix is the shortest prefix kept: shorter ones (`\`, `_`)
// would hide too many names.
const minAutoloadPrefix = 3

// prefixTested returns, lower case, the literal prefix call tests the class
// name in variable param for: the needle of strpos() or str_starts_with()
// on it, of at least minAutoloadPrefix characters once leading backslashes
// are dropped.
func prefixTested(call *syntax.FuncCall, param string) (string, bool) {
	name, ok := call.Name.(*syntax.Name)
	if param == "" || !ok || call.Args == nil || len(call.Args.Args) < 2 {
		return "", false
	}
	fn := strings.ToLower(name.Value[strings.LastIndexByte(name.Value, '\\')+1:])
	if fn != "strpos" && fn != "str_starts_with" {
		return "", false
	}
	var haystack, needle syntax.Expr // nil for a placeholder
	if arg, ok := call.Args.Args[0].(*syntax.Arg); ok {
		haystack = arg.Value
	}
	if arg, ok := call.Args.Args[1].(*syntax.Arg); ok {
		needle = arg.Value
	}
	if v, ok := haystack.(*syntax.Variable); !ok || v.Name != param {
		return "", false
	}
	// Decoding never lengthens a quoted string: a short one is rejected
	// before.
	lit, ok := needle.(*syntax.Literal)
	if !ok || len(lit.Raw) < minAutoloadPrefix+2 {
		return "", false
	}
	v, _ := syntax.LiteralStringValue(lit)
	v = strings.TrimLeft(v, `\`)
	if len(v) < minAutoloadPrefix {
		return "", false
	}
	return strings.ToLower(v), true
}

func (x *extractor) text(n syntax.Node) string {
	if n == nil {
		return ""
	}
	s := n.Span()
	return string(x.f.Src[s.Start:s.End])
}

func (x *extractor) resolver(at uint32) types.Resolver {
	return func(w string) string {
		if x.templates[w] {
			return ""
		}
		if def, ok := x.aliases[w]; ok {
			return aliasRef(w, def)
		}
		return x.r.Class(w, at)
	}
}

// aliasRef is what a resolver gives type alias w defined as def: the
// definition, or the mark of an imported one.
func aliasRef(w, def string) string {
	if def == importedAlias {
		return AliasMark + w
	}
	return "=" + def
}

// withTemplates adds the @template names of d for the duration of fn.
func (x *extractor) withTemplates(d *phpdoc.Doc, fn func()) {
	if d == nil {
		fn()
		return
	}
	names := d.Templates()
	if len(names) == 0 {
		fn()
		return
	}
	saved := x.templates
	x.templates = map[string]bool{}
	for k := range saved {
		x.templates[k] = true
	}
	for _, n := range names {
		x.templates[n] = true
	}
	fn()
	x.templates = saved
}

func (x *extractor) typeStr(n syntax.Expr, at uint32) string {
	t := types.FromNode(n, x.resolver(at))
	if t.IsUnknown() {
		return ""
	}
	return t.String()
}

func (x *extractor) docTypeStr(text string, at uint32) string {
	if t := x.docType(text, at); !t.IsUnknown() {
		return t.DocString() // keeps array shapes (round-trips through FromDoc)
	}
	return ""
}

// docTypeStrs is docTypeStr with the type's values (see Type.ValueDocString)
// when it documents integer ranges or non-empty strings, "" otherwise.
func (x *extractor) docTypeStrs(text string, at uint32) (doc, values string) {
	t := x.docType(text, at)
	if t.IsUnknown() {
		return "", ""
	}
	return t.DocString(), t.ValueDocString()
}

// docType parses PHPDoc type text written at position at, with the file's
// names, aliases and constants resolved.
func (x *extractor) docType(text string, at uint32) types.Type {
	resolveType := x.docTypeResolver(at)
	return types.FromDocResolved(text, x.resolver(at), func(expr string) types.Type {
		t := resolveType(expr)
		if !t.IsUnknown() {
			return t
		}
		for _, prefix := range []string{"key-of<", "value-of<"} {
			if strings.HasPrefix(expr, prefix) && strings.HasSuffix(expr, ">") {
				owner := expr[len(prefix) : len(expr)-1]
				if !strings.ContainsAny(owner, "<>{}[]:") {
					return types.Of(prefix + `\` + strings.TrimPrefix(x.resolver(at)(owner), `\`) + ">")
				}
			}
		}
		owner, member, ok := strings.Cut(expr, "::")
		if !ok {
			return types.Unknown
		}
		// Keep genuine, qualified PHPDoc constant syntax in index metadata.
		// It is parsed with the project index by inference, never exposed as a
		// runtime class type. Unsupported consumers conservatively read mixed.
		return types.Of(`\` + strings.TrimPrefix(x.resolver(at)(owner), `\`) + "::" + member)
	})
}

// DocComment returns the doc comment directly preceding node n ("" if none).
func DocComment(f *syntax.File, n syntax.Node) string {
	start := n.Span().Start
	toks := f.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= start })
	for j := i - 1; j >= 0; j-- {
		t := toks[j]
		switch t.Kind {
		case syntax.TWhitespace, syntax.TComment:
			continue
		case syntax.TDocComment:
			return string(f.Src[t.Start:t.End])
		}
		break
	}
	// Doc comment between attributes and the declaration.
	s := n.Span()
	for k := i; k < len(toks) && toks[k].Start < s.End; k++ {
		switch toks[k].Kind {
		case syntax.TDocComment:
			return string(f.Src[toks[k].Start:toks[k].End])
		case syntax.TAttribute, syntax.TWhitespace, syntax.TComment, syntax.TString, syntax.TNameQualified,
			syntax.TNameFullyQualified, syntax.TLParen, syntax.TRParen, syntax.TRBracket, syntax.TComma,
			syntax.TConstantEncapsedString, syntax.TLNumber:
			continue
		}
		break
	}
	return ""
}

func (x *extractor) doc(n syntax.Node) *phpdoc.Doc {
	if c := DocComment(x.f, n); c != "" {
		return phpdoc.Parse(c)
	}
	return nil
}

func (x *extractor) class(n *syntax.ClassLike) {
	d := x.doc(n)
	saved, savedTpl := x.aliases, x.classTpl
	x.classTpl = nil
	if d != nil {
		if a := d.TypeAliases(); len(a) > 0 {
			for _, it := range d.ImportedTypes() {
				a[it.Local] = importedAlias
			}
			x.aliases = a
		}
		for _, t := range d.Templates() {
			if x.classTpl == nil {
				x.classTpl = map[string]bool{}
			}
			x.classTpl[t] = true
		}
	}
	x.withTemplates(d, func() { x.classBody(n) })
	x.aliases, x.classTpl = saved, savedTpl
}

// docTypeResolver reads constant declarations without depending on extraction
// order, including enums declared after the documented declaration.
func (x *extractor) docTypeResolver(at uint32) types.DocTypeResolver {
	return func(expr string) types.Type {
		return ResolveDocType(expr, x.resolver(at), func(name string) *Class {
			if x.docClasses != nil {
				return x.docClasses[strings.ToLower(name)]
			}
			x.docClasses = map[string]*Class{}
			syntax.InspectFile(x.f, func(n syntax.Node) bool {
				c, ok := n.(*syntax.ClassLike)
				if !ok || c.Name == nil {
					return true
				}
				fqn := c.Name.Value
				if ns := x.r.Namespace(c.Span().Start); ns != "" {
					fqn = ns + `\` + fqn
				}
				result := &Class{FQN: fqn, Kind: c.ClassKind, Consts: map[string]*ClassConst{}}
				if c.EnumType != nil {
					result.Props = map[string]*Property{"value": {Type: x.typeStr(c.EnumType, c.Span().Start)}}
				}
				for _, member := range c.Members {
					switch m := member.(type) {
					case *syntax.EnumCase:
						result.Consts[m.Name.Value] = &ClassConst{Name: m.Name.Value, Case: true, Value: x.text(m.Value)}
					case *syntax.ClassConst:
						for _, k := range m.Consts {
							result.Consts[k.Name.Value] = &ClassConst{Name: k.Name.Value, Type: x.typeStr(m.Type, m.Span().Start)}
						}
					}
				}
				x.docClasses[strings.ToLower(fqn)] = result
				return true
			})
			return x.docClasses[strings.ToLower(name)]
		})
	}
}
