package flow

import (
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// Byte bounds of the text these instructions build or scan.
const (
	pureTextBytes  = 4096
	pureTrimBytes  = 256
	defaultTrimSet = " \n\r\t\v\x00"
)

// textShape compiles byte-level case changes, trims with a character list and
// padding. Case changes are ASCII-only, hence locale-independent, from PHP 8.2.
func (c *pureCompiler) textShape(node *syntax.FuncCall, name string, params []index.Param) *pureExpr {
	var kind uint8
	size, required := 1, 1
	switch name {
	case "strtolower", "strtoupper", "ucfirst", "lcfirst":
		if c.env.types.PHP.Below(phpversion.PHP82) {
			return nil
		}
		kind = pureCase
	case "trim", "ltrim", "rtrim":
		kind, size = pureTrimChars, 2
		// Stub metadata may repeat the optional character list per version.
		if len(params) == 3 && params[2].Name == params[1].Name && params[1].Optional && params[2].Optional {
			params = params[:2]
		}
	default:
		kind, size, required = pureStrPad, 4, 2
	}
	if len(params) != size {
		return nil
	}
	roles := make([]pureParam, size)
	for i, p := range params {
		roles[i].name = p.Name
	}
	args := c.arguments(node, roles, params)
	if args == nil {
		return nil
	}
	for _, arg := range args[:required] {
		if arg == nil {
			return nil
		}
	}
	mode := &constantScalar{k: 's', s: name}
	switch kind {
	case pureCase:
		child := args[0]
		if child.cost >= pureLimit || child.depth >= MaxDepth {
			return nil
		}
		return &pureExpr{kind: kind, literal: mode, left: child, projects: true, cost: child.cost + 1, depth: child.depth + 1}
	case pureTrimChars:
		set := args[1]
		if set == nil {
			set = pureConstant(constantScalar{k: 's', s: defaultTrimSet})
		}
		result := pureNode(kind, args[0], set, nil)
		if result != nil {
			result.literal, result.projects = mode, true
		}
		return result
	}
	pad, how := args[2], args[3]
	if pad == nil {
		pad = pureConstant(constantScalar{k: 's', s: " "})
	}
	if how == nil {
		how = pureConstant(constantScalar{k: 'i', i: 1})
	}
	result := pureNode(kind, args[0], args[1], pad)
	if result == nil || result.cost+how.cost > pureLimit || how.depth >= MaxDepth {
		return nil
	}
	result.args = []*pureExpr{how}
	result.cost += how.cost
	result.projects = true
	return result
}

// pureBytes reads a proven string of at most limit bytes.
func pureBytes(value Value, limit int) (string, bool) {
	scalar, known := knownScalar(value)
	return scalar.s, known && scalar.k == 's' && len(scalar.s) <= limit && len(value.Sources) <= MaxSources
}

func (e *Env) evalTextShape(expr *pureExpr, left Value, slots []Value, r *scopeResult, depth int) (Value, bool) {
	text, ok := pureBytes(left, pureTextBytes)
	if !ok {
		return Value{}, false
	}
	switch expr.kind {
	case pureCase:
		r.operations += len(text)
		if r.operations > MaxTransfers {
			return Value{}, false
		}
		return pureSliceValue(left, asciiCase(expr.literal.s, text)), true
	case pureTrimChars:
		set, known := e.evalPure(expr.right, slots, r, depth-1)
		cut, ck := pureBytes(set, pureTrimBytes)
		if !known || !ck || strings.Contains(cut, "..") {
			return Value{}, false
		}
		r.operations += len(text) + len(cut)
		if r.operations > MaxTransfers {
			return Value{}, false
		}
		return pureSliceValue(left, trimBytes(expr.literal.s, text, cut)), true
	}
	return e.evalPad(expr, left, text, slots, r, depth)
}

func (e *Env) evalPad(expr *pureExpr, input Value, text string, slots []Value, r *scopeResult, depth int) (Value, bool) {
	length, ok := e.evalPure(expr.right, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	padding, ok := e.evalPure(expr.other, slots, r, depth-1)
	if !ok {
		return Value{}, false
	}
	how, ok := e.evalPure(expr.args[0], slots, r, depth-1)
	width, wk := pureShapeInteger(length)
	mode, mk := pureShapeInteger(how)
	pad, pk := pureBytes(padding, pureTrimBytes)
	if !ok || !wk || !mk || !pk || pad == "" || mode < 0 || mode > 2 || width > pureTextBytes {
		return Value{}, false
	}
	total := int(width) - len(text)
	if total <= 0 {
		return pureSliceValue(input, text), true
	}
	r.operations += int(width)
	r.copies += int(width)
	if r.operations > MaxTransfers || r.copies > MaxCopied {
		return Value{}, false
	}
	before := 0
	switch mode {
	case 0:
		before = total
	case 2:
		before = total / 2
	}
	sources, complete := union(input.Sources, padding.Sources)
	if !complete {
		return Value{}, false
	}
	value := scalarValue(constantScalar{k: 's', s: padBytes(pad, before) + text + padBytes(pad, total-before)})
	value.Sources = sources
	return value, true
}

// padBytes repeats pad from its first byte up to size bytes, as str_pad does on
// each side.
func padBytes(pad string, size int) string {
	var b strings.Builder
	b.Grow(size)
	for b.Len() < size {
		b.WriteString(pad[:min(len(pad), size-b.Len())])
	}
	return b.String()
}

func asciiCase(name, text string) string {
	b := []byte(text)
	for i := range b {
		if i > 0 && (name == "ucfirst" || name == "lcfirst") {
			break
		}
		switch {
		case (name == "strtolower" || name == "lcfirst") && 'A' <= b[i] && b[i] <= 'Z':
			b[i] += 'a' - 'A'
		case (name == "strtoupper" || name == "ucfirst") && 'a' <= b[i] && b[i] <= 'z':
			b[i] -= 'a' - 'A'
		}
	}
	return string(b)
}

// trimBytes removes the bytes of cut, which PHP reads as a byte set.
func trimBytes(name, text, cut string) string {
	var set [256]bool
	for i := 0; i < len(cut); i++ {
		set[cut[i]] = true
	}
	start, end := 0, len(text)
	if name != "rtrim" {
		for start < end && set[text[start]] {
			start++
		}
	}
	if name != "ltrim" {
		for end > start && set[text[end-1]] {
			end--
		}
	}
	return text[start:end]
}
