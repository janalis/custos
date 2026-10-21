package mcp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/output/diff"
	"custos/internal/php/syntax"
	"custos/internal/project"
)

type Choice struct {
	FindingID string `json:"findingId"`
	Fix       int    `json:"fix"`
}
type change struct {
	Path           string
	Source, Output []byte
	Perm           os.FileMode
	Edits          []diagnostic.TextEdit
}
type plan struct {
	ID, Identity, Workspace, Checkpoint string
	Expires                             time.Time
	Changes                             []change
	Preview                             bool
}

func (s *Server) prepare(ctx context.Context, w *workspace, cp *checkpoint, snap *project.Snapshot, in Input) (Output, error) {
	rule, err := s.checkPrepare(ctx, w, cp, in)
	if err != nil {
		return nil, err
	}
	if in.PlanID != "" {
		return s.planPage(ctx, s.plans[in.PlanID], in)
	}
	selected := map[string][]diagnostic.TextEdit{}
	parse := syntax.Options{Version: snap.Config.PHP, ShortOpenTag: snap.Config.ShortOpenTag}
	for _, choice := range in.Choices {
		ev := cp.finding(choice.FindingID)
		result := project.AnalyzeBuffer(snap.Engine, ev.Path, snap.Sources[ev.Path], parse)
		found := false
		for _, fd := range result.Findings {
			if fd.Rule != ev.Rule || fd.Span.Start != ev.Start || fd.Span.End != ev.End || choice.Fix >= len(fd.Fixes) {
				continue
			}
			edits, ok := fixing.Edits(fd.Fixes[choice.Fix])
			if !ok {
				return nil, errors.New("fix failed to compute its edits")
			}
			selected[ev.Path] = append(selected[ev.Path], edits...)
			found = true
			break
		}
		if !found {
			return nil, errors.New("finding no longer reproducible")
		}
	}
	p := &plan{ID: randomID(), Identity: identity(ctx), Workspace: w.ID, Checkpoint: cp.ID, Expires: time.Now().Add(10 * time.Minute), Preview: cp.Buffer}
	files := snap.Files
	if cp.Buffer {
		files = nil
		for _, r := range snap.Results {
			files = append(files, r.Path)
		}
	}
	var names []string
	for _, name := range files {
		if in.Path == "" || in.Path == name {
			names = append(names, name)
		}
	}
	fixed := make([][]byte, len(names))
	if in.Iterative {
		filter := func(fd diagnostic.Finding) bool { return rule == "" || fd.Rule == rule }
		if err := project.Each(ctx, len(names), func(i int) {
			fixed[i] = fixing.FixSourceContext(ctx, snap.Engine, names[i], snap.Sources[names[i]], fixing.Options{Parse: parse, Filter: filter}).Source
		}); err != nil {
			return nil, err
		}
	}
	for i, name := range names {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		src := snap.Sources[name]
		var out []byte
		edits := selected[name]
		if in.Iterative {
			out = fixed[i]
			// The loop may rewrite the same region several times: describe
			// its net effect as minimal line replacements.
			for _, e := range diff.Edits(string(src), string(out)) {
				edits = append(edits, diagnostic.TextEdit{Span: syntax.Span{Start: uint32(e.Start), End: uint32(e.End)}, NewText: e.Text})
			}
		} else {
			if err := validateEdits(src, edits); err != nil {
				return nil, err
			}
			out, _ = fixing.Apply(src, edits)
		}
		if string(src) == string(out) {
			continue
		}
		if len(out) > syntax.MaxFileSize || len(syntax.ParseBest(name, out, parse).Errors) > 0 {
			return nil, errors.New("fix output is not parsable")
		}
		perm := os.FileMode(0o600)
		if !cp.Buffer {
			st, err := w.root.Lstat(name)
			if err != nil {
				return nil, err
			}
			if !st.Mode().IsRegular() {
				return nil, errors.New("fix input must be a regular file")
			}
			perm = st.Mode().Perm()
		}
		p.Changes = append(p.Changes, change{Path: name, Source: src, Output: out, Perm: perm, Edits: edits})
	}
	total := 0
	pathBytes := 0
	for _, c := range p.Changes {
		pathBytes += len(c.Path)
	}
	if pathBytes > 8192 {
		return nil, errors.New("plan path budget exceeded; select fewer files")
	}
	for _, existing := range s.plans {
		for _, c := range existing.Changes {
			total += len(c.Source) + len(c.Output)
		}
	}
	for _, c := range p.Changes {
		total += len(c.Source) + len(c.Output)
	}
	if len(p.Changes) > 100 || total > 64<<20 {
		return nil, errors.New("fix plan budget exceeded; select fewer files")
	}
	s.plans[p.ID] = p
	return s.planPage(ctx, p, in)
}

func (cp *checkpoint) finding(id string) *evidence {
	for i := range cp.Findings {
		if cp.Findings[i].ID == id {
			return &cp.Findings[i]
		}
	}
	return nil
}

// checkPrepare rejects a request from its arguments and the checkpoint alone,
// before any snapshot is validated or rebuilt. It returns the canonical rule.
func (s *Server) checkPrepare(ctx context.Context, w *workspace, cp *checkpoint, in Input) (string, error) {
	if in.PlanID != "" {
		p := s.plans[in.PlanID]
		if p == nil || p.Identity != identity(ctx) || p.Workspace != w.ID || p.Checkpoint != cp.ID {
			return "", errors.New("plan unavailable; prepare a new one")
		}
		return "", nil
	}
	if !in.Iterative && len(in.Choices) == 0 {
		return "", errors.New("select fix choices or set iterative to true")
	}
	if len(s.plans) >= 20 {
		return "", errors.New("plan capacity reached; apply a plan or wait for expiry")
	}
	for _, choice := range in.Choices {
		ev := cp.finding(choice.FindingID)
		if ev == nil {
			return "", fmt.Errorf("finding %q is not in this checkpoint", choice.FindingID)
		}
		if choice.Fix < 0 || choice.Fix >= len(ev.Fixes) {
			return "", fmt.Errorf("finding %s has %d fix choices", ev.ID, len(ev.Fixes))
		}
	}
	if in.Rule == "" {
		return "", nil
	}
	return canonicalRule(in.Rule)
}

func validateEdits(src []byte, edits []diagnostic.TextEdit) error {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].Span.Start != edits[j].Span.Start {
			return edits[i].Span.Start < edits[j].Span.Start
		}
		return edits[i].Span.End < edits[j].Span.End
	})
	for i, e := range edits {
		if e.Span.End < e.Span.Start || int(e.Span.End) > len(src) {
			return errors.New("invalid edit")
		}
		if i > 0 {
			prev := edits[i-1]
			if prev.Span.End > e.Span.Start || prev.Span.Start == e.Span.Start || (e.Span.Start == e.Span.End && prev.Span.End == e.Span.Start) {
				return errors.New("conflicting edits")
			}
		}
	}
	return nil
}

func (s *Server) planPage(ctx context.Context, p *plan, in Input) (Output, error) {
	type part struct{ file, edit, start, end int }
	var parts []part
	for file, c := range p.Changes {
		parts = append(parts, part{file: file, edit: -1})
		for edit, e := range c.Edits {
			valid := utf8.ValidString(e.NewText)
			for offset := 0; ; {
				end := min(len(e.NewText), offset+2048)
				if valid {
					for end < len(e.NewText) && end > offset && !utf8.RuneStart(e.NewText[end]) {
						end--
					}
				}
				parts = append(parts, part{file: file, edit: edit, start: offset, end: end})
				if end == len(e.NewText) {
					break
				}
				offset = end
			}
		}
	}
	selection := Input{WorkspaceID: p.Workspace, CheckpointID: p.Checkpoint, PlanID: p.ID, Limit: in.Limit, Cursor: in.Cursor}
	out, err := pageLazy(ctx, p.ID, selection, len(parts), func(i int) any {
		item := parts[i]
		c := p.Changes[item.file]
		if item.edit < 0 {
			preview := "Diff omitted for large input; paginated edit text contains the exact replacement."
			if len(c.Source)+len(c.Output) < 16384 && strings.Count(string(c.Source), "\n")+strings.Count(string(c.Output), "\n") < 1000 {
				preview = diff.Unified(c.Path, string(c.Source), string(c.Output), 3)
			}
			if len(preview) > 4096 {
				preview = preview[:4096] + "\nPreview truncated; exact edit chunks follow."
			}
			return Output{"path": c.Path, "diff": preview}
		}
		e := c.Edits[item.edit]
		chunk := e.NewText[item.start:item.end]
		return Output{"path": c.Path, "edit": item.edit, "start": e.Span.Start, "end": e.Span.End, "textOffset": item.start, "text": chunk, "textBase64": base64.StdEncoding.EncodeToString([]byte(chunk)), "textBytes": len(e.NewText)}
	})
	if err == nil {
		out["planId"] = p.ID
		out["expires"] = p.Expires
		out["previewOnly"] = p.Preview
	}
	return out, err
}

func (s *Server) apply(ctx context.Context, w *workspace, in Input) (Output, error) {
	p := s.plans[in.PlanID]
	if p == nil || p.Identity != identity(ctx) || p.Workspace != w.ID || p.Preview || time.Now().After(p.Expires) {
		return nil, errors.New("disk plan unavailable")
	}
	cp, err := s.checkpoint(ctx, w, p.Checkpoint)
	if err != nil {
		return nil, err
	}
	// Targets are compared byte for byte before each replacement, so a
	// missing snapshot is not analyzed again first.
	if err := s.validate(ctx, w, cp, false); err != nil {
		return nil, err
	}
	// Consume before the first write. A partially applied plan cannot replay.
	delete(s.plans, p.ID)
	var outcomes []any
	partial := false
	for _, c := range p.Changes {
		err := ctx.Err()
		if err == nil {
			err = replaceContext(ctx, w, c)
		}
		outcome := Output{"path": c.Path, "applied": err == nil}
		if err != nil {
			outcome["error"] = err.Error()
			partial = true
		}
		outcomes = append(outcomes, outcome)
	}
	// The written files supersede this checkpoint.
	s.remove(cp.ID)
	out := Output{"outcomes": outcomes, "partialFailure": partial}
	if ctx.Err() == nil {
		next, err := s.analyze(ctx, w, Input{Paths: cp.Paths})
		if err == nil {
			out["analysis"] = next
		} else {
			out["analysisError"] = err.Error()
		}
	}
	return out, nil
}

func replace(w *workspace, c change) error { return replaceContext(context.Background(), w, c) }

func replaceContext(ctx context.Context, w *workspace, c change) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Pin the parent directory so a concurrent symlink/rename swap cannot move
	// the replacement into a newly substituted parent outside the workspace.
	parent, err := w.root.OpenRoot(path.Dir(c.Path))
	if err != nil {
		return err
	}
	defer parent.Close()
	name := path.Base(c.Path)
	st, err := parent.Lstat(name)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() || st.Mode().Perm() != c.Perm {
		return errors.New("target type or permissions changed")
	}
	current, err := project.ReadConfined(parent, name, syntax.MaxFileSize)
	if err != nil {
		return err
	}
	if string(current) != string(c.Source) {
		return errors.New("target changed")
	}
	tmp := ".custos-mcp-" + randomID()
	f, err := openTemporary(parent, tmp, c.Perm)
	if err != nil {
		return err
	}
	defer func() { _ = parent.Remove(tmp) }()
	_, err = f.Write(c.Output)
	if err == nil {
		err = f.Chmod(c.Perm)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if beforeCommit != nil {
		beforeCommit(parent, name)
	}
	// Recheck after writing the temporary file to reduce accidental lost edits.
	current, err = project.ReadConfined(parent, name, syntax.MaxFileSize)
	if err != nil || string(current) != string(c.Source) {
		return fmt.Errorf("target changed during replacement")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return parent.Rename(tmp, name)
}

// I/O seams exercise failed writes and concurrent path substitution in tests.
var openTemporary = func(root *os.Root, name string, perm os.FileMode) (*os.File, error) {
	return root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
}
var beforeCommit func(*os.Root, string)
