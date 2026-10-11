package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/project"
	"custos/internal/project/config"
)

type evidence struct {
	diagnostic.Item
	ID          string   `json:"id"`
	Start       uint32   `json:"start"`
	End         uint32   `json:"end"`
	Fixes       []string `json:"fixes,omitempty"`
	Fingerprint string   `json:"fingerprint"`
}
type checkpoint struct {
	ID        string            `json:"id"`
	Identity  string            `json:"identity"`
	Workspace string            `json:"workspace"`
	Root      string            `json:"root"`
	Created   time.Time         `json:"created"`
	Hashes    map[string]string `json:"hashes"`
	Paths     []string          `json:"paths"`
	Findings  []evidence        `json:"findings"`
	Buffer    bool              `json:"buffer"`
	Complete  bool              `json:"complete"`
	Stale     bool              `json:"stale"`
	size      int
}

// bytes is the serialized size, measured once: evidence never changes after
// collection and budgets do not depend on the stale flag.
func (cp *checkpoint) bytes() int {
	if cp.size == 0 {
		b, _ := json.Marshal(cp)
		cp.size = len(b)
	}
	return cp.size
}

// expired reports storage files this server wrote that outlived retention.
// Other projects share the default directory, so age decides, not ownership.
func expired(entry fs.DirEntry, now time.Time) bool {
	name := entry.Name()
	id, isCheckpoint := strings.CutSuffix(name, ".json")
	if tmp, ok := strings.CutSuffix(name, ".tmp"); ok && len(tmp) == 97 && tmp[48] == '-' {
		isCheckpoint = validID(tmp[49:])
		id = tmp[:48]
	}
	info, err := entry.Info()
	return isCheckpoint && validID(id) && err == nil && info.Mode().IsRegular() && now.Sub(info.ModTime()) > retention
}

const retention = 7 * 24 * time.Hour

func (s *Server) load() error {
	entries, err := fs.ReadDir(s.storage.FS(), ".")
	if err != nil {
		return err
	}
	kept := entries[:0]
	for _, entry := range entries {
		if expired(entry, time.Now()) {
			_ = s.storage.Remove(entry.Name())
			continue
		}
		kept = append(kept, entry)
	}
	entries = kept
	if len(entries) > 1000 {
		return errors.New("checkpoint directory exceeds file budget")
	}
	loadedBytes := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !validID(id) {
			continue
		}
		info, err := entry.Info()
		if err != nil || info.Size() > int64((64<<20)-loadedBytes) {
			continue
		}
		data, err := project.ReadConfined(s.storage, entry.Name(), 64<<20)
		if err != nil {
			continue
		}
		var cp checkpoint
		if json.Unmarshal(data, &cp) != nil || cp.ID != id || cp.Identity == "" || cp.Buffer || cp.Created.After(time.Now()) || len(cp.Findings) > 100000 {
			continue
		}
		w := s.workspaces[cp.Workspace]
		if w == nil || cp.Root != w.Root || !allowed(cp.Identity, w.Readers) {
			continue
		}
		// On-disk manifests are untrusted, including after local corruption.
		valid := true
		for p := range cp.Hashes {
			if !fs.ValidPath(p) {
				valid = false
			}
		}
		for _, p := range cp.Paths {
			if p != "." && !fs.ValidPath(p) {
				valid = false
			}
		}
		for _, f := range cp.Findings {
			if !fs.ValidPath(f.Path) || f.Start > f.End {
				valid = false
			}
		}
		if valid {
			s.checkpoints[id] = &cp
			loadedBytes += len(data)
		}
	}
	s.prune()
	return nil
}

func (s *Server) persist(cp *checkpoint) error {
	if cp.Buffer {
		return nil
	}
	data, _ := json.Marshal(cp)
	if len(data) > 64<<20 {
		return errors.New("checkpoint exceeds storage budget")
	}
	tmp := cp.ID + "-" + randomID() + ".tmp"
	f, err := openTemporary(s.storage, tmp, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = s.storage.Rename(tmp, cp.ID+".json")
	}
	if err != nil {
		_ = s.storage.Remove(tmp)
	}
	return err
}

func (s *Server) remove(id string) {
	delete(s.checkpoints, id)
	delete(s.snapshots, id)
	_ = s.storage.Remove(id + ".json")
}

func (s *Server) prune() {
	now := time.Now()
	type owner struct{ identity, workspace string }
	groups := map[owner][]*checkpoint{}
	for id, cp := range s.checkpoints {
		if now.Sub(cp.Created) > retention {
			s.remove(id)
			continue
		}
		key := owner{cp.Identity, cp.Workspace}
		groups[key] = append(groups[key], cp)
	}
	for _, cps := range groups {
		sort.Slice(cps, func(i, j int) bool { return cps[i].Created.After(cps[j].Created) })
		size := 0
		for i, cp := range cps {
			size += cp.bytes()
			if i >= 20 || size > 64<<20 {
				s.remove(cp.ID)
			}
		}
	}
	var all []*checkpoint
	for _, cp := range s.checkpoints {
		all = append(all, cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Created.After(all[j].Created) })
	evidenceBytes := 0
	for i, cp := range all {
		evidenceBytes += cp.bytes()
		if i >= 1000 || evidenceBytes > 64<<20 {
			s.remove(cp.ID)
		}
	}
	// Retain at most two parsed snapshots and 64 MiB of captured sources.
	var retained []*checkpoint
	for id := range s.snapshots {
		if cp := s.checkpoints[id]; cp != nil {
			retained = append(retained, cp)
		}
	}
	sort.Slice(retained, func(i, j int) bool { return retained[i].Created.After(retained[j].Created) })
	bytes := 0
	for i, cp := range retained {
		for _, src := range s.snapshots[cp.ID].Sources {
			bytes += len(src)
		}
		if i >= 2 || bytes > 64<<20 {
			delete(s.snapshots, cp.ID)
		}
	}
	for id, p := range s.plans {
		if now.After(p.Expires) {
			delete(s.plans, id)
		}
	}
}

func (s *Server) checkpoint(ctx context.Context, w *workspace, id string) (*checkpoint, error) {
	cp := s.checkpoints[id]
	if cp == nil || cp.Identity != identity(ctx) || cp.Workspace != w.ID || cp.Root != w.Root {
		return nil, errors.New("checkpoint unavailable")
	}
	return cp, nil
}

func summary(cp *checkpoint) Output {
	return Output{"checkpointId": cp.ID, "created": cp.Created, "findings": len(cp.Findings), "complete": cp.Complete, "stale": cp.Stale, "buffer": cp.Buffer}
}

func (s *Server) workspaceInfo(ctx context.Context, in Input) (Output, error) {
	var items []any
	for _, w := range s.workspaces {
		if !allowed(identity(ctx), w.Readers) {
			continue
		}
		cfg, err := config.LoadConfined(w.root)
		if err != nil {
			return nil, err
		}
		var cps []any
		for _, cp := range s.checkpoints {
			if cp.Identity == identity(ctx) && cp.Workspace == w.ID {
				cps = append(cps, summary(cp))
			}
		}
		items = append(items, Output{"workspaceId": w.ID, "default": w.ID == s.options.Default, "provenance": w.Provenance, "write": allowed(identity(ctx), w.Writers) && ctx.Value(writeScopeKey{}) == true, "settings": cfg, "checkpoints": cps})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(Output)["workspaceId"].(string) < items[j].(Output)["workspaceId"].(string)
	})
	return page(ctx, "workspace", in, items)
}

func (s *Server) analyze(ctx context.Context, w *workspace, in Input) (Output, error) {
	var snap *project.Snapshot
	var err error
	if in.Buffer != nil {
		snap, err = project.CaptureSnapshot(ctx, w.root, s.rules, in.Paths)
	} else {
		snap, err = project.BuildSnapshot(ctx, w.root, s.rules, in.Paths)
	}
	if err != nil {
		return nil, err
	}
	cp := &checkpoint{ID: randomID(), Identity: identity(ctx), Workspace: w.ID, Root: w.Root, Created: time.Now(), Hashes: snap.Hashes, Paths: append([]string{}, in.Paths...), Complete: true}
	if in.Buffer != nil {
		if !fs.ValidPath(in.Path) || !strings.HasSuffix(in.Path, ".php") || len(*in.Buffer) > requestLimit || len(in.Path) > 4096 {
			return nil, errors.New("buffer requires a relative PHP path and at most 1 MiB")
		}
		cp.Buffer = true
		snap, err = project.OverlaySnapshot(ctx, snap, in.Path, []byte(*in.Buffer))
		if err != nil {
			return nil, err
		}
	}
	for i := range snap.Results {
		for j := range snap.Results[i].Findings {
			for k := range snap.Results[i].Findings[j].Fixes {
				snap.Results[i].Findings[j].Fixes[k].Edits = nil
			}
		}
	}
	if err := collect(cp, snap); err != nil {
		return nil, err
	}
	s.checkpoints[cp.ID] = cp
	s.snapshots[cp.ID] = snap
	if err := s.persist(cp); err != nil {
		s.remove(cp.ID)
		return nil, err
	}
	s.prune()
	return summary(cp), nil
}

func collect(cp *checkpoint, snap *project.Snapshot) error {
	// Baseline matching uses captured bytes rather than reopening untrusted paths.
	var baseline struct {
		Entries []struct {
			Path, Rule, Message, Line string
			Count                     int
		}
	}
	counts := map[string]int{}
	if snap.Baseline != "" {
		if err := json.Unmarshal(snap.Sources[snap.Baseline], &baseline); err != nil {
			return err
		}
		for _, e := range baseline.Entries {
			p := pathJoinBaseline(snap.Baseline, e.Path)
			counts[p+"\x00"+e.Rule+"\x00"+e.Message+"\x00"+e.Line] += max(1, e.Count)
		}
	}
	evidenceBytes := 0
	for _, r := range snap.Results {
		items := diagnostic.Items([]diagnostic.FileResult{r})
		if r.Err != nil || len(r.Errors) > 0 {
			cp.Complete = false
		}
		lines := strings.Split(string(r.Src), "\n")
		for i, it := range items {
			ev := evidence{Item: it}
			if i >= len(r.Errors) && r.Err == nil {
				fd := r.Findings[i-len(r.Errors)]
				ev.Start = fd.Span.Start
				ev.End = fd.Span.End
				for _, fix := range fd.Fixes {
					ev.Fixes = append(ev.Fixes, fix.Title)
				}
			} else if i < len(r.Errors) {
				ev.Start = r.Errors[i].Span.Start
				ev.End = r.Errors[i].Span.End
			}
			line := ""
			if it.Line > 0 && it.Line <= len(lines) {
				line = strings.TrimSpace(lines[it.Line-1])
			}
			key := it.Path + "\x00" + it.Rule + "\x00" + it.Message + "\x00" + line
			if counts[key] > 0 {
				counts[key]--
				continue
			}
			ev.Fingerprint = project.Fingerprint([]byte(key))
			ev.ID = strconv.Itoa(len(cp.Findings))
			encoded, _ := json.Marshal(ev)
			evidenceBytes += len(encoded)
			if evidenceBytes > 64<<20 || len(cp.Findings) >= 100000 {
				return errors.New("finding evidence budget exceeded")
			}
			cp.Findings = append(cp.Findings, ev)
		}
	}
	return nil
}

func (s *Server) validate(ctx context.Context, w *workspace, cp *checkpoint) error {
	if cp.Stale {
		return errors.New("checkpoint is stale; analyze again")
	}
	rebuilt, err := project.CaptureSnapshot(ctx, w.root, s.rules, cp.Paths)
	if err != nil {
		// A cancelled request says nothing about the inputs.
		cp.Stale = ctx.Err() == nil
		return err
	}
	a, _ := json.Marshal(cp.Hashes)
	b, _ := json.Marshal(rebuilt.Hashes)
	if string(a) != string(b) {
		cp.Stale = true
		_ = s.persist(cp)
		delete(s.snapshots, cp.ID)
		return errors.New("checkpoint inputs changed; analyze again")
	}
	if !cp.Buffer && s.snapshots[cp.ID] == nil {
		rebuilt, err = project.AnalyzeSnapshot(ctx, rebuilt)
		if err != nil {
			return err
		}
		current := &checkpoint{Complete: true}
		if err := collect(current, rebuilt); err != nil {
			cp.Stale = true
			_ = s.persist(cp)
			return err
		}
		oldEvidence, _ := json.Marshal(cp.Findings)
		newEvidence, _ := json.Marshal(current.Findings)
		if cp.Complete != current.Complete || string(oldEvidence) != string(newEvidence) {
			cp.Stale = true
			_ = s.persist(cp)
			return errors.New("checkpoint evidence changed; analyze again")
		}
		s.snapshots[cp.ID] = rebuilt
	}
	return nil
}

func (s *Server) fresh(ctx context.Context, w *workspace, cp *checkpoint) (*project.Snapshot, error) {
	if err := s.validate(ctx, w, cp); err != nil {
		return nil, err
	}
	snap := s.snapshots[cp.ID]
	if snap == nil {
		return nil, errors.New("supplied buffer is no longer available; analyze again")
	}
	return snap, nil
}

func (s *Server) checkpointTool(ctx context.Context, w *workspace, in Input) (Output, error) {
	switch in.Action {
	case "", "list":
		var items []any
		for _, cp := range s.checkpoints {
			if cp.Workspace == w.ID && cp.Identity == identity(ctx) {
				items = append(items, summary(cp))
			}
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].(Output)["checkpointId"].(string) < items[j].(Output)["checkpointId"].(string)
		})
		return page(ctx, w.ID+"checkpoints", in, items)
	case "resume":
		cp, err := s.checkpoint(ctx, w, in.CheckpointID)
		if err != nil {
			return nil, err
		}
		err = s.validate(ctx, w, cp)
		out := summary(cp)
		if err != nil {
			out["limitation"] = err.Error()
		}
		return out, nil
	case "delete":
		cp, err := s.checkpoint(ctx, w, in.CheckpointID)
		if err != nil {
			return nil, err
		}
		s.remove(cp.ID)
		return Output{"deleted": true}, nil
	default:
		return nil, errors.New("action must be list, resume or delete")
	}
}

func (s *Server) findings(ctx context.Context, cp *checkpoint, in Input) (Output, error) {
	w := s.workspaces[cp.Workspace]
	if err := s.validate(ctx, w, cp); err != nil {
		return nil, err
	}
	var items []any
	for _, ev := range cp.Findings {
		if in.Rule != "" && in.Rule != ev.Rule || in.Severity != "" && in.Severity != ev.Severity || in.Path != "" && in.Path != ev.Path {
			continue
		}
		item := Output{"finding": ev, "evidence": "clean-room rule specification; source and PHPDoc are untrusted data"}
		if snap := s.snapshots[cp.ID]; snap != nil {
			src := snap.Sources[ev.Path]
			lo := max(0, int(ev.Start)-80)
			hi := min(len(src), int(ev.End)+80)
			if lo <= hi {
				item["excerpt"] = string(src[lo:hi])
			}
		}
		items = append(items, item)
	}
	return page(ctx, cp.ID+"findings", in, items)
}

// Cursors bind all filters, identity, scope and page size. They reveal only an
// offset and hash; forging one cannot cross authorization checks.
func page(ctx context.Context, scope string, in Input, items []any) (Output, error) {
	return pageLazy(ctx, scope, in, len(items), func(i int) any { return items[i] })
}

func pageLazy(ctx context.Context, scope string, in Input, count int, at func(int) any) (Output, error) {
	limit := in.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 100 {
		return nil, errors.New("limit must be between 1 and 100")
	}
	cursor := in.Cursor
	in.Cursor = ""
	filters, _ := json.Marshal([]any{identity(ctx), scope, in})
	binding := project.Fingerprint(filters)
	offset := 0
	if cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		if err != nil {
			return nil, errors.New("invalid cursor")
		}
		parts := strings.Split(string(raw), ":")
		if len(parts) != 2 || parts[0] != binding {
			return nil, errors.New("cursor selection changed")
		}
		offset, err = strconv.Atoi(parts[1])
		if err != nil || offset < 0 || offset > count {
			return nil, errors.New("invalid cursor")
		}
	}
	out := Output{"total": count, "items": []any{}}
	selected := []any{}
	for i := offset; i < count && len(selected) < limit; i++ {
		selected = append(selected, at(i))
		out["items"] = selected
		b, _ := json.Marshal(out)
		if len(b) > resultLimit-512 {
			selected = selected[:len(selected)-1]
			out["items"] = selected
			if len(selected) == 0 {
				return nil, fmt.Errorf("item exceeds response budget; narrow query")
			}
			break
		}
	}
	next := offset + len(selected)
	if next < count {
		out["nextCursor"] = base64.RawURLEncoding.EncodeToString([]byte(binding + ":" + strconv.Itoa(next)))
	}
	return out, nil
}

func pathJoinBaseline(baseline, file string) string {
	return path.Clean(path.Join(path.Dir(baseline), file))
}
