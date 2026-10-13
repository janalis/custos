package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// This file is the one place that says which code of PHP itself runs code of
// the project: what a function named as a callback may do (pure), and what a
// function, a method or a construction written as a call does (runner,
// loads, invokes, builds). The flow asks it where a guard may be undone
// (handed), the summaries where a request table may be filled (spares).

// pure reports a PHP function that only computes on the text and the numbers
// it is given: it loads no class, calls no handler and reads no variable of
// the scope that calls it. The list is closed, and names only functions that
// are part of every PHP build, which a project cannot declare in their
// place.
func pure(name string) bool {
	switch name {
	case "strlen", "trim", "ltrim", "rtrim", "chop", "strtoupper", "strtolower", "ucfirst", "lcfirst", "ucwords", "strrev", "nl2br",
		"strcmp", "strcasecmp", "strncmp", "strncasecmp", "strnatcmp", "strnatcasecmp",
		"intval", "floatval", "doubleval", "boolval", "strval", "abs", "ceil", "floor", "round", "sqrt",
		"is_string", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_bool", "is_null", "is_numeric", "is_scalar", "is_array",
		"addslashes", "stripslashes", "htmlspecialchars", "htmlspecialchars_decode", "htmlentities", "html_entity_decode", "strip_tags",
		"urlencode", "urldecode", "rawurlencode", "rawurldecode", "base64_encode", "base64_decode", "bin2hex", "hex2bin", "md5", "sha1", "crc32",
		"quotemeta", "preg_quote", "escapeshellarg", "escapeshellcmd", "ord", "chr", "dechex", "hexdec", "decbin", "bindec", "decoct", "octdec":
		return true
	}
	return false
}

// runner reports a PHP function that runs code of the project whatever it is
// given: an error handler, an autoloader, the methods that restore an
// object, the handler of a session, which also holds the objects stored in
// it, the destructors of the objects a collection frees, the handler of an
// output buffer that is flushed or dropped, the callbacks set on a transfer
// or a parser, the callbacks an array names.
func runner(name string) bool {
	switch name {
	case "trigger_error", "user_error", "class_exists", "interface_exists", "trait_exists", "enum_exists", "spl_autoload_call", "spl_autoload", "unserialize",
		"session_start", "session_write_close", "session_commit", "session_destroy", "session_regenerate_id", "session_reset", "session_gc", "session_decode",
		"session_abort", "session_create_id", "session_encode",
		"gc_collect_cycles",
		"ob_flush", "ob_clean", "ob_end_flush", "ob_end_clean", "ob_get_flush", "ob_get_clean",
		"curl_exec", "curl_multi_exec", "xml_parse", "xml_parse_into_struct",
		"preg_replace_callback_array":
		return true
	}
	return false
}

// loads reports a call of a PHP function that runs code of the project by
// what it is given, other than a function to call: a class named by text,
// which is loaded (not one named by an object, `self::class` or
// `static::class`, which is loaded already), a constant written with `::`, a
// class to construct, a filter that calls back. Arguments given by name or
// unpacked may be any of these.
func loads(name string, list *syntax.ArgList) bool {
	args, ordered := positional(list)
	switch name {
	case "is_subclass_of", "method_exists", "property_exists", "get_parent_class", "class_implements", "class_parents", "class_uses",
		"get_class_methods", "get_class_vars":
		return !ordered || len(args) > 0 && !instance(value(args[0]))
	case "is_a":
		// Text is only read as a class name when that is asked for.
		return !ordered || len(args) > 2 && constantName(value(args[2])) != "false" && !instance(value(args[0]))
	case "is_callable":
		return !ordered || len(args) > 0 && !unbound(value(args[0]))
	case "constant", "defined":
		return !ordered || len(args) > 0 && !unbound(value(args[0]))
	case "simplexml_load_string", "simplexml_load_file":
		return !ordered || len(args) > 1 && constantName(value(args[1])) != "null"
	case "mysqli_fetch_object":
		return !ordered || len(args) > 1
	case "pg_fetch_object":
		return !ordered || len(args) > 2
	case "filter_var", "filter_var_array":
		return !ordered || filters(args, 1)
	case "filter_input", "filter_input_array":
		return !ordered || filters(args, 2)
	}
	return false
}

// invokes reports a method PHP itself declares, named class::method in
// lower case, that runs code of the project: it constructs a class or calls
// a function it is given the name of, serves a request with the functions
// registered, or moves an iterator that wraps another one.
func invokes(qualified string, list *syntax.ArgList) bool {
	args, ordered := positional(list)
	switch qualified = strings.TrimPrefix(qualified, `\`); qualified {
	case "pdostatement::fetchobject", "mysqli_result::fetch_object":
		return !ordered || len(args) > 0
	case "pdostatement::fetchall":
		return !ordered || len(args) > 1
	case "soapserver::handle":
		return true
	}
	class, _, _ := strings.Cut(qualified, "::")
	switch class {
	case "iteratoriterator", "filteriterator", "callbackfilteriterator", "recursivefilteriterator", "recursivecallbackfilteriterator", "parentiterator",
		"limititerator", "cachingiterator", "recursivecachingiterator", "appenditerator", "infiniteiterator", "norewinditerator",
		"regexiterator", "recursiveregexiterator", "multipleiterator", "recursiveiteratoriterator", "recursivetreeiterator":
		return true
	}
	return false
}

// builds reports the construction of a class of PHP, named in lower case,
// that loads a class named by text: a reflection of it, or an `ArrayObject`
// given the class of its iterators.
func builds(class string, list *syntax.ArgList) bool {
	args, ordered := positional(list)
	switch class = strings.TrimPrefix(class, `\`); {
	case strings.HasPrefix(class, "reflection"):
		return !ordered || len(args) > 0 && !instance(value(args[0]))
	case class == "arrayobject":
		return !ordered || len(args) > 2
	}
	return false
}

// positional returns the arguments of a call written in order; it reports
// false when one is given by name or unpacked, and nothing says which
// parameter takes what.
func positional(list *syntax.ArgList) ([]syntax.Expr, bool) {
	if list == nil {
		return nil, true
	}
	for _, x := range list.Args {
		if arg, ok := x.(*syntax.Arg); !ok || arg.Unpack || arg.Name != nil {
			return nil, false
		}
	}
	return list.Args, true
}

// value returns what an argument written in order gives.
func value(arg syntax.Expr) syntax.Expr {
	return arg.(*syntax.Arg).Value
}

// instance reports an argument that names a class already loaded: an object
// written as `$this`, `new` or `clone`, or the class of the code that runs.
func instance(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		return n.NameExpr == nil && n.Name == "this"
	case *syntax.New, *syntax.Clone:
		return true
	case *syntax.ClassConstFetch:
		class, named := n.Class.(*syntax.Name)
		return named && className(n) && (strings.EqualFold(class.Value, "self") || strings.EqualFold(class.Value, "static"))
	}
	return false
}

// unbound reports an argument that names no class: text written without
// `::`, or a function written in place.
func unbound(x syntax.Expr) bool {
	switch syntax.UnwrapParens(x).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
		return true
	}
	text, known := literalText(x)
	return known && !strings.Contains(text, "::")
}

// filters reports arguments of a filter function that may ask for a
// callback: `FILTER_CALLBACK` is written in one of them, or the filter, the
// argument at index at, is not written as constants.
func filters(args []syntax.Expr, at int) bool {
	for _, arg := range args {
		named := false
		syntax.Inspect(arg, func(n syntax.Node) bool {
			if c, ok := n.(*syntax.ConstFetch); ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "FILTER_CALLBACK") {
				named = true
			}
			return !named
		})
		if named {
			return true
		}
	}
	return at < len(args) && !constants(value(args[at]))
}

// constants reports an expression written with constants and literals only,
// arrays of them included.
func constants(x syntax.Expr) bool {
	fixed := true
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.ConstFetch, *syntax.ClassConstFetch, *syntax.Literal, *syntax.Binary, *syntax.Paren, *syntax.Name, *syntax.Identifier,
			*syntax.Array, *syntax.ArrayItem:
		default:
			fixed = false
		}
		return fixed
	})
	return fixed
}
