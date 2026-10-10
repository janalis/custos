// Package pregmatchalllayoutmismatch implements the native PregMatchAllLayoutMismatch inspection.
package pregmatchalllayoutmismatch

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

const message = "Use indices matching the selected regex result layout."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PregMatchAllLayoutMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	outer := n.(*syntax.ArrayDimFetch)
	if guarded(outer) || !astquery.StableArrayRead(ctx.File, outer) {
		return
	}
	inner, ok := outer.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	variable, ok := inner.Var.(*syntax.Variable)
	if !ok {
		return
	}
	first, known := semanticquery.NativeInt(ctx, inner.Dim)
	if !known || first < 0 {
		return
	}
	second, known := semanticquery.NativeInt(ctx, outer.Dim)
	if !known || second < 0 {
		return
	}
	c := priorRegex(ctx, outer, variable, "preg_match_all")
	if c == nil {
		return
	}
	flags := semanticquery.CallArgument(c.Args, 3, "flags")
	mode := int64(1)
	if flags != nil {
		mode, known = semanticquery.NativeInt(ctx, flags)
		if !known {
			return
		}
	}
	if mode != 1 && mode != 2 {
		return
	}
	matches, captures, known := literalMatches(ctx, c)
	if !known {
		return
	}
	rows, columns := int64(captures+1), int64(len(matches))
	if mode == 2 {
		rows, columns = columns, rows
	}
	if first >= rows || second >= columns {
		ctx.ReportNode(outer, message)
	}
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

func guarded(at syntax.Node) bool {
	for p := at.Parent(); p != nil; p = p.Parent() {
		switch x := p.(type) {
		case *syntax.Isset, *syntax.Empty:
			return true
		case *syntax.Binary:
			if x.Op.Kind == syntax.TCoalesce {
				return true
			}
		}
		if _, ok := p.(syntax.Stmt); ok {
			break
		}
	}
	return false
}
