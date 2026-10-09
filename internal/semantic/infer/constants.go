package infer

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

// resolvedConstType respects constant imports and namespace fallback before
// consulting the known types of global runtime constants.
func (e *Env) resolvedConstType(n *syntax.ConstFetch) types.Type {
	switch v := strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)); v {
	case "true", "false":
		return types.Of(v)
	case "null":
		return types.Null
	}
	fqn, fallback := e.Names.Const(n.Name.Value, n.Span().Start)
	if t, found := e.namedConstType(fqn); found {
		return t
	}
	if fallback != "" {
		if t, found := e.namedConstType(fallback); found {
			return t
		}
	}
	return types.Unknown
}

func (e *Env) namedConstType(fqn string) (types.Type, bool) {
	c := e.Index.Constant(fqn, e.PHP)
	// Embedded constants sometimes lack introduction metadata.
	// Source declarations can provide polyfills before a builtin is introduced.
	if (c == nil || c.Builtin) && e.PHP != 0 && e.PHP < builtinConstSince(fqn) {
		return types.Unknown, false
	}
	if c != nil {
		// Index lookups may return an unavailable declaration as a best-effort
		// candidate. Constant evaluation needs strict runtime availability.
		if e.PHP != 0 && !c.Avail.In(e.PHP) {
			return types.Unknown, false
		}
	}
	// Runtime builtins retain their value when a source declaration attempts
	// to redefine them. Only names resolved to those globals reach this case.
	if t := builtinConstType(fqn, e.PHP); !t.IsUnknown() {
		return t, true
	}
	if c != nil {
		return literalTextType(c.Value), true
	}
	return types.Unknown, false
}

func builtinConstType(name string, ver phpversion.Version) types.Type {
	if ver != 0 && ver < builtinConstSince(name) {
		return types.Unknown
	}
	switch name {
	case "PHP_INT_MIN":
		return types.Int
	case "PHP_OS_FAMILY", "PHP_FLOAT_EPSILON", "PHP_FLOAT_MAX", "PHP_FLOAT_MIN":
		if name == "PHP_OS_FAMILY" {
			return types.String
		}
		return types.Float
	case "PHP_INT_MAX", "PHP_INT_SIZE", "PHP_VERSION_ID", "PHP_MAJOR_VERSION", "PHP_MINOR_VERSION", "E_ALL", "E_ERROR", "E_WARNING", "E_NOTICE", "E_STRICT", "E_DEPRECATED":
		return types.Int
	case "PHP_EOL", "PHP_VERSION", "PHP_OS", "DIRECTORY_SEPARATOR", "PATH_SEPARATOR":
		return types.String
	case "M_PI", "NAN", "INF":
		return types.Float
	}
	return types.Unknown
}

// builtinConstSince supplements missing version metadata in embedded stubs.
func builtinConstSince(name string) phpversion.Version {
	switch name {
	case "PHP_INT_MIN":
		return phpversion.PHP70
	case "PHP_OS_FAMILY", "PHP_FLOAT_EPSILON", "PHP_FLOAT_MAX", "PHP_FLOAT_MIN":
		return phpversion.PHP72
	}
	return 0
}
