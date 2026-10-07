package util

import (
	"strings"

	"custos/internal/syntax"
)

// FuncNamePart returns the name node of a call to a named function and its
// last segment as written (`\Foo\bar` -> `bar`); ok is false for dynamic
// calls such as `$f()`.
func FuncNamePart(call *syntax.FuncCall) (name *syntax.Name, part string, ok bool) {
	name, ok = call.Name.(*syntax.Name)
	if !ok {
		return nil, "", false
	}
	v := name.Value
	return name, v[strings.LastIndexByte(v, '\\')+1:], true
}

// IsFuncNamed reports whether e is a plain function call whose name part
// equals part exactly (case-sensitive, any namespace qualifier accepted).
func IsFuncNamed(e syntax.Node, part string) (*syntax.FuncCall, bool) {
	call, ok := e.(*syntax.FuncCall)
	if !ok {
		return nil, false
	}
	_, p, ok := FuncNamePart(call)
	if !ok || p != part {
		return nil, false
	}
	return call, true
}

// NamePartSpan returns the span of the last segment of a name (the part after
// the last backslash).
func NamePartSpan(name *syntax.Name) syntax.Span {
	s := name.Span()
	i := strings.LastIndexByte(name.Value, '\\')
	return syntax.Span{Start: s.Start + uint32(i+1), End: s.End}
}

// ArgCount returns the number of entries in a call's argument list, spreads
// and placeholders included (0 without a list).
func ArgCount(call *syntax.FuncCall) int {
	if call.Args == nil {
		return 0
	}
	return len(call.Args.Args)
}
