package semanticquery

import (
	"strings"

	"custos/internal/php/syntax"
)

// infallibleStreams are the PHP streams an fopen() cannot fail to open:
// memory, temporary and standard streams.
var infallibleStreams = []string{"php://memory", "php://temp", "php://input", "php://output", "php://stdin", "php://stdout", "php://stderr"}

// InfallibleOpen reports whether fopen call opens one of PHP's own streams
// named by a literal: php://memory (with or without options), php://temp
// followed by anything (PHP only reads the prefix: `php://temp,` opens),
// php://input, php://output and the standard streams.
func InfallibleOpen(call *syntax.FuncCall) bool {
	lit, ok := syntax.UnwrapParens(CallArgument(call.Args, 0, "filename")).(*syntax.Literal)
	if !ok {
		return false
	}
	name, ok := syntax.LiteralStringValue(lit)
	if !ok {
		return false
	}
	name = strings.ToLower(name)
	for _, s := range infallibleStreams {
		if name == s || s == "php://memory" && strings.HasPrefix(name, s+"/") || s == "php://temp" && strings.HasPrefix(name, s) {
			return true
		}
	}
	return false
}
