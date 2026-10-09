package infer

import (
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/types"
)

// Env infers expression types within one file. Not safe for concurrent use.
type Env struct {
	File  *syntax.File
	Names *names.Resolver
	Index *index.Index // project index layered over the stubs
	PHP   phpversion.Version

	chains map[syntax.Expr]nullsafeFact
	cache  map[syntax.Expr]types.Type
	scopes map[syntax.Node]*scopeVars
	busy   map[syntax.Expr]bool
	// bases holds the type of variable reads before the element writes
	// reaching them are applied (see baseType).
	bases map[*syntax.Variable]types.Type

	// Return types inferred from bodies (see bodyReturn), by declaration span.
	bodies    map[syntax.Span]types.Type
	bodyBusy  map[syntax.Span]bool
	bodyDepth int
	decls     map[syntax.Span]syntax.Node

	// Index-time return inference (see AnnotateReturns): the index holds
	// only this file over the stubs, and deps records the namespaced
	// function names resolved through the global fallback.
	annotating bool
	deps       []string

	// generics caches genBindings.
	generics map[string]map[string]tplBindings

	// docs caches DocOf; docScopes caches the template names and type
	// aliases in scope at a declaration (see resolverFor).
	docs      map[syntax.Node]*phpdoc.Doc
	docScopes map[syntax.Node]*docScope

	// asserts caches assertsOf.
	asserts map[syntax.Expr]*callAsserts

	// Untyped property inference (see propinfer.go): writes per class,
	// results and recursion guard by property span, classes by span.
	propWritesCache map[*syntax.ClassLike]*classWrites
	props           map[syntax.Span]types.Type
	propBusy        map[syntax.Span]bool
	classes         map[syntax.Span]syntax.Node

	// guardIdx caches guardIndexOf by statement-list owner (nil: file).
	guardIdx map[syntax.Node]*guardIndex

	// conds caches parsedCond; builtinCalls caches isBuiltinCall.
	conds        map[string]*types.Cond
	builtinCalls map[*syntax.FuncCall]bool

	// dynReads caches dynamicRead per variable read (computed by
	// variableBase); assignsMemo caches stmtAssigns.
	dynReads          map[*syntax.Variable]uint8
	assignsMemo       map[assignKey]bool
	captures          map[captureKey]types.Type
	captureBusy       map[captureKey]bool
	captureDepth      int
	conditionalChains map[syntax.Expr]bool

	// native: types come from native declarations only (see Native);
	// nativeTwin caches the native Env of this one.
	native     bool
	nativeTwin *Env
}

// Native returns an Env over the same file and index whose types ignore
// user PHPDoc: parameter, property and return docs, inline @var, templates,
// conditional returns, assertions and @param-out of project declarations
// (builtin stub docs still count, they describe PHP), and the index-time
// inferred types (computed with docs). A rule can thus tell whether a
// type rests on PHPDoc only (UnnecessaryCasting). Cached per Env.
func (e *Env) Native() *Env {
	if e.native {
		return e
	}
	if e.nativeTwin == nil {
		e.nativeTwin = NewEnv(e.File, e.Names, e.Index, e.PHP)
		e.nativeTwin.native = true
	}
	return e.nativeTwin
}

// IsNative reports whether e is a native Env (see Native).
func (e *Env) IsNative() bool { return e.native }

// userDoc reports whether documented facts of a declaration (builtin: from
// the stubs) are ignored by this Env (see Native).
func (e *Env) userDoc(builtin bool) bool { return e.native && !builtin }

// docScope holds the @template names and type aliases declared on a
// declaration and its enclosing ones.
type docScope struct {
	tpl     map[string]bool
	aliases map[string]string
}

// DocOf returns the parsed doc comment of declaration n (nil when it has
// none), parsed once per Env.
func (e *Env) DocOf(n syntax.Node) *phpdoc.Doc {
	if d, ok := e.docs[n]; ok {
		return d
	}
	var d *phpdoc.Doc
	if c := index.DocComment(e.File, n); c != "" {
		d = phpdoc.Parse(c)
	}
	if e.docs == nil {
		e.docs = map[syntax.Node]*phpdoc.Doc{}
	}
	e.docs[n] = d
	return d
}

// scopeDocs returns the template names and aliases declared on the
// function-likes and classes enclosing n (n included), innermost last.
func (e *Env) scopeDocs(n syntax.Node) *docScope {
	for n != nil {
		switch n.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.PropertyHook, *syntax.ClassLike:
		default:
			n = n.Parent()
			continue
		}
		break
	}
	if n == nil {
		return nil
	}
	if s, ok := e.docScopes[n]; ok {
		return s
	}
	outer := e.scopeDocs(n.Parent())
	s := outer
	if d := e.DocOf(n); d != nil {
		names, aliases := d.Templates(), d.TypeAliases()
		if len(names) > 0 || len(aliases) > 0 {
			s = &docScope{tpl: map[string]bool{}, aliases: map[string]string{}}
			if outer != nil {
				for k := range outer.tpl {
					s.tpl[k] = true
				}
				for k, v := range outer.aliases {
					s.aliases[k] = v
				}
			}
			for _, t := range names {
				s.tpl[t] = true
			}
			for k, v := range aliases {
				s.aliases[k] = v
			}
		}
	}
	if e.docScopes == nil {
		e.docScopes = map[syntax.Node]*docScope{}
	}
	e.docScopes[n] = s
	return s
}

// NewEnv creates an inference environment.
func NewEnv(f *syntax.File, r *names.Resolver, ix *index.Index, php phpversion.Version) *Env {
	return &Env{
		File: f, Names: r, Index: ix, PHP: php, cache: map[syntax.Expr]types.Type{}, scopes: map[syntax.Node]*scopeVars{}, busy: map[syntax.Expr]bool{},
		bodies: map[syntax.Span]types.Type{}, bodyBusy: map[syntax.Span]bool{},
	}
}

func (e *Env) resolver(at uint32) types.Resolver {
	return func(w string) string { return e.className(w, at) }
}

// className resolves a class name written in a doc type at offset at. For
// a "!name" probe (see names.Resolver.Class) a class declared under that
// name in the current namespace also counts.
func (e *Env) className(w string, at uint32) string {
	if probe, ok := strings.CutPrefix(w, "!"); ok {
		if fqn := e.Names.Class(w, at); fqn != "" {
			return fqn
		}
		if fqn := e.Names.Class(probe, at); e.Index.Class(fqn, e.PHP) != nil {
			return fqn
		}
		return ""
	}
	return e.Names.Class(w, at)
}

// resolverFor resolves names in docs attached around n, mapping @template
// names declared on the enclosing function/method/class to "" (mixed).
func (e *Env) resolverFor(n syntax.Node, at uint32) types.Resolver {
	ds := e.scopeDocs(n)
	if ds == nil {
		return e.resolver(at)
	}
	return func(w string) string {
		if ds.tpl[w] {
			return ""
		}
		if def, ok := ds.aliases[w]; ok {
			return "=" + def
		}
		return e.className(w, at)
	}
}
