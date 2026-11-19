package index

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// phpStormAttr returns the attribute of attrs named name in PhpStorm's
// attribute namespace, nil when there is none.
func (x *extractor) phpStormAttr(attrs []*syntax.AttributeGroup, name string) *syntax.Attribute {
	for _, g := range attrs {
		for _, a := range g.Attrs {
			fqn := strings.TrimPrefix(x.r.Class(a.Name.Value, a.Span().Start), `\`)
			if strings.EqualFold(fqn, `JetBrains\PhpStorm\`+name) {
				return a
			}
		}
	}
	return nil
}

// attrArg returns the argument of a named name, or the one at position pos
// when it is not passed by name; nil when absent.
func attrArg(a *syntax.Attribute, name string, pos int) syntax.Expr {
	if a.Args == nil {
		return nil
	}
	for i, e := range a.Args.Args {
		arg, ok := e.(*syntax.Arg)
		if !ok {
			continue
		}
		if arg.Name != nil {
			if arg.Name.Value == name {
				return arg.Value
			}
			continue
		}
		if i == pos {
			return arg.Value
		}
	}
	return nil
}

// isQuoted reports PHP string literal text.
func isQuoted(s string) bool {
	return len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0]
}

// shapeAttrDoc is the doc type of a #[ArrayShape] or #[ObjectShape]
// attribute of attrs ("" without a readable one). PhpStorm's shapes cannot
// say which keys are optional or whether others exist, and the stubs list
// keys that are only sometimes present, so every key reads as optional and
// the shape as open.
func (x *extractor) shapeAttrDoc(attrs []*syntax.AttributeGroup, name, kind string, at uint32) string {
	a := x.phpStormAttr(attrs, name)
	if a == nil {
		return ""
	}
	arr, ok := attrArg(a, "shape", 0).(*syntax.Array)
	if !ok {
		return ""
	}
	entries := make([]string, 0, len(arr.Items)+1)
	for _, it := range arr.Items {
		if it == nil || it.Key == nil || it.Value == nil {
			return ""
		}
		key := x.text(it.Key)
		if isQuoted(key) {
			key = unquote(key)
		}
		if key == "" || strings.ContainsAny(key, `'"\$`) {
			return ""
		}
		entries = append(entries, "'"+key+"'?:"+x.shapeAttrValue(it.Value))
	}
	entries = append(entries, "...")
	t := x.docType(kind+"{"+strings.Join(entries, ",")+"}", at)
	if doc := t.DocString(); strings.HasPrefix(doc, kind+"{") {
		return doc
	}
	return ""
}

// shapeAttrValue is the doc type text of one value of a shape attribute: a
// string holding a type, `Foo::class`, else mixed.
func (x *extractor) shapeAttrValue(e syntax.Expr) string {
	text := x.text(e)
	if isQuoted(text) {
		if v := unquote(text); v != "" {
			return v
		}
		return "mixed"
	}
	if class, ok := strings.CutSuffix(text, "::class"); ok {
		return class
	}
	return "mixed"
}

// refineDoc applies shape, the doc type of a shape attribute for the kind
// member (array or object), over the documented type doc or else the native
// type: a tag that already has a shape or does not admit kind keeps its say,
// and the other members of the type stay.
func refineDoc(doc, native, kind, shape string) string {
	if shape == "" {
		return doc
	}
	base := doc
	if base == "" {
		base = native
	}
	if base == "" {
		return shape
	}
	t := types.FromDoc(base, nil)
	if !t.Has(kind) || strings.Contains(base, kind+"{") {
		return doc
	}
	if rest := t.Without(kind); !rest.IsUnknown() {
		return shape + "|" + rest.DocString()
	}
	return shape
}

// attrDoc refines doc with the shape attributes of attrs.
func (x *extractor) attrDoc(doc, native string, attrs []*syntax.AttributeGroup, at uint32) string {
	if len(attrs) == 0 {
		return doc
	}
	doc = refineDoc(doc, native, "array", x.shapeAttrDoc(attrs, "ArrayShape", "array", at))
	return refineDoc(doc, native, "object", x.shapeAttrDoc(attrs, "ObjectShape", "object", at))
}

// attrReturnDoc is attrDoc for a return type, where #[NoReturn] also
// documents never unless a tag or a value-returning native type says more.
func (x *extractor) attrReturnDoc(doc, native string, attrs []*syntax.AttributeGroup, at uint32) string {
	doc = x.attrDoc(doc, native, attrs, at)
	if doc == "" && (native == "" || native == "void" || native == "never") && x.phpStormAttr(attrs, "NoReturn") != nil {
		return "never"
	}
	return doc
}

// expectedValues is the value doc of a #[ExpectedValues] attribute on an int
// or string declared as doc or native; "" when it has none or lists anything
// that is not a known value of that type. A nullable declaration keeps null.
func (x *extractor) expectedValues(doc, native string, attrs []*syntax.AttributeGroup, at uint32) string {
	a := x.phpStormAttr(attrs, "ExpectedValues")
	if a == nil {
		return ""
	}
	base := doc
	if base == "" {
		base = native
	}
	declared := types.FromDoc(base, nil)
	plain := declared.Without("null")
	kind := "int"
	if !plain.OnlyOf("int") {
		if !plain.OnlyOf("string") {
			return ""
		}
		kind = "string"
	}
	text := x.expectedValuesText(a)
	if text == "" {
		return ""
	}
	t := x.docType(text, at)
	if !t.OnlyOf(kind) {
		return ""
	}
	values := t.ValueDocString()
	if values != "" && declared.Has("null") {
		return values + "|null"
	}
	return values
}

// expectedValuesText is the doc type text of the literal values a
// #[ExpectedValues] lists: its values, or any combination of its flags. ""
// for anything else: constants have no known value while a file is indexed.
func (x *extractor) expectedValuesText(a *syntax.Attribute) string {
	if arr, ok := attrArg(a, "values", 0).(*syntax.Array); ok {
		return strings.Join(x.expectedList(arr), "|")
	}
	if arr, ok := attrArg(a, "flags", 1).(*syntax.Array); ok {
		if parts := x.expectedList(arr); len(parts) > 0 {
			return "int-mask<" + strings.Join(parts, ",") + ">"
		}
	}
	return ""
}

// expectedList is the doc text of each entry of arr; nil when one of them is
// not a literal.
func (x *extractor) expectedList(arr *syntax.Array) []string {
	parts := make([]string, 0, len(arr.Items))
	for _, it := range arr.Items {
		if it == nil || it.Value == nil {
			return nil
		}
		text := x.text(it.Value)
		if _, err := strconv.ParseInt(text, 0, 64); err != nil && !isQuoted(text) {
			return nil
		}
		parts = append(parts, text)
	}
	return parts
}
