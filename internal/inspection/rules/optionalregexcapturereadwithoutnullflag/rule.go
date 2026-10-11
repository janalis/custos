// Package optionalregexcapturereadwithoutnullflag implements the native OptionalRegexCaptureReadWithoutNullFlag inspection.
package optionalregexcapturereadwithoutnullflag

import (
	"regexp"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Guard the unmatched optional capture or request null placeholders."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OptionalRegexCaptureReadWithoutNullFlag" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	fetch := n.(*syntax.ArrayDimFetch)
	if !astquery.StableArrayRead(ctx.File, fetch) {
		return
	}
	variable, ok := fetch.Var.(*syntax.Variable)
	if !ok {
		return
	}
	index, known := semanticquery.NativeInt(ctx, fetch.Dim)
	if !known || index < 1 || guarded(ctx, fetch, variable.Name, index) {
		return
	}
	c := priorRegex(ctx, fetch, variable, "preg_match")
	if c == nil {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 3, "flags")
	if flags != nil {
		value, known := semanticquery.NativeInt(ctx, flags)
		if !known || value != 0 {
			return
		}
	}
	matches, captures, known := literalMatches(ctx, c)
	if !known || len(matches) == 0 || index > int64(captures) {
		return
	}
	indices := matches[0]
	if indices[index*2] != -1 {
		return
	}
	for i := int(index + 1); i <= captures; i++ {
		if indices[i*2] != -1 {
			return
		}
	}
	ctx.ReportNode(fetch, message)
}

type regexFacts struct {
	matches  [][]int
	captures int
	known    bool
}

func literalMatches(ctx *analysis.Context, c *syntax.FuncCall) ([][]int, int, bool) {
	facts := ctx.Memo("regex:"+strconv.FormatUint(uint64(c.Span().Start), 10), func() any {
		matches, captures, known := computeMatches(ctx, c)
		return regexFacts{matches, captures, known}
	}).(regexFacts)
	return facts.matches, facts.captures, facts.known
}

func computeMatches(ctx *analysis.Context, c *syntax.FuncCall) ([][]int, int, bool) {
	if offset := semanticquery.CallArgument(c.Args, 4, "offset"); offset != nil {
		n, known := semanticquery.NativeInt(ctx, offset)
		if !known || n != 0 {
			return nil, 0, false
		}
	}
	pattern, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "pattern"))
	if !known || len(pattern) > 32768 || len(pattern) < 2 {
		return nil, 0, false
	}
	count, known := semanticquery.NativeRegexCaptures(pattern)
	if !known {
		return nil, 0, false
	}
	end := strings.LastIndexByte(pattern, pattern[0])
	if end != len(pattern)-1 {
		return nil, 0, false
	}
	body := pattern[1:end]
	if strings.Contains(body, "(?") {
		return nil, 0, false
	}
	for i := 0; i < len(body); i++ {
		if body[i] == '\\' {
			i++
			if i == len(body) || !strings.ContainsRune("dDwW\\.[](){}?*+^$|-", rune(body[i])) {
				return nil, 0, false
			}
		}
	}
	subject, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 1, "subject"))
	if !known || len(subject) > 32768 || strings.ContainsAny(subject, "\r\n") {
		return nil, 0, false
	}
	for _, b := range []byte(subject) {
		if b >= 128 {
			return nil, 0, false
		}
	}
	re, err := regexp.Compile(body)
	if err != nil {
		return nil, 0, false
	}
	if re.MatchString("") {
		return nil, 0, false
	}
	matches := re.FindAllStringSubmatchIndex(subject, 1024)
	if len(matches) == 1024 {
		return nil, 0, false
	}
	return matches, count, true
}

func priorRegex(ctx *analysis.Context, at syntax.Node, variable *syntax.Variable, name string) *syntax.FuncCall {
	prior := flowquery.NativePriorStatements(ctx.File, at)
	if len(prior) == 0 {
		return nil
	}
	st, ok := prior[len(prior)-1].(*syntax.ExprStmt)
	if !ok {
		return nil
	}
	c, ok := st.Expr.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, name) {
		return nil
	}
	output, ok := semanticquery.CallArgument(c.Args, 2, "matches").(*syntax.Variable)
	if !ok || output.Name != variable.Name {
		return nil
	}
	return c
}

// guarded reports reads that cannot observe the absent capture: existence
// probes, by-reference binds, and operands that only evaluate once a guard in
// the same expression has established the capture.
func guarded(ctx *analysis.Context, fetch *syntax.ArrayDimFetch, name string, index int64) bool {
	for child, p := syntax.Node(fetch), fetch.Parent(); p != nil; child, p = p, p.Parent() {
		switch x := p.(type) {
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Assign:
			if x.ByRef && x.Value == child {
				return true
			}
		case *syntax.ArrayItem:
			// Only the value is bound; a key is read as usual.
			if x.ByRef && x.Value == child {
				return true
			}
		case *syntax.Binary:
			switch {
			case x.Left == child:
				if x.Op.Kind == syntax.TCoalesce {
					return true
				}
			case x.Op.Kind == syntax.TBooleanAnd:
				if proves(ctx, x.Left, name, index, true) {
					return true
				}
			case x.Op.Kind == syntax.TBooleanOr:
				if proves(ctx, x.Left, name, index, false) {
					return true
				}
			}
		case *syntax.Ternary:
			if x.Cond != child && proves(ctx, x.Cond, name, index, x.Then == child) {
				return true
			}
		}
		if _, ok := p.(syntax.Stmt); ok {
			break
		}
	}
	return false
}

// proves reports whether cond evaluating to outcome establishes that the
// capture exists.
func proves(ctx *analysis.Context, cond syntax.Expr, name string, index int64, outcome bool) bool {
	switch x := cond.(type) {
	case *syntax.Paren:
		return proves(ctx, x.Expr, name, index, outcome)
	case *syntax.Unary:
		return x.Op.Kind == syntax.TExclaim && proves(ctx, x.Expr, name, index, !outcome)
	case *syntax.Binary:
		if x.Op.Kind == syntax.TBooleanAnd && outcome || x.Op.Kind == syntax.TBooleanOr && !outcome {
			return proves(ctx, x.Left, name, index, outcome) || proves(ctx, x.Right, name, index, outcome)
		}
	case *syntax.Isset:
		if outcome {
			for _, v := range x.Vars {
				if sameCapture(ctx, v, name, index) {
					return true
				}
			}
		}
	case *syntax.Empty:
		return !outcome && sameCapture(ctx, x.Expr, name, index)
	}
	return false
}

func sameCapture(ctx *analysis.Context, e syntax.Expr, name string, index int64) bool {
	fetch, ok := e.(*syntax.ArrayDimFetch)
	if !ok {
		return false
	}
	variable, ok := fetch.Var.(*syntax.Variable)
	if !ok || variable.Name != name {
		return false
	}
	value, known := semanticquery.NativeInt(ctx, fetch.Dim)
	return known && value == index
}
