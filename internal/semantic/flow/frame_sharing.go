package flow

// Sharing markers are monotonic and never recycled while a frame holds one.
// A private environment and its synchronous summary workers share the arena;
// its markers contain no semantic facts and never enter published summaries.
type sharingFlags struct {
	chunk []bool
	next  int
}

func (a *sharingFlags) allocate() *bool {
	if a.next == len(a.chunk) {
		a.chunk = make([]bool, 128)
		a.next = 0
	}
	flag := &a.chunk[a.next]
	a.next++
	return flag
}

func (e *Env) sharingArena() *sharingFlags {
	if e.sharing == nil {
		e.sharing = &sharingFlags{}
	}
	return e.sharing
}

func (f *frame) shareFlag() *bool {
	if f.owner == nil {
		return new(bool)
	}
	if f.owner.sharing == nil {
		f.owner.sharing = &sharingFlags{}
	}
	return f.owner.sharing.allocate()
}
