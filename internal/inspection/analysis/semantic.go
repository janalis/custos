package analysis

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
	"custos/internal/semantic/types"
)

// SemanticRule marks rules that need the project symbol index (classes and
// functions declared in other files). When any enabled rule implements it,
// the runner builds the project index before analysis.
type SemanticRule interface {
	Semantic()
}

// NeedsIndex reports whether an enabled rule requires the project index.
func (e *Engine) NeedsIndex() bool {
	for _, r := range e.rules {
		if _, ok := r.rule.(SemanticRule); ok {
			return true
		}
	}
	return false
}

// WithIndex returns an engine copy configured with ix; nil selects builtins only.
// The original engine remains safe for concurrent analyses.
func (e *Engine) WithIndex(ix *index.Index) *Engine {
	configured := *e
	configured.index = ix
	return &configured
}

// projectIndex returns the configured project index or the stubs.
func (e *Engine) projectIndex() *index.Index {
	if e.index != nil {
		return e.index
	}
	return stubs.Index()
}

// Names returns the file's name resolver (built on first use).
func (c *Context) Names() *names.Resolver {
	if c.names == nil {
		c.names = names.New(c.File)
	}
	return c.names
}

// Index returns a symbol index containing this file's own declarations
// layered over the project index and the PHP stubs (built on first use).
func (c *Context) Index() *index.Index {
	if c.index == nil {
		c.index = index.New(c.engine.projectIndex())
		c.index.Add(index.Extract(c.File))
	}
	return c.index
}

// Types returns the type inference environment for this file.
func (c *Context) Types() *infer.Env {
	if c.types == nil {
		c.types = infer.NewEnv(c.File, c.Names(), c.Index(), c.PHP)
	}
	return c.types
}

// TypeOf is shorthand for c.Types().TypeOf(x).
func (c *Context) TypeOf(x syntax.Expr) types.Type { return c.Types().TypeOf(x) }

// IsGlobalFunctionCall reports whether call is a call to the global function
// `global` (case-insensitive), honouring namespaces, `use function` and
// same-named functions declared in the current namespace (this file or the
// project index), which PHP prefers over the global fallback.
func (c *Context) IsGlobalFunctionCall(call *syntax.FuncCall, global string) bool {
	name := c.GlobalFunctionName(call)
	return name != "" && strings.EqualFold(name, global)
}

// GlobalFunctionName returns the lower-case name of the global function a
// named call targets, or "" when it is dynamic or targets a namespaced
// function: a qualified non-global name, a non-global `use function`
// import, or an unqualified call in a namespace where a function of that
// name is declared (it wins over PHP's global fallback). Rules matching
// builtins switch on it instead of on the written name.
func (c *Context) GlobalFunctionName(call *syntax.FuncCall) string {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return ""
	}
	fqn, fallback := c.Names().Function(name.Value, name.Span().Start)
	if fallback != "" {
		if c.Index().Function(fqn, c.PHP) != nil {
			return ""
		}
		return lower(fallback)
	}
	if strings.Contains(fqn, `\`) {
		return ""
	}
	return lower(fqn)
}

// FunctionName returns the resolved global-fallback name of a call to a
// named function (lower-case FQN without leading backslash), or "" for
// dynamic calls. In a namespace, an unqualified call resolves to the global
// function name (PHP's runtime fallback) unless imported.
func (c *Context) FunctionName(call *syntax.FuncCall) string {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return ""
	}
	fqn, fallback := c.Names().Function(name.Value, name.Span().Start)
	if fallback != "" {
		return lower(fallback)
	}
	return lower(fqn)
}

func lower(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			b := []byte(s)
			for j := i; j < len(b); j++ {
				if b[j] >= 'A' && b[j] <= 'Z' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return s
}

// NewTestContext returns a context for running a rule directly (tests and
// debugging only; bypasses suppression and panic isolation).
func NewTestContext(e *Engine, f *syntax.File) *Context {
	c := &Context{File: f, Src: f.Src, PHP: e.cfg.PHP, ComparisonStyle: e.cfg.ComparisonStyle, engine: e}
	if len(e.rules) > 0 {
		c.cur = &e.rules[0]
	}
	return c
}

// Memo returns the value cached under key for the current rule and file,
// computing it with fn on first use. Rules use it for per-file derived data
// (parsed options, lookup tables) instead of recomputing it per node.
func (c *Context) Memo(key string, fn func() any) any {
	k := c.cur.meta.ID + "\x00" + key
	if v, ok := c.memo[k]; ok {
		return v
	}
	if c.memo == nil {
		c.memo = map[string]any{}
	}
	v := fn()
	c.memo[k] = v
	return v
}
