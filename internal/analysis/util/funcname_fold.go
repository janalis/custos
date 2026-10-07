package util

import (
	"strings"

	"custos/internal/syntax"
)

// IsFuncNamedFold is IsFuncNamed with a case-insensitive comparison of the
// name part, as PHP compares function names. part must be lower-case.
func IsFuncNamedFold(e syntax.Node, part string) (*syntax.FuncCall, bool) {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return nil, false
	}
	_, p, ok := FuncNamePart(call)
	if !ok || !strings.EqualFold(p, part) {
		return nil, false
	}
	return call, true
}
