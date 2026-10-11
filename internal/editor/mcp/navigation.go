package mcp

import (
	"context"
	"errors"
	"sort"
	"strings"

	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/project/config"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func (s *Server) ruleInfo(ctx context.Context, w *workspace, in Input) (Output, error) {
	rules, err := ruleCatalogue()
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadConfined(w.root)
	if err != nil {
		return nil, err
	}
	var items []any
	for _, r := range rules {
		d, _ := meta.Describe(r.ID)
		if in.Rule != "" && in.Rule != r.ID && in.Rule != r.LegacyID {
			continue
		}
		if in.Query != "" && !strings.Contains(strings.ToLower(r.ID+" "+r.Group+" "+d.Summary), strings.ToLower(in.Query)) {
			continue
		}
		items = append(items, Output{"rule": r, "explanation": d, "effectiveOverride": cfg.Rules[r.ID], "targetPHP": cfg.PHP.String(), "suppressions": []string{"@custos-ignore " + r.ID, "@noinspection " + r.LegacyID}, "versionRequirements": Output{"min": d.PHPMin, "max": d.PHPMax}})
	}
	return page(ctx, w.ID+"rules", in, items)
}

func (s *Server) symbols(ctx context.Context, cp *checkpoint, snap *project.Snapshot, in Input) (Output, error) {
	var items []any
	counts := map[string]int{}
	match := func(name, path string) bool {
		return (in.Path == "" || in.Path == path) && strings.Contains(strings.ToLower(name), strings.ToLower(in.Query))
	}
	own := make(map[string]bool, len(snap.OwnFiles))
	for _, path := range snap.OwnFiles {
		own[path] = true
	}
	// Declarations arrive grouped by file: index each file's lines once.
	var li *syntax.LineIndex
	indexed := ""
	add := func(name, path, kind string, span syntax.Span, signature any) {
		if !match(name, path) {
			return
		}
		counts[strings.ToLower(name)]++
		if li == nil || indexed != path {
			li, indexed = syntax.NewLineIndex(snap.Sources[path]), path
		}
		line, column := li.RuneColumn(span.Start)
		provenance := "project"
		if !own[path] {
			provenance = "vendor"
		}
		items = append(items, Output{"name": name, "path": path, "kind": kind, "start": span.Start, "end": span.End, "line": line + 1, "column": column + 1, "signature": signature, "provenance": provenance})
	}
	for _, fs := range snap.Symbols {
		for _, c := range fs.Classes {
			add(c.FQN, fs.Path, "class", c.Span, Output{"parent": c.Parent, "interfaces": c.Interfaces, "traits": c.Traits})
			for _, m := range c.Methods {
				add(c.FQN+"::"+m.Name, fs.Path, "method", m.Span, m)
			}
		}
		for _, f := range fs.Functions {
			add(f.FQN, fs.Path, "function", f.Span, f)
		}
		for _, c := range fs.Constants {
			add(c.FQN, fs.Path, "constant", c.Span, c)
		}
	}
	for _, item := range items {
		o := item.(Output)
		o["ambiguous"] = counts[strings.ToLower(o["name"].(string))] > 1
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i].(Output), items[j].(Output)
		if a["name"] != b["name"] {
			return a["name"].(string) < b["name"].(string)
		}
		return a["path"].(string) < b["path"].(string)
	})
	out, err := page(ctx, cp.ID+"symbols", in, items)
	if err == nil {
		out["limitations"] = "Declaration search includes project and vendor; dynamic runtime declarations and embedded builtin search are not included."
	}
	return out, err
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

func (s *Server) typeAt(snap *project.Snapshot, in Input) (Output, error) {
	src, ok := snap.Sources[in.Path]
	if !ok || !strings.HasSuffix(in.Path, ".php") || uint64(in.Offset) >= uint64(len(src)) {
		return nil, errors.New("path or byte offset unavailable")
	}
	file := syntax.ParseBest(in.Path, src, syntax.Options{Version: snap.Config.PHP, ShortOpenTag: snap.Config.ShortOpenTag})
	var expr syntax.Expr
	syntax.InspectFile(file, func(n syntax.Node) bool {
		span := n.Span()
		if in.Offset < span.Start || in.Offset >= span.End {
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
		return Output{"resolved": false, "reason": "No expression at this byte offset."}, nil
	}
	env := infer.NewEnv(file, names.New(file), snap.Index, snap.Config.PHP)
	doc := env.TypeOf(expr)
	native := env.Native().TypeOf(expr)
	line, column := syntax.NewLineIndex(src).RuneColumn(expr.Span().Start)
	return Output{"native": native.String(), "phpDocAware": doc.String(), "resolved": !doc.IsUnknown(), "start": expr.Span().Start, "end": expr.Span().End, "line": line + 1, "column": column + 1, "limitations": "Static inference; PHPDoc is untrusted, runtime and ambiguous declarations can limit precision.", "syntaxErrors": len(file.Errors)}, nil
}

var ruleCatalogue = meta.All
