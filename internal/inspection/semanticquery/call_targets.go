package semanticquery

import (
	"path"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// CallTarget is one declaration a call may bind its arguments to.
type CallTarget struct {
	Params []index.Param
	// Builtin marks a declaration of PHP itself (the stubs).
	Builtin bool
}

// TrustBuiltin reports whether the stub declaration of builtin callee (a
// function FQN or `Class::method`) declared in stub file states what the
// consumer checks; CallTargets gives up on a call reaching any other.
type TrustBuiltin func(file, callee string) bool

// bundledStubs lists the stub directories of the extensions bundled with
// PHP. The stubs of other extensions document their API, not what PHP
// enforces.
var bundledStubs = map[string]bool{
	"Core": true, "FFI": true, "PDO": true, "Phar": true, "Reflection": true, "SPL": true,
	"SimpleXML": true, "bcmath": true, "bz2": true, "calendar": true, "ctype": true, "curl": true,
	"date": true, "dba": true, "dom": true, "exif": true, "fileinfo": true, "filter": true,
	"ftp": true, "gd": true, "gettext": true, "gmp": true, "hash": true, "iconv": true, "intl": true,
	"json": true, "ldap": true, "libxml": true, "mbstring": true, "mysqli": true, "odbc": true,
	"openssl": true, "pcntl": true, "pcre": true, "pgsql": true, "posix": true, "random": true,
	"readline": true, "session": true, "shmop": true, "soap": true, "sockets": true, "sodium": true,
	"sqlite3": true, "standard": true, "sysvmsg": true, "sysvsem": true, "sysvshm": true,
	"tidy": true, "tokenizer": true, "uri": true, "xml": true, "xmlreader": true, "xmlwriter": true,
	"xsl": true, "zip": true, "zlib": true,
}

// BundledStub reports whether stub file declares a builtin of an extension
// bundled with PHP.
func BundledStub(file string) bool { return bundledStubs[path.Base(path.Dir(file))] }

// CallTargets lists every declaration a call or construction may bind to
// and names the nominal one (`f`, `Class::method`); false when any of them
// is unknown or uncertain. callee is the call's agreed contract (see
// ResolveCallee).
//
//   - A function call binds to its single declaration: not a polyfill of a
//     function PHP declares in some version, not declared more than once.
//   - A construction binds to the constructor of a concrete class whose
//     hierarchy is resolved.
//   - An instance method call binds to the method as each class of the
//     receiver resolves it and, unless the method is private or final or
//     its class final, as every indexed subtype of those classes resolves
//     it; a `static::` call likewise from the enclosing class; another
//     static call to the method its class resolves.
//
// Every class involved must be declared once with a resolved hierarchy; a
// method served by a `@method` tag, reached through `__call` or
// `__callStatic`, or declared by a trait the receiver names is uncertain.
func CallTargets(ctx *analysis.Context, call syntax.Node, callee Callee, trusted TrustBuiltin) (string, []CallTarget, bool) {
	r := targetResolver{ctx: ctx, trusted: trusted}
	switch c := call.(type) {
	case *syntax.FuncCall:
		return r.function(c)
	case *syntax.MethodCall:
		id := c.Name.(*syntax.Identifier)
		name := ""
		var all []CallTarget
		for _, cls := range ctx.TypeOf(c.Var).Without("null").Classes() {
			n, ts, ok := r.methodTargets(strings.TrimPrefix(cls, `\`), id.Value, true)
			if !ok {
				return "", nil, false
			}
			if name == "" {
				name = n
			}
			all = append(all, ts...)
		}
		return name, all, true
	case *syntax.StaticCall:
		id := c.Name.(*syntax.Identifier)
		cls := StaticCallClass(ctx, c.Class)
		// Only `static::` binds late; a named class, self and parent call
		// the method that class resolves.
		late := strings.EqualFold(c.Class.(*syntax.Name).Value, "static")
		return r.methodTargets(cls, id.Value, late)
	}
	return r.construction(call.(*syntax.New), callee)
}

type targetResolver struct {
	ctx     *analysis.Context
	trusted TrustBuiltin
}

// function resolves a function call to its single declaration.
func (r targetResolver) function(c *syntax.FuncCall) (string, []CallTarget, bool) {
	ix := r.ctx.Index()
	f := r.ctx.Types().ResolveFunction(c)
	if f.Builtin && !r.trusted(f.File, f.FQN) {
		return "", nil, false
	}
	if !f.Builtin {
		// A polyfill of a function PHP declares in another version, or a
		// function the project declares more than once (conditionally
		// loaded variants), may not be the declaration that runs.
		if other := ix.Function(f.FQN, 0); other != nil && other.Builtin {
			return "", nil, false
		}
		if len(ix.FunctionDecls(f.FQN, 0)) != 1 {
			return "", nil, false
		}
	}
	return strings.TrimPrefix(f.FQN, `\`), []CallTarget{{Params: f.Params, Builtin: f.Builtin}}, true
}

// construction resolves the constructor a `new` runs.
func (r targetResolver) construction(c *syntax.New, callee Callee) (string, []CallTarget, bool) {
	ix := r.ctx.Index()
	cls := r.ctx.Types().ClassRef(c.Class.(*syntax.Name))
	declared := ix.Class(cls, r.ctx.PHP)
	if declared == nil || declared.Kind != syntax.KindClass || declared.Abstract || !HierarchyResolved(ix, cls, r.ctx.PHP) {
		return "", nil, false
	}
	if callee.Builtin {
		if m := ix.FindMethod(cls, "__construct", r.ctx.PHP); m == nil || !r.trustedMethod(m) {
			return "", nil, false
		}
	}
	return strings.TrimPrefix(declared.FQN, `\`) + "::__construct", []CallTarget{{Params: callee.Params, Builtin: callee.Builtin}}, true
}

// methodTargets lists method name as class cls resolves it and, with
// overrides, as every indexed subtype of cls resolves it, since the
// runtime object may be any of them.
func (r targetResolver) methodTargets(cls, name string, overrides bool) (string, []CallTarget, bool) {
	ix := r.ctx.Index()
	c := ix.Class(cls, r.ctx.PHP)
	// A trait method runs in the classes using the trait, which may declare
	// the callee themselves.
	if c == nil || c.Kind == syntax.KindTrait {
		return "", nil, false
	}
	m, ok := r.method(cls, name)
	if !ok {
		return "", nil, false
	}
	label := strings.TrimPrefix(m.Class, `\`) + "::" + m.Name
	targets := []CallTarget{{Params: m.Params, Builtin: m.Builtin}}
	if !overrides || m.Visibility == index.Private || m.Final || c.Final {
		return label, targets, true
	}
	seen := map[string]bool{}
	queue := []string{cls}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, child := range ix.ChildrenAll(cur) {
			k := strings.ToLower(strings.TrimPrefix(child, `\`))
			if seen[k] {
				continue
			}
			seen[k] = true
			sub, ok := r.method(child, name)
			if !ok {
				return "", nil, false
			}
			targets = append(targets, CallTarget{Params: sub.Params, Builtin: sub.Builtin})
			queue = append(queue, child)
		}
	}
	return label, targets, true
}

// method resolves name on class cls when that declaration is certain: the
// class and the one declaring the method are declared once, every ancestor
// is indexed, and neither a `@method` tag nor a magic handler serves it.
func (r targetResolver) method(cls, name string) (*index.Method, bool) {
	ix := r.ctx.Index()
	ver := r.ctx.PHP
	if len(ix.ClassDecls(cls, ver)) != 1 || !HierarchyResolved(ix, cls, ver) {
		return nil, false
	}
	m := ix.FindMethod(cls, name, ver)
	if m == nil || m.Magic || len(ix.ClassDecls(m.Class, ver)) != 1 || m.Builtin && !r.trustedMethod(m) {
		return nil, false
	}
	// A method the caller cannot see is handed to __call or __callStatic.
	if m.Visibility != index.Public && (ix.FindMethod(cls, "__call", ver) != nil || ix.FindMethod(cls, "__callStatic", ver) != nil) {
		return nil, false
	}
	return m, true
}

// trustedMethod reports whether builtin method m is trusted.
func (r targetResolver) trustedMethod(m *index.Method) bool {
	decls := r.ctx.Index().ClassDecls(m.Class, 0)
	return len(decls) == 1 && r.trusted(decls[0].File, m.Class+"::"+m.Name)
}
