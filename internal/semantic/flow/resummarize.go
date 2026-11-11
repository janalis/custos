package flow

import (
	"maps"
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// How a callable's summary read a contract (File.callables): at all, from
// the file's own summaries, or from the snapshot.
const (
	readAny uint8 = 1 << iota
	readOwn
	readHeld
)

// tracking records what the summary of the callable key reads until the
// returned function runs.
func (e *Env) tracking(s *summarizing, key string) func() {
	reads := e.begin()
	s.stack = append(s.stack, key)
	return func() {
		s.stack = s.stack[:len(s.stack)-1]
		e.end(reads)
		bits := s.track[key]
		if bits == nil {
			bits = map[string]uint8{}
			s.track[key] = bits
		}
		for read := range reads {
			bits[read] |= readAny
		}
	}
}

// answered records how the callable being summarized had key answered.
func (s *summarizing) answered(key string, own bool) {
	top := s.stack[len(s.stack)-1]
	bits := s.track[top]
	if bits == nil {
		bits = map[string]uint8{}
		s.track[top] = bits
	}
	if own {
		bits[key] |= readOwn
	} else {
		bits[key] |= readHeld
	}
}

// Reextract is Extract keeping, as Resummarize does, the summaries of prior
// that no contract differing between earlier and snapshot reaches.
func Reextract(file *syntax.File, ix *index.Index, php phpversion.Version, snapshot *Snapshot, options Options, prior *File, earlier *Snapshot) *File {
	_, out := Resummarize(file, infer.NewEnv(file, names.New(file), ix, php, infer.Options{Assertions: options.Assertions}), snapshot, options, prior, earlier)
	return out
}

// Resummarize is Summarize for a file whose summaries prior were computed,
// from the same source, against earlier: the summary of a callable that
// read no contract differing between earlier and snapshot, directly or
// through a callable of the file summarized again, is kept as it was.
func Resummarize(file *syntax.File, types *infer.Env, snapshot *Snapshot, options Options, prior *File, earlier *Snapshot) (*Env, *File) {
	e, order := summaryInventory(file, types, snapshot, options)
	s := e.summarizing
	s.track = map[string]map[string]uint8{}
	out := s.out
	kept := prior.unreached(earlier, snapshot, s)
	if len(kept) > 0 {
		s.cached = make(map[string]Summary, len(kept))
	}
	for key := range kept {
		s.cached[key] = prior.Summaries[key]
		out.Summaries[key] = prior.Summaries[key]
		s.track[key] = prior.callables[key]
		s.state[s.declared[key][0]] = summaryMade
	}
	for _, n := range order {
		e.summarize(n)
	}
	out.callables = s.track
	e.summarizing = nil
	consulted := e.consulted
	e.consulted = nil
	// The answers of the kept callables join those just read.
	answers := make(map[string]answer, len(e.answers))
	maps.Copy(answers, e.answers)
	e.answers = answers
	for key := range kept {
		for read, bits := range prior.callables[key] {
			consulted[read] = true
			if bits&(readOwn|readHeld) == 0 {
				continue
			}
			got := e.answers[read]
			got.held = got.held || bits&readHeld != 0
			got.own = got.own || bits&readOwn != 0
			e.answers[read] = got
		}
	}
	out.consulted = consulted
	for key, read := range e.answers {
		switch {
		case read.held:
			out.held = append(out.held, key)
		case out.served == nil:
			out.served = map[string]bool{key: true}
		default:
			out.served[key] = true
		}
	}
	slices.Sort(out.held)
	return e, out
}

// unreached returns the keys of the callables whose summaries f can give
// again under s: declared once, summarized in f, reading no contract that
// differs between earlier and s, nor the summary of a callable of the file
// that is summarized again.
func (f *File) unreached(earlier, s *Snapshot, state *summarizing) map[string]bool {
	if f == nil || f.callables == nil || f.options != state.out.options || f.Path != state.out.Path || snapshotOptions(earlier) != f.options {
		return nil
	}
	reached := f.changed(earlier, s)
	for key, declared := range state.declared {
		if _, made := f.callables[key]; !made || len(declared) != 1 {
			reached[key] = true
		}
	}
	for grown := true; grown; {
		grown = false
		for key, reads := range f.callables {
			if reached[key] {
				continue
			}
			for read := range reads {
				if reached[read] {
					reached[key], grown = true, true
					break
				}
			}
		}
	}
	kept := map[string]bool{}
	for key := range f.callables {
		if !reached[key] && len(state.declared[key]) == 1 {
			kept[key] = true
		}
	}
	return kept
}

// changed returns the keys of the contracts f consulted that differ
// between earlier and s, as Current compares them.
func (f *File) changed(earlier, s *Snapshot) map[string]bool {
	out := map[string]bool{}
	mine := s != nil && s.base == nil && s.files[f.Path] == f
	for name := range f.consulted {
		key := strings.ToLower(name)
		if mine && f.served[key] && s.alone(f.Path, key) {
			continue
		}
		if held := s.held(name); held != nil && held == earlier.held(name) {
			continue
		}
		before, had := earlier.lookup(name)
		now, has := s.lookup(name)
		if had != has || had && !sameSummary(&before, &now) {
			out[key] = true
		}
	}
	return out
}
