package util

import (
	"bytes"

	"custos/internal/syntax"
)

// IsFuncLike reports whether n is a function, method, closure or arrow
// function.
func IsFuncLike(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction:
		return true
	}
	return false
}

// EnclosingFuncLike returns the nearest function, method, closure or arrow
// function strictly enclosing n (nil in top-level code).
func EnclosingFuncLike(n syntax.Node) syntax.Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if IsFuncLike(p) {
			return p
		}
	}
	return nil
}

// FuncLikeBody returns the braced body of a function-like (nil for arrow
// functions and abstract methods).
func FuncLikeBody(n syntax.Node) *syntax.Block {
	switch f := n.(type) {
	case *syntax.Function:
		return f.Body
	case *syntax.Method:
		return f.Body
	case *syntax.Closure:
		return f.Body
	}
	return nil
}

// FuncLikeParams returns the parameters of a function-like.
func FuncLikeParams(n syntax.Node) []*syntax.Param {
	switch f := n.(type) {
	case *syntax.Function:
		return f.Params
	case *syntax.Method:
		return f.Params
	case *syntax.Closure:
		return f.Params
	case *syntax.ArrowFunction:
		return f.Params
	}
	return nil
}

// Equivalent reports whether a and b are the same expression: same node
// kind and, for two simple variables, the same name; otherwise the same
// significant token sequence (whitespace and comments ignored).
func Equivalent(f *syntax.File, a, b syntax.Node) bool {
	if a == nil || b == nil || a.Kind() != b.Kind() {
		return false
	}
	if va, ok := a.(*syntax.Variable); ok {
		vb := b.(*syntax.Variable)
		if va.NameExpr == nil && vb.NameExpr == nil {
			return va.Name == vb.Name
		}
	}
	sa, sb := a.Span(), b.Span()
	if bytes.Equal(f.Src[sa.Start:sa.End], f.Src[sb.Start:sb.End]) {
		return true
	}
	i, j := TokenIndex(f, sa.Start), TokenIndex(f, sb.Start)
	for {
		i = skipTrivia(f, i, sa.End)
		j = skipTrivia(f, j, sb.End)
		endA := i >= len(f.Tokens) || f.Tokens[i].End > sa.End
		endB := j >= len(f.Tokens) || f.Tokens[j].End > sb.End
		if endA || endB {
			return endA && endB
		}
		ta, tb := f.Tokens[i], f.Tokens[j]
		if ta.Kind != tb.Kind || !bytes.Equal(f.Src[ta.Start:ta.End], f.Src[tb.Start:tb.End]) {
			return false
		}
		i++
		j++
	}
}

func skipTrivia(f *syntax.File, i int, end uint32) int {
	for i < len(f.Tokens) && f.Tokens[i].End <= end && f.Tokens[i].Kind.IsTrivia() {
		i++
	}
	return i
}

// LastNamePart returns the last segment of a written name (`\A\b` -> `b`).
func LastNamePart(written string) string {
	for i := len(written) - 1; i >= 0; i-- {
		if written[i] == '\\' {
			return written[i+1:]
		}
	}
	return written
}

// QuotedStringContent returns the raw content between the quotes of a
// single- or double-quoted string literal without interpolation.
func QuotedStringContent(e syntax.Node) (string, bool) {
	l, ok := e.(*syntax.Literal)
	if !ok || l.LitKind != syntax.LitString || len(l.Raw) < 2 {
		return "", false
	}
	q := l.Raw[0]
	if (q != '\'' && q != '"') || l.Raw[len(l.Raw)-1] != q {
		return "", false
	}
	return l.Raw[1 : len(l.Raw)-1], true
}
