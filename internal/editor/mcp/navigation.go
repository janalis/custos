package mcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// canonicalRule resolves a rule or legacy ID regardless of case.
func canonicalRule(id string) (string, error) {
	rules, err := ruleCatalogue()
	if err != nil {
		return "", err
	}
	var near []string
	for _, r := range rules {
		if strings.EqualFold(id, r.ID) || strings.EqualFold(id, r.LegacyID) {
			return r.ID, nil
		}
		if len(near) < 5 && strings.Contains(strings.ToLower(r.ID), strings.ToLower(id)) {
			near = append(near, r.ID)
		}
	}
	if len(near) > 0 {
		return "", fmt.Errorf("unknown rule %q; did you mean %s?", id, strings.Join(near, ", "))
	}
	return "", fmt.Errorf("unknown rule %q; search with custos_rules", id)
}

// words splits CamelCase so "null coalescing" matches NullCoalescing.
func words(id string) string {
	var b strings.Builder
	for i, r := range id {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func overrideView(rc analysis.RuleConfig) Output {
	out := Output{}
	if rc.Enabled != nil {
		out["enabled"] = *rc.Enabled
	}
	if rc.Severity != "" {
		out["severity"] = rc.Severity
	}
	options := Output{}
	for name, value := range rc.Options {
		// "@"-prefixed options are derived by custos, not user settings.
		if !strings.HasPrefix(name, "@") {
			options[name] = value
		}
	}
	if len(options) > 0 {
		out["options"] = options
	}
	return out
}

func settingsView(cfg *config.Config) Output {
	style := "regular"
	if cfg.ComparisonStyle == analysis.StyleYoda {
		style = "yoda"
	}
	rules := Output{}
	for id, rc := range cfg.Rules {
		if view := overrideView(rc); len(view) > 0 {
			rules[id] = view
		}
	}
	return Output{"root": cfg.Root, "php": cfg.PHP.String(), "phpSource": cfg.PHPSource, "comparisonStyle": style, "shortOpenTag": cfg.ShortOpenTag, "paths": cfg.Paths, "exclude": cfg.Exclude, "baseline": cfg.Baseline, "rules": rules}
}

func (s *Server) ruleInfo(ctx context.Context, w *workspace, in Input) (Output, error) {
	rules, err := ruleCatalogue()
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadConfined(w.root)
	if err != nil {
		return nil, err
	}
	query := strings.Fields(strings.ToLower(in.Query))
	var items []any
	for _, r := range rules {
		d, _ := meta.Describe(r.ID)
		if in.Rule != "" && !strings.EqualFold(in.Rule, r.ID) && !strings.EqualFold(in.Rule, r.LegacyID) {
			continue
		}
		haystack := strings.ToLower(r.ID + " " + words(r.ID) + " " + r.Group + " " + d.Summary)
		matched := true
		for _, word := range query {
			matched = matched && strings.Contains(haystack, word)
		}
		if !matched {
			continue
		}
		item := Output{"rule": r, "explanation": d, "targetPHP": cfg.PHP.String(), "suppressions": []string{"@custos-ignore " + r.ID, "@noinspection " + r.LegacyID}}
		if view := overrideView(cfg.Rules[r.ID]); len(view) > 0 {
			item["override"] = view
		}
		items = append(items, item)
	}
	return page(ctx, w.ID+"rules", in, items)
}

type symbol struct {
	name, path, kind, signature, returns string
	span                                 syntax.Span
	rank                                 int
	vendor, magic                        bool
}

// rank orders matches: exact name, a whole trailing segment, any suffix,
// then any substring. A negative rank is no match.
func rank(name, query string) int {
	name = strings.ToLower(name)
	switch {
	case name == query:
		return 0
	case strings.HasSuffix(name, `\`+query) || strings.HasSuffix(name, "::"+query):
		return 1
	case strings.HasSuffix(name, query):
		return 2
	case strings.Contains(name, query):
		return 3
	}
	return -1
}

func parameters(params []index.Param) string {
	var parts []string
	for _, p := range params {
		var b strings.Builder
		if t := firstNonEmpty(p.Type, p.DocType); t != "" {
			b.WriteString(t + " ")
		}
		if p.ByRef {
			b.WriteByte('&')
		}
		if p.Variadic {
			b.WriteString("...")
		}
		b.WriteString("$" + p.Name)
		if p.Default != "" {
			b.WriteString(" = " + p.Default)
		}
		parts = append(parts, b.String())
	}
	return strings.Join(parts, ", ")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// returnType reports the best known return type and where it comes from.
func returnType(native, doc, inferred string) (suffix, source string) {
	switch {
	case native != "":
		return ": " + native, "native"
	case doc != "":
		return ": " + doc, "phpdoc"
	case inferred != "":
		return ": " + inferred, "inferred"
	}
	return "", ""
}

func classSignature(c *index.Class) string {
	var b strings.Builder
	if c.Abstract {
		b.WriteString("abstract ")
	}
	if c.Final {
		b.WriteString("final ")
	}
	if c.Readonly {
		b.WriteString("readonly ")
	}
	b.WriteString([]string{"class ", "interface ", "trait ", "enum "}[c.Kind] + c.FQN)
	if c.Parent != "" {
		b.WriteString(" extends " + c.Parent)
	}
	if len(c.Interfaces) > 0 {
		b.WriteString(" implements " + strings.Join(c.Interfaces, ", "))
	}
	return b.String()
}

func methodSignature(m *index.Method) (string, string) {
	var b strings.Builder
	if m.Abstract {
		b.WriteString("abstract ")
	}
	if m.Final {
		b.WriteString("final ")
	}
	b.WriteString([]string{"public ", "protected ", "private "}[m.Visibility])
	if m.Static {
		b.WriteString("static ")
	}
	suffix, source := returnType(m.Return, m.DocReturn, m.Inferred)
	b.WriteString("function " + m.Name + "(" + parameters(m.Params) + ")" + suffix)
	return b.String(), source
}

func constantSignature(c *index.Constant) string {
	if c.Value == "" {
		return "const " + c.FQN
	}
	return "const " + c.FQN + " = " + c.Value
}

func (s *Server) symbols(ctx context.Context, cp *checkpoint, snap *project.Snapshot, in Input) (Output, error) {
	query := strings.ToLower(strings.TrimPrefix(in.Query, `\`))
	scope := in.Scope
	if scope == "" {
		scope = "all"
		if query == "" {
			scope = "project"
		}
	}
	own := make(map[string]bool, len(snap.OwnFiles))
	for _, path := range snap.OwnFiles {
		own[path] = true
	}
	var found []symbol
	counts := map[string]int{}
	add := func(sym symbol) {
		sym.vendor = !own[sym.path]
		sym.rank = rank(sym.name, query)
		if sym.rank < 0 || (in.Path != "" && in.Path != sym.path) || (in.Kind != "" && in.Kind != sym.kind) || (scope == "project" && sym.vendor) || (scope == "vendor" && !sym.vendor) {
			return
		}
		counts[strings.ToLower(sym.name)]++
		found = append(found, sym)
	}
	for _, fs := range snap.Symbols {
		for _, c := range fs.Classes {
			if c.Anonymous {
				continue
			}
			add(symbol{name: c.FQN, path: fs.Path, kind: "class", span: c.Span, signature: classSignature(c)})
			for _, m := range c.Methods {
				signature, source := methodSignature(m)
				add(symbol{name: c.FQN + "::" + m.Name, path: fs.Path, kind: "method", span: m.Span, signature: signature, returns: source, magic: m.Magic})
			}
		}
		for _, f := range fs.Functions {
			suffix, source := returnType(f.Return, f.DocReturn, f.Inferred)
			add(symbol{name: f.FQN, path: fs.Path, kind: "function", span: f.Span, signature: "function " + f.FQN + "(" + parameters(f.Params) + ")" + suffix, returns: source})
		}
		for _, c := range fs.Constants {
			add(symbol{name: c.FQN, path: fs.Path, kind: "constant", span: c.Span, signature: constantSignature(c)})
		}
	}
	sort.Slice(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		if a.vendor != b.vendor {
			return b.vendor
		}
		if a.name != b.name {
			return a.name < b.name
		}
		return a.path < b.path
	})
	// Only the returned page needs positions: vendor sources are read on demand.
	var li *syntax.LineIndex
	indexed := ""
	out, err := pageLazy(ctx, cp.ID+"symbols", in, len(found), func(i int) any {
		sym := found[i]
		item := Output{"name": sym.name, "path": sym.path, "kind": sym.kind, "signature": sym.signature, "provenance": "project", "ambiguous": counts[strings.ToLower(sym.name)] > 1}
		if sym.vendor {
			item["provenance"] = "vendor"
		}
		if sym.returns != "" {
			item["returnTypeFrom"] = sym.returns
		}
		if sym.magic {
			item["declaredBy"] = "@method PHPDoc tag"
			return item
		}
		if indexed != sym.path {
			li, indexed = nil, sym.path
			if src, err := snap.Source(sym.path); err == nil {
				li = syntax.NewLineIndex(src)
			}
		}
		item["start"], item["end"] = sym.span.Start, sym.span.End
		if li != nil {
			line, column := li.RuneColumn(sym.span.Start)
			item["line"], item["column"] = line+1, column+1
		}
		return item
	})
	if err == nil {
		out["scope"] = scope
		out["limitations"] = "Static declarations only: runtime declarations, anonymous classes and PHP builtins are not listed."
	}
	return out, err
}

// offsetAt converts a one-based line and rune column to a byte offset.
func offsetAt(src []byte, line, column int) (uint32, error) {
	li := syntax.NewLineIndex(src)
	if line < 1 || line > li.Lines() || column < 1 {
		return 0, errors.New("line or column outside the file")
	}
	offset := int(li.OffsetUTF16(line-1, 0))
	for ; column > 1 && offset < len(src) && src[offset] != '\n' && src[offset] != '\r'; column-- {
		_, size := utf8.DecodeRune(src[offset:])
		offset += size
	}
	return uint32(offset), nil
}

func (s *Server) typeAt(snap *project.Snapshot, in Input) (Output, error) {
	src, err := snap.Source(in.Path)
	if err != nil || !strings.HasSuffix(in.Path, ".php") {
		return nil, fmt.Errorf("path is not an analyzed or vendor PHP file: %s", in.Path)
	}
	offset := in.Offset
	if in.Line > 0 {
		if offset, err = offsetAt(src, in.Line, max(in.Column, 1)); err != nil {
			return nil, err
		}
	}
	if uint64(offset) >= uint64(len(src)) {
		return nil, errors.New("byte offset outside the file")
	}
	file := syntax.ParseBest(in.Path, src, syntax.Options{Version: snap.Config.PHP, ShortOpenTag: snap.Config.ShortOpenTag})
	var expr syntax.Expr
	syntax.InspectFile(file, func(n syntax.Node) bool {
		span := n.Span()
		if offset < span.Start || offset >= span.End {
			return false
		}
		switch n.(type) {
		case *syntax.Identifier, *syntax.Name, *syntax.Arg, *syntax.ArgList:
			return true
		}
		if e, ok := n.(syntax.Expr); ok {
			if expr == nil || span.End-span.Start < expr.Span().End-expr.Span().Start {
				expr = e
			}
		}
		return true
	})
	if expr == nil {
		return Output{"resolved": false, "reason": "No expression at this position."}, nil
	}
	env := infer.NewEnv(file, names.New(file), snap.Index, snap.Config.PHP)
	doc := env.TypeOf(expr)
	native := env.Native().TypeOf(expr)
	line, column := syntax.NewLineIndex(src).RuneColumn(expr.Span().Start)
	return Output{"native": native.String(), "phpDocAware": doc.String(), "resolved": !doc.IsUnknown(), "expression": excerpt(src, expr.Span().Start, expr.Span().End, 0, 120), "start": expr.Span().Start, "end": expr.Span().End, "line": line + 1, "column": column + 1, "limitations": "Static inference; PHPDoc is untrusted, runtime and ambiguous declarations can limit precision.", "syntaxErrors": len(file.Errors)}, nil
}

// excerpt quotes src[start-context:end+context], at most limit bytes when
// limit is positive, widened or narrowed to whole UTF-8 sequences.
func excerpt(src []byte, start, end uint32, context, limit int) string {
	lo := min(len(src), max(0, int(start)-context))
	hi := max(lo, min(len(src), int(end)+context))
	if limit > 0 {
		hi = min(hi, lo+limit)
	}
	for lo < hi && !utf8.RuneStart(src[lo]) {
		lo++
	}
	for hi > lo && hi < len(src) && !utf8.RuneStart(src[hi]) {
		hi--
	}
	return string(src[lo:hi])
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

var ruleCatalogue = meta.All
