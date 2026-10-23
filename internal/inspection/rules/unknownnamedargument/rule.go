// Package unknownnamedargument implements the UnknownNamedArgument inspection.
package unknownnamedargument

import (
	"path"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "UnknownNamedArgument" }
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall}
}

const message = "Use a declared parameter name."

// verifiedStubs lists the stub directories of extensions bundled with PHP
// whose parameter names were compared with the engine's own reflection.
// Stub names of any other extension are documentation, not a contract: PHP
// binds named arguments against the names the extension itself declares.
var verifiedStubs = map[string]bool{
	"Core": true, "FFI": true, "PDO": true, "Phar": true, "Reflection": true, "SPL": true,
	"SimpleXML": true, "Zend OPcache": true, "bcmath": true, "bz2": true, "calendar": true,
	"ctype": true, "curl": true, "date": true, "dba": true, "dom": true, "exif": true,
	"fileinfo": true, "filter": true, "ftp": true, "gd": true, "gettext": true, "gmp": true,
	"hash": true, "iconv": true, "intl": true, "json": true, "ldap": true, "libsodium": true,
	"libxml": true, "mbstring": true, "mysqli": true, "odbc": true, "openssl": true,
	"pcntl": true, "pcre": true, "pgsql": true, "posix": true, "random": true,
	"readline": true, "session": true, "shmop": true, "soap": true, "sockets": true,
	"sodium": true, "sqlite3": true, "standard": true, "sysvmsg": true, "sysvsem": true,
	"sysvshm": true, "tidy": true, "tokenizer": true, "uri": true, "xml": true,
	"xmlreader": true, "xmlwriter": true, "xsl": true, "zip": true, "zlib": true,
}

// driftedStubs lists the builtins of verified directories whose stub
// parameter list differs from the one PHP declares (overloaded signatures
// reduced to one form, renamed parameters).
var driftedStubs = map[string]bool{
	"session_set_cookie_params": true, "dba_fetch": true, "ldap_sasl_bind": true,
	"intlcalendar::set": true,
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	var args *syntax.ArgList
	switch call := n.(type) {
	case *syntax.FuncCall:
		args = call.Args
	case *syntax.MethodCall:
		args = call.Args
	case *syntax.StaticCall:
		args = call.Args
	}
	if args == nil {
		return
	}
	// The nominal declaration settles almost every call; the overrides of
	// the receiver's subtypes are collected only for a name it lacks.
	for _, overrides := range []bool{false, true} {
		names := map[string]bool{}
		if !targets(ctx, n, overrides, names) {
			return
		}
		var unknown []*syntax.Arg
		for _, node := range args.Args {
			if arg, ok := node.(*syntax.Arg); ok && arg.Name != nil && !names[arg.Name.Value] {
				unknown = append(unknown, arg)
			}
		}
		if !overrides && len(unknown) > 0 {
			continue
		}
		for _, arg := range unknown {
			ctx.ReportNode(arg, message)
		}
		return
	}
}

// targets adds the parameter names of every declaration the call can bind
// to; false when a target is unknown, collects named arguments or cannot be
// trusted. With overrides, the methods a subtype of the receiver declares in
// place of the nominal one count as targets too.
func targets(ctx *analysis.Context, call syntax.Node, overrides bool, names map[string]bool) bool {
	switch c := call.(type) {
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(c)
		if f == nil || f.Builtin && !trustedStub(f.File, f.FQN) || !addParams(names, f.Params) {
			return false
		}
		// A project function declared several times (loaded conditionally)
		// accepts the names of whichever declaration runs.
		for _, decl := range ctx.Index().FunctionDecls(f.FQN, ctx.PHP) {
			if !addParams(names, decl.Params) {
				return false
			}
		}
		return true
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		classes := ctx.TypeOf(c.Var).Classes()
		if !ok || len(classes) == 0 {
			return false
		}
		for _, cls := range classes {
			if !methodTargets(ctx, strings.TrimPrefix(cls, `\`), id.Value, overrides, names) {
				return false
			}
		}
		return true
	}
	c := call.(*syntax.StaticCall)
	id, ok := c.Name.(*syntax.Identifier)
	cls := semanticquery.StaticCallClass(ctx, c.Class)
	if !ok || cls == "" {
		return false
	}
	// Only `static::` binds late: a named class, self and parent call the
	// method that class resolves, whatever a subclass declares.
	late := strings.EqualFold(c.Class.(*syntax.Name).Value, "static")
	return methodTargets(ctx, cls, id.Value, overrides && late, names)
}

// methodTargets adds the parameters of method name as class cls resolves it
// and, with overrides, as each indexed subtype of cls resolves it.
func methodTargets(ctx *analysis.Context, cls, name string, overrides bool, names map[string]bool) bool {
	ix := ctx.Index()
	c := ix.Class(cls, ctx.PHP)
	// A trait method runs in the classes using the trait, which are not
	// enumerable from here and may declare the callee themselves.
	if c == nil || c.Kind == syntax.KindTrait {
		return false
	}
	m := ix.FindMethod(cls, name, ctx.PHP)
	if redeclared(ctx, cls) || !addMethod(ctx, m, names) || redeclared(ctx, m.Class) {
		return false
	}
	if !overrides || m.Visibility == index.Private || m.Final || c.Final {
		return true
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
			if redeclared(ctx, child) || !addMethod(ctx, ix.FindMethod(child, name, ctx.PHP), names) {
				return false
			}
			queue = append(queue, child)
		}
	}
	return true
}

// redeclared reports whether the project declares class cls more than once
// (E3): which declaration runs depends on what the autoloader loads.
func redeclared(ctx *analysis.Context, cls string) bool {
	return len(ctx.Index().ClassDecls(cls, ctx.PHP)) > 1
}

// addMethod adds the parameters of a resolved method; false for an unknown
// method, a `@method` tag (served by __call, which accepts any name), an
// untrusted builtin or a variadic collector.
func addMethod(ctx *analysis.Context, m *index.Method, names map[string]bool) bool {
	if m == nil || m.Magic {
		return false
	}
	if m.Builtin {
		file := ""
		for _, decl := range ctx.Index().ClassDecls(m.Class, 0) {
			file = decl.File
		}
		if !trustedStub(file, m.Class+"::"+m.Name) {
			return false
		}
	}
	return addParams(names, m.Params)
}

// addParams adds the parameter names; false when a variadic parameter
// collects every unknown named argument.
func addParams(names map[string]bool, params []index.Param) bool {
	for _, param := range params {
		if param.Variadic {
			return false
		}
		names[param.Name] = true
	}
	return true
}

// trustedStub reports whether the stub declaration of callee (a function
// FQN or `Class::method`) in file carries the parameter names PHP declares.
func trustedStub(file, callee string) bool {
	callee = strings.ToLower(strings.TrimPrefix(callee, `\`))
	return verifiedStubs[path.Base(path.Dir(file))] && !driftedStubs[callee]
}

func (rule) Semantic() {}
