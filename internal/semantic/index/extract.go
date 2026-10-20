package index

import (
	"slices"
	"sort"
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	"custos/internal/semantic/names"
	"custos/internal/semantic/types"
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
	aliases    map[string]string // @phpstan-type / import-type names in scope
}

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
			if x.declarationBuiltin(n, nil) != "" {
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
	if len(calls) > 0 {
		declared := make(map[string]bool, len(x.out.Functions))
		for _, fn := range x.out.Functions {
			declared[key(fn.FQN)] = true
		}
		for _, call := range calls {
			switch x.declarationBuiltin(call, declared) {
			case "define":
				x.define(call)
			case "class_alias":
				x.classAlias(call)
			}
		}
		// Deferred define calls must retain source order relative to const
		// declarations: duplicate declarations use the first indexed symbol.
		sort.SliceStable(x.out.Constants, func(i, j int) bool {
			return x.out.Constants[i].Span.Start < x.out.Constants[j].Span.Start
		})
	}
	return x.out
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
			return "=" + def
		}
		return x.r.Class(w, at)
	}
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
	resolveType := x.docTypeResolver(at)
	t := types.FromDocResolved(text, x.resolver(at), func(expr string) types.Type {
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
	if t.IsUnknown() {
		return ""
	}
	return t.DocString() // keeps array shapes (round-trips through FromDoc)
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
