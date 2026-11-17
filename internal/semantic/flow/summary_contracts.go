package flow

import (
	"strings"

	"custos/internal/semantic/names"
)

// begin starts recording the contracts a result computed while summaries
// are made reads; nil when they are not being made.
func (s *summaryResolver) begin() *readSet {
	if s.consulted == nil {
		return nil
	}
	reads := &readSet{}
	s.reading = append(s.reading, reads)
	return reads
}

// end stops recording for reads, which the result computing it, if any,
// read as well.
func (s *summaryResolver) end(reads *readSet) {
	if reads == nil {
		return
	}
	s.reading = s.reading[:len(s.reading)-1]
	s.depend(reads)
}

// depend records that the result being computed read one computed from
// reads.
func (s *summaryResolver) depend(reads *readSet) {
	if len(s.reading) > 0 {
		s.reading[len(s.reading)-1].addAll(reads)
	}
}

// contract returns the summary of a project callable, recording the name
// and the answer while the summaries of the file are being computed.
func (e *Env) contract(name string) (Summary, bool) {
	if e.consulted == nil || name == "" {
		if e.local != nil && name != "" {
			if summary, known, own := e.localContract(names.Lower(name)); own {
				return summary, known
			}
		}
		return e.snapshot.lookup(name)
	}
	e.consulted[name] = true
	key := names.Lower(name)
	id := e.keyTable().id(key)
	if len(e.reading) > 0 {
		e.reading[len(e.reading)-1].add(id)
	}
	summary, ok, own := e.serve(key, id)
	if !own {
		summary, ok = e.snapshot.lookup(key)
	}
	if state := e.summarizing; state != nil && len(state.stack) > 0 {
		state.answered(key, own)
	}
	if state := e.summarizing; state != nil && state.dependencies != nil {
		// Local workers publish into one immutable view. Stable own and external
		// answers cannot change; retain only local recursion holes for replay.
		if !own && len(state.declared[key]) == 1 && state.alone(id, key) {
			*e.answer(id) = answer{known: true, held: true}
		}
		return summary, ok
	}
	at := e.answer(id)
	read := *at
	if !read.known {
		read = answer{known: true, ok: ok}
		if e.answered != nil {
			e.answered[key] = summary
		}
	}
	if own {
		read.own = true
	} else {
		read.held = true
	}
	*at = read
	return summary, ok
}

// answer returns the entry of e.answers for the key numbered id.
func (s *summaryResolver) answer(id uint32) *answer {
	if int(id) >= len(s.answers) {
		s.answers = append(s.answers, make([]answer, int(id)+1-len(s.answers))...)
	}
	return &s.answers[id]
}

// answer is what a contract read while summaries were computed gave: a
// summary or none (ok), the file's own (own), the snapshot's (held), or
// each at some time. The summary itself is kept apart (Env.answered).
type answer struct {
	// known tells that the contract was read at all.
	known     bool
	ok        bool
	own, held bool
}

// serve returns the summary of the file being summarized for a key the
// snapshot with the file's summaries would answer from them: a class of the
// file, or a callable of it declared once, summarized now unless its
// summary is being made. own is false for any other key.
func (e *Env) serve(key string, id uint32) (summary Summary, ok, own bool) {
	s := e.summarizing
	if s == nil || !s.alone(id, key) {
		return Summary{}, false, false
	}
	if strings.HasPrefix(key, sinkClassKey("")) {
		summary, ok = s.out.Summaries[key]
		return summary, ok, ok
	}
	if declared := s.declared[key]; len(declared) == 1 && s.state[declared[0]] != summaryBegun {
		// The reader depends on the summary, which Rebase compares, not on
		// what computing it read.
		if cached, known := s.cached[key]; known {
			return cached, true, true
		}
		reading := e.reading
		e.reading = nil
		e.summarize(declared[0])
		e.reading = reading
		return s.out.Summaries[key], true, true
	}
	return Summary{}, false, false
}

// alone is s.base.alone for the file being summarized and the key
// numbered id: base is set before any contract is read and never changes.
func (s *summarizing) alone(id uint32, key string) bool {
	return s.aloneKeys.holds(id, s.base, s.out.Path, key)
}

// aloneKeys remembers, per contract key number, whether a snapshot that
// never changes declares the key in no file but one: 2 when it does, 1 when
// it does not.
type aloneKeys []uint8

func (c *aloneKeys) holds(id uint32, base *Snapshot, path, key string) bool {
	if int(id) >= len(*c) {
		*c = append(*c, make([]uint8, int(id)+1-len(*c))...)
	}
	known := (*c)[id]
	if known == 0 {
		known = 1
		if base.alone(path, key) {
			known = 2
		}
		(*c)[id] = known
	}
	return known == 2
}
