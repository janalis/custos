package flow

import (
	"strings"
)

// begin starts recording the contracts a result computed while summaries
// are made reads; nil when they are not being made.
func (s *summaryResolver) begin() map[string]bool {
	if s.consulted == nil {
		return nil
	}
	reads := map[string]bool{}
	s.reading = append(s.reading, reads)
	return reads
}

// end stops recording for reads, which the result computing it, if any,
// read as well.
func (s *summaryResolver) end(reads map[string]bool) {
	if reads == nil {
		return
	}
	s.reading = s.reading[:len(s.reading)-1]
	s.depend(reads)
}

// depend records that the result being computed read one computed from
// reads.
func (s *summaryResolver) depend(reads map[string]bool) {
	if len(s.reading) > 0 {
		addReads(s.reading[len(s.reading)-1], reads)
	}
}

func readsAny(reads, keys map[string]bool) bool {
	for key := range reads {
		if keys[key] {
			return true
		}
	}
	return false
}

func addReads(into, from map[string]bool) {
	for key := range from {
		into[key] = true
	}
}

// contract returns the summary of a project callable, recording the name
// and the answer while the summaries of the file are being computed.
func (e *Env) contract(name string) (Summary, bool) {
	if e.consulted == nil || name == "" {
		if e.local != nil && name != "" {
			if summary, known, own := e.localContract(strings.ToLower(name)); own {
				return summary, known
			}
		}
		return e.snapshot.lookup(name)
	}
	e.consulted[name] = true
	key := strings.ToLower(name)
	if len(e.reading) > 0 {
		e.reading[len(e.reading)-1][key] = true
	}
	summary, ok, own := e.serve(key)
	if !own {
		summary, ok = e.snapshot.lookup(key)
	}
	if state := e.summarizing; state != nil && len(state.stack) > 0 {
		state.answered(key, own)
	}
	if state := e.summarizing; state != nil && state.dependencies != nil {
		// Local workers publish into one immutable view. Stable own and external
		// answers cannot change; retain only local recursion holes for replay.
		if !own && len(state.declared[key]) == 1 && state.base.alone(e.file.Path, key) {
			if e.answers == nil {
				e.answers = map[string]answer{}
			}
			e.answers[key] = answer{held: true}
		}
		return summary, ok
	}
	read, again := e.answers[key]
	if !again {
		if e.answers == nil {
			e.answers = map[string]answer{}
		}
		read = answer{summary: summary, ok: ok}
	}
	if own {
		read.own = true
	} else {
		read.held = true
	}
	e.answers[key] = read
	return summary, ok
}

// answer is what a contract read while summaries were computed gave: the
// file's own summary (own), the snapshot's (held), or each at some time.
type answer struct {
	summary   Summary
	ok        bool
	own, held bool
}

// serve returns the summary of the file being summarized for a key the
// snapshot with the file's summaries would answer from them: a class of the
// file, or a callable of it declared once, summarized now unless its
// summary is being made. own is false for any other key.
func (e *Env) serve(key string) (summary Summary, ok, own bool) {
	s := e.summarizing
	if s == nil || !s.base.alone(s.out.Path, key) {
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
