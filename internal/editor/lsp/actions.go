package lsp

import (
	"encoding/json"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
)

// actionData identifies a fix for lazy resolution.
type actionData struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
	Rule    string `json:"rule,omitempty"`
	Start   uint32 `json:"start,omitempty"`
	End     uint32 `json:"end,omitempty"`
	Fix     int    `json:"fix,omitempty"`
	All     bool   `json:"all,omitempty"`
	// Suppress asks for a `// @custos-ignore <Rule>` comment instead of a fixing.
	Suppress bool `json:"suppress,omitempty"`
}

func (s *Server) codeActions(p codeActionParams) []codeAction {
	s.mu.Lock()
	d, ok := s.docs[p.TextDocument.URI]
	if !ok || d.analyzed != d.version {
		s.mu.Unlock()
		return []codeAction{}
	}
	lines, findings, version := d.lines, d.findings, d.version
	s.mu.Unlock()

	want := func(kind string) bool {
		if len(p.Context.Only) == 0 {
			return true
		}
		for _, k := range p.Context.Only {
			if kind == k || strings.HasPrefix(kind, k+".") {
				return true
			}
		}
		return false
	}
	start := lines.OffsetUTF16(p.Range.Start.Line, p.Range.Start.Character)
	end := lines.OffsetUTF16(p.Range.End.Line, p.Range.End.Character)
	actions := []codeAction{}
	fixable := 0
	// Diagnostics the client sends back identify findings exactly via data.
	named := map[diagData]bool{}
	for _, d := range p.Context.Diagnostics {
		if d.Data != nil {
			named[*d.Data] = true
		}
	}
	var suppress []codeAction
	for _, f := range findings {
		if len(f.Fixes) > 0 {
			fixable++
		}
		inRange := f.Span.End >= start && f.Span.Start <= end
		if !want("quickfix") || (!inRange && !named[diagData{Rule: f.Rule, Start: f.Span.Start, End: f.Span.End}]) {
			continue
		}
		diag := toDiagnostic(lines, f)
		for i, fx := range f.Fixes {
			data := actionData{URI: p.TextDocument.URI, Version: version, Rule: f.Rule, Start: f.Span.Start, End: f.Span.End, Fix: i}
			a := codeAction{Title: fx.Title, Kind: "quickfix", Diagnostics: []protocolDiagnostic{diag}, IsPreferred: i == 0, Data: mustJSON(data)}
			if !s.resolve {
				a.Edit = s.buildEdit(data)
			}
			actions = append(actions, a)
		}
		data := actionData{URI: p.TextDocument.URI, Version: version, Rule: f.Rule, Start: f.Span.Start, End: f.Span.End, Suppress: true}
		a := codeAction{Title: "Suppress " + f.Rule + " for this statement", Kind: "quickfix", Diagnostics: []protocolDiagnostic{diag}, Data: mustJSON(data)}
		// Only offered when it works (a line to annotate, and re-analysis
		// confirms just this finding goes away), even with lazy resolve.
		edit := s.buildEdit(data)
		if edit == nil {
			continue
		}
		if !s.resolve {
			a.Edit = edit
		}
		suppress = append(suppress, a)
	}
	// Fixes first, then the suppressions (less often wanted).
	actions = append(actions, suppress...)
	if fixable > 0 && want(kindFixAll) {
		data := actionData{URI: p.TextDocument.URI, Version: version, All: true}
		a := codeAction{Title: "custos: fix all problems in file", Kind: kindFixAll, Data: mustJSON(data)}
		if !s.resolve {
			a.Edit = s.buildEdit(data)
		}
		actions = append(actions, a)
	}
	return actions
}

func (s *Server) resolveAction(a codeAction) (any, *rpcError) {
	var data actionData
	if err := json.Unmarshal(a.Data, &data); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "bad code action data"}
	}
	a.Edit = s.buildEdit(data)
	if a.Edit == nil {
		return nil, &rpcError{Code: codeRequestFailed, Message: "document changed; fix no longer applies"}
	}
	return a, nil
}

// buildEdit computes the workspace edit for a fix (or fix-all).
func (s *Server) buildEdit(data actionData) *workspaceEdit {
	s.mu.Lock()
	d, ok := s.docs[data.URI]
	if !ok || d.version != data.Version || d.analyzed != d.version {
		s.mu.Unlock()
		return nil
	}
	text, lines, findings, path, e, cfg := d.text, d.lines, d.findings, d.path, s.engine, s.cfg
	s.mu.Unlock()

	var edits []diagnostic.TextEdit
	if data.Suppress {
		edit, ok := suppressEdit(e, path, text, findings, data, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
		if !ok {
			return nil
		}
		edits = []diagnostic.TextEdit{edit}
	} else if data.All {
		res := fixing.FixSource(e, path, text, fixing.Options{Parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}})
		if res.Applied == 0 {
			return nil
		}
		edits = []diagnostic.TextEdit{{Span: syntax.Span{Start: 0, End: uint32(len(text))}, NewText: string(res.Source)}}
	} else {
		for _, f := range findings {
			if f.Rule == data.Rule && f.Span.Start == data.Start && f.Span.End == data.End && data.Fix >= 0 && data.Fix < len(f.Fixes) {
				edits = f.Fixes[data.Fix].Edits()
				break
			}
		}
		if edits == nil {
			return nil
		}
	}
	return docEdit(data.URI, data.Version, lines, edits)
}

func docEdit(uri string, version int, lines *syntax.LineIndex, edits []diagnostic.TextEdit) *workspaceEdit {
	tes := make([]textEdit, 0, len(edits))
	for _, e := range edits {
		tes = append(tes, textEdit{Range: toRange(lines, e.Span), NewText: e.NewText})
	}
	v := version
	return &workspaceEdit{DocumentChanges: []textDocumentEdit{{TextDocument: versionedTextDocumentIdentifier{URI: uri, Version: &v}, Edits: tes}}}
}

func (s *Server) executeCommand(p executeCommandParams) (any, *rpcError) {
	if len(p.Arguments) == 0 {
		return nil, &rpcError{Code: codeInvalidParams, Message: "missing document URI argument"}
	}
	var uri string
	if err := json.Unmarshal(p.Arguments[0], &uri); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "first argument must be a document URI"}
	}
	var only string
	if p.Command == cmdFixRule {
		if len(p.Arguments) < 2 || json.Unmarshal(p.Arguments[1], &only) != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: "custos.fixRule needs (uri, ruleId)"}
		}
	} else if p.Command != cmdFixFile {
		return nil, &rpcError{Code: codeMethodNotFound, Message: "unknown command " + p.Command}
	}
	s.mu.Lock()
	d, ok := s.docs[uri]
	if !ok {
		s.mu.Unlock()
		return nil, &rpcError{Code: codeInvalidParams, Message: "document not open"}
	}
	text, lines, version, path, e, cfg := d.text, d.lines, d.version, d.path, s.engine, s.cfg
	s.mu.Unlock()
	opt := fixing.Options{Parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}}
	if only != "" {
		m, ok := meta.Lookup(only)
		if !ok {
			return nil, &rpcError{Code: codeInvalidParams, Message: "unknown rule " + only}
		}
		opt.Filter = func(f diagnostic.Finding) bool { return f.Rule == m.ID }
	}
	res := fixing.FixSource(e, path, text, opt)
	if res.Applied == 0 {
		return nil, nil
	}
	edit := docEdit(uri, version, lines, []diagnostic.TextEdit{{Span: syntax.Span{Start: 0, End: uint32(len(text))}, NewText: string(res.Source)}})
	if err := s.c.request("workspace/applyEdit", map[string]any{"label": "custos fix", "edit": edit}); err != nil {
		return nil, &rpcError{Code: codeInternalError, Message: err.Error()}
	}
	return nil, nil
}

// suppressEdit builds the `// @custos-ignore` edit for one finding and checks
// it by re-analysing the edited source: exactly that finding must go away
// (a comment before a file's first statement would silence the rule in the
// whole file, so such an edit is refused).
func suppressEdit(e *analysis.Engine, path string, text []byte, findings []diagnostic.Finding, data actionData, opt syntax.Options) (diagnostic.TextEdit, bool) {
	before := 0
	found := false
	for _, f := range findings {
		if f.Rule == data.Rule {
			before++
			found = found || (f.Span.Start == data.Start && f.Span.End == data.End)
		}
	}
	if !found {
		return diagnostic.TextEdit{}, false
	}
	edit, ok := analysis.SuppressEdit(syntax.ParseBest(path, text, opt), data.Rule, syntax.Span{Start: data.Start, End: data.End})
	if !ok {
		return diagnostic.TextEdit{}, false
	}
	edited := make([]byte, 0, len(text)+len(edit.NewText))
	edited = append(append(append(edited, text[:edit.Span.Start]...), edit.NewText...), text[edit.Span.Start:]...)
	after := 0
	for _, f := range e.Analyze(syntax.ParseBest(path, edited, opt)) {
		if f.Rule == data.Rule {
			after++
		}
	}
	return edit, after == before-1
}
