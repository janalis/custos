// Package phpdoc parses PHPDoc comments: tags and type expressions.
package phpdoc

import (
	"strings"
)

// Tag is one `@name rest` entry.
type Tag struct {
	Name string // without '@', e.g. "return"
	Text string // rest of the tag (continuation lines joined with '\n')
}

// Doc is a parsed doc comment.
type Doc struct {
	Summary string
	Tags    []Tag
}

// Parse parses a /** ... */ comment (or any comment text).
func Parse(comment string) *Doc {
	body := strings.TrimSpace(comment)
	body = strings.TrimPrefix(body, "/**")
	body = strings.TrimPrefix(body, "/*")
	body = strings.TrimSuffix(body, "*/")
	d := &Doc{}
	var summary []string
	// continuation lines per tag, joined at the end (appending to Text
	// line by line was quadratic on long tags)
	var cont [][]string
	cur := -1
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@") {
			name := line[1:]
			text := ""
			if i := strings.IndexAny(name, " \t("); i >= 0 {
				text = strings.TrimSpace(name[i:])
				name = name[:i]
			}
			d.Tags = append(d.Tags, Tag{Name: name, Text: text})
			cont = append(cont, nil)
			cur = len(d.Tags) - 1
			continue
		}
		if cur >= 0 {
			if line != "" {
				cont[cur] = append(cont[cur], line)
			}
			continue
		}
		summary = append(summary, line)
	}
	for i, lines := range cont {
		if len(lines) > 0 {
			d.Tags[i].Text += "\n" + strings.Join(lines, "\n")
		}
	}
	d.Summary = strings.TrimSpace(strings.Join(summary, "\n"))
	return d
}

// Tag returns the first tag with the given name (also matching psalm-/phpstan- prefixed variants when prefixed is true).
func (d *Doc) Tag(name string) (Tag, bool) {
	for _, t := range d.Tags {
		if t.Name == name {
			return t, true
		}
	}
	return Tag{}, false
}

// Has reports whether a tag exists.
func (d *Doc) Has(name string) bool {
	_, ok := d.Tag(name)
	return ok
}

// All returns every tag with the given name.
func (d *Doc) All(name string) []Tag {
	var out []Tag
	for _, t := range d.Tags {
		if t.Name == name {
			out = append(out, t)
		}
	}
	return out
}

// SplitType splits tag text into its leading type expression and the rest
// (balanced over <>, (), {}, []), e.g. "array<int, string> $x desc".
func SplitType(text string) (typ, rest string) {
	text = strings.TrimSpace(text)
	depth := 0
	runEnd := 0 // end of the last scanned blank run (rescanning it per blank was quadratic)
	for i := 0; i < len(text); i++ {
		switch c := text[i]; c {
		case '<', '(', '{', '[':
			depth++
		case '>', ')', '}', ']':
			if depth > 0 {
				depth--
			}
		case ' ', '\t', '\n':
			if depth == 0 {
				// Allow spaces around | and & at top level ("int | null").
				j := i
				if i < runEnd {
					j = runEnd
				} else {
					for j < len(text) && (text[j] == ' ' || text[j] == '\t') {
						j++
					}
					runEnd = j
				}
				if j < len(text) && (text[j] == '|' || text[j] == '&') {
					i = j
					continue
				}
				if i > 0 && (text[i-1] == '|' || text[i-1] == '&' || text[i-1] == ',' || text[i-1] == ':') {
					continue
				}
				return squashType(text[:i]), strings.TrimSpace(text[i:])
			}
		}
	}
	return squashType(text), ""
}

// squashType removes the spaces of a type expression, except in conditional
// types (`(T is null ? A : B)`), where they separate words: there runs of
// whitespace collapse to one space.
func squashType(t string) string {
	if strings.Contains(t, "(") && strings.Contains(t, " is ") {
		return strings.Join(strings.Fields(t), " ")
	}
	return strings.ReplaceAll(t, " ", "")
}

// VarName extracts a leading `$name` (without '$') from text.
func VarName(text string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "...")
	text = strings.TrimPrefix(text, "&")
	if !strings.HasPrefix(text, "$") {
		return ""
	}
	end := 1
	for end < len(text) && (isIdent(text[end])) {
		end++
	}
	return text[1:end]
}

func isIdent(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c >= 0x80
}

// Param is a parsed @param tag.
type Param struct {
	Type string
	Name string
}

// Params returns all @param tags.
func (d *Doc) Params() []Param { return d.ParamsOf("param") }

// ParamsOf returns the tags named tag ("param", "phpstan-param",
// "psalm-param") parsed as @param tags.
func (d *Doc) ParamsOf(tag string) []Param {
	var out []Param
	for _, t := range d.All(tag) {
		typ, rest := SplitType(t.Text)
		if strings.HasPrefix(typ, "$") { // "@param $x" without type, or "@param $x Type"
			out = append(out, Param{Name: VarName(typ), Type: trailingType(rest)})
			continue
		}
		out = append(out, Param{Type: typ, Name: VarName(rest)})
	}
	return out
}

// ReturnType returns the @return type text ("" when absent).
func (d *Doc) ReturnType() string {
	if t, ok := d.Tag("return"); ok {
		typ, _ := SplitType(t.Text)
		return typ
	}
	return ""
}

// VarType returns the @var type for variable name (or the first @var when
// name is "" or the tag names no variable).
func (d *Doc) VarType(name string) string {
	for _, t := range d.All("var") {
		typ, rest := SplitType(t.Text)
		if strings.HasPrefix(typ, "$") { // "@var $x Type"
			if name == "" || VarName(typ) == name {
				t2, _ := SplitType(rest)
				return t2
			}
			continue
		}
		v := VarName(rest)
		if name == "" || v == "" || v == name {
			return typ
		}
	}
	return ""
}

// trailingType returns the type of a "@param $x Type [desc]" tag (name
// first, as PhpStorm accepts): the first word of rest when it is the only
// word or clearly looks like a type, "" otherwise (plain description).
func trailingType(rest string) string {
	typ, desc := SplitType(rest)
	if typ == "" {
		return ""
	}
	if desc == "" || strings.ContainsAny(typ, `|&[]<\?`) || (typ[0] >= 'A' && typ[0] <= 'Z') || docBuiltin[strings.ToLower(typ)] {
		return typ
	}
	return ""
}

var docBuiltin = map[string]bool{
	"int": true, "integer": true, "float": true, "double": true, "string": true, "bool": true,
	"boolean": true, "true": true, "false": true, "null": true, "array": true, "callable": true,
	"iterable": true, "object": true, "mixed": true, "void": true, "resource": true, "self": true,
	"static": true,
}

// TemplateParam is a declared template: `@template T of Bound`.
type TemplateParam struct {
	Name    string
	Bound   string // type text after `of` / `as`; "" when absent
	Default string // type text after `=` (`@template T of array-key = int`); "" when absent
}

// TemplateParams returns the templates declared by @template (and the
// variants Templates accepts) with their bounds, in declaration order.
func (d *Doc) TemplateParams() []TemplateParam {
	var out []TemplateParam
	for _, t := range d.Tags {
		if !isTemplateTag(t.Name) {
			continue
		}
		f := strings.Fields(t.Text)
		if len(f) == 0 {
			continue
		}
		p := TemplateParam{Name: f[0]}
		rest := strings.TrimSpace(strings.TrimSpace(t.Text)[len(f[0]):])
		if len(f) > 2 && (f[1] == "of" || f[1] == "as") {
			p.Bound, rest = SplitType(strings.TrimSpace(rest[len(f[1]):]))
			rest = strings.TrimSpace(rest)
		}
		if strings.HasPrefix(rest, "=") {
			p.Default, _ = SplitType(strings.TrimSpace(rest[1:]))
		}
		out = append(out, p)
	}
	return out
}

func isTemplateTag(name string) bool {
	switch name {
	case "template", "template-covariant", "template-contravariant",
		"psalm-template", "psalm-template-covariant", "phpstan-template", "phpstan-template-covariant":
		return true
	}
	return false
}

// Templates returns the names declared by @template (and the psalm-/phpstan-
// prefixed and covariant/contravariant variants).
func (d *Doc) Templates() []string {
	var out []string
	for _, t := range d.Tags {
		switch t.Name {
		case "template", "template-covariant", "template-contravariant",
			"psalm-template", "psalm-template-covariant", "phpstan-template", "phpstan-template-covariant":
			if f := strings.Fields(t.Text); len(f) > 0 {
				out = append(out, f[0])
			}
		}
	}
	return out
}

// MaxAliasLen is the longest type alias definition kept (the same cap as
// types.MaxDocTypeLen): a longer one is recorded as "" (mixed), so that
// resolvers do not copy a huge definition at every use of the alias.
const MaxAliasLen = 4096

// TypeAliases returns local type aliases declared with @phpstan-type /
// @psalm-type (name -> definition text; "" beyond MaxAliasLen) and
// imported ones with @phpstan-import-type / @psalm-import-type (name -> "").
func (d *Doc) TypeAliases() map[string]string {
	var out map[string]string
	add := func(k, v string) {
		if out == nil {
			out = map[string]string{}
		}
		out[k] = v
	}
	for _, t := range d.Tags {
		switch t.Name {
		case "phpstan-type", "psalm-type":
			text := strings.TrimSpace(t.Text)
			name, def := text, ""
			if i := strings.IndexAny(text, " \t=\n"); i >= 0 {
				name, def = text[:i], strings.TrimSpace(strings.TrimLeft(text[i:], " \t="))
			}
			if name != "" {
				if def = strings.Join(strings.Fields(def), " "); len(def) > MaxAliasLen {
					def = ""
				}
				add(name, def)
			}
		case "phpstan-import-type", "psalm-import-type":
			f := strings.Fields(t.Text)
			if len(f) == 0 {
				continue
			}
			name := f[0]
			for i := 1; i+1 < len(f); i++ {
				if f[i] == "as" {
					name = f[i+1]
				}
			}
			add(name, "")
		}
	}
	return out
}
