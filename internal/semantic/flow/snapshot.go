package flow

import (
	"maps"
	"reflect"
	"slices"

	"custos/internal/php/syntax"
	"custos/internal/semantic/names"
)

// Sink describes an external operation reached inside a project wrapper.
type Sink struct {
	Name     string
	Context  Context
	Argument Value
	Path     string
	Span     syntax.Span
}

// Summary is a compact, syntax-free callable return and sink contract.
type Summary struct {
	pure              *pureContract
	defaults          []Value
	forward           *forwardReturn
	argumentSensitive bool
	returned          *returnedCallable
	resource          *resourceReturn
	effects           []resourceEffect
	// dispatch is a reserved, syntax-free class metadata contract.
	dispatch  *sinkClass
	Return    Value
	Sinks     []Sink
	Generator bool
	// Never marks a callable every path of which throws, exits or calls
	// another such callable: a call to it does not return.
	Never bool
	// ByRef holds, per by-reference parameter index, what the callable may
	// leave in the argument when it ends: request data it wrote, and a
	// parameter source for each argument whose value may still be there
	// (its own when some path does not overwrite it). Fixed tells that the
	// body is the one every call runs to its end: declared unconditionally,
	// outside a trait, and not a generator.
	ByRef map[int]Value
	Fixed bool
	// Quiet marks a callable whose body is the one every call runs and
	// which runs no code of the project but its own: no call that may, no
	// constructor, no callback, or a generator, which runs nothing until it
	// is iterated. A call to it cannot enter its caller again.
	Quiet bool
	// Spares marks a callable whose body is the one every call runs and
	// stores nothing in a request table: it names none and calls only what
	// PHP declares (see Env.Spares).
	Spares bool
}

// File contains summaries from one indexed file. It retains no source or AST.
type File struct {
	options   Options
	Path      string
	Summaries map[string]Summary
	// consulted names the callables whose contracts the summaries were
	// computed from: nothing else of a snapshot can change them.
	consulted map[string]bool
	// held lists the keys of those the snapshot answered (and not the
	// file's own summaries).
	held []string
	// served holds the keys of those only the file's own summaries
	// answered: a snapshot gives the same answers while the file alone
	// declares them.
	served map[string]bool
	// callables records, per callable key, the contracts its summary read
	// and how they were answered (see readAny), so that Resummarize can
	// keep the summaries no changed contract reaches.
	callables map[string]map[string]uint8
}

// Reads returns the summary keys (as File.Summaries lists them) of the
// contracts the snapshot supplied while the summaries were computed: within
// a recursion, the summaries depend on what the snapshot held for them.
func (f *File) Reads() []string { return f.held }

// Consulted returns the summary keys of every contract the summaries were
// computed from: those Reads lists and those the file's own summaries
// answered. Nothing else of a snapshot can change them.
func (f *File) Consulted() []string {
	keys := slices.Clone(f.held)
	for key := range f.served {
		keys = append(keys, key)
	}
	return keys
}

// Current reports that the summaries, extracted under the earlier snapshot,
// would come out the same under s: no contract they were computed from
// differs between the two. When s holds these very summaries, those the
// file served itself are unchanged while it alone declares them; one
// another file declares too compares as any other contract.
func (f *File) Current(earlier, s *Snapshot) bool {
	if f.options != snapshotOptions(earlier) || f.options != snapshotOptions(s) {
		return false
	}
	return len(f.differing(earlier, s, true)) == 0
}

// Basis returns the part of s that summaries extracted under s read: a
// snapshot that, for every contract f consulted, holds and looks up what s
// does, and so serves as earlier to Current and Reextract in place of s.
// It keeps far less than s alive.
func (f *File) Basis(s *Snapshot) *Snapshot {
	if s == nil || s.base != nil {
		return s
	}
	out := &Snapshot{options: s.options, summaries: make(map[string]*Summary, len(f.consulted))}
	for name := range f.consulted {
		key := names.Fold(name)
		if v, ok := s.summaries[key]; ok {
			out.summaries[key] = v
		} else if n := s.shared[key]; n > 0 {
			if out.shared == nil {
				out.shared = map[string]int{}
			}
			out.shared[key] = n
		}
	}
	return out
}

// Snapshot is an immutable project-local summary collection. WithFile and
// WithoutFile return copies; readers never observe mutations or need locks.
//
// Replacing one file is what every analysed file does to see its own buffer,
// so it does not rebuild the collection: the result overlays the file on its
// base and answers lookups from both.
type Snapshot struct {
	options Options
	files   map[string]*File
	// summaries points at the contracts so that a copy of the map (see
	// Replace) copies no Summary: values this large are stored out of line.
	summaries map[string]*Summary
	// shared counts the files declaring a name more than one file declares.
	// Such a name is ambiguous and has no entry in summaries.
	shared map[string]int
	// base and own make an overlay: base with the summaries of own's path
	// replaced by own. The maps above are then unused.
	base *Snapshot
	own  *File
}

// NewSnapshot creates an empty summary collection.
func NewSnapshot(files ...*File) *Snapshot {
	var options Options
	for _, f := range files {
		if f != nil {
			options = f.options
			break
		}
	}
	return NewSnapshotWithOptions(options, files...)
}

// NewSnapshotWithOptions preserves runtime assumptions even for empty snapshots.
func NewSnapshotWithOptions(options Options, files ...*File) *Snapshot {
	s := &Snapshot{options: options, files: map[string]*File{}, summaries: map[string]*Summary{}}
	for _, f := range files {
		if f != nil && f.options == options {
			s.files[f.Path] = f
		}
	}

	s.rebuild()
	return s
}

// WithFile replaces all summaries of a file in a new snapshot.
func (s *Snapshot) WithFile(f *File) *Snapshot {
	if s == nil || s.options != f.options {
		return NewSnapshotWithOptions(f.options, f)
	}
	return &Snapshot{options: s.options, base: s.flat(), own: f}
}

// Holds reports that the snapshot already has exactly these summaries for
// the file: overlaying them would change no lookup.
func (s *Snapshot) Holds(f *File) bool {
	if s == nil {
		return false
	}
	held := s.files[f.Path]
	if s.base != nil {
		if held = s.base.files[f.Path]; s.own.Path == f.Path {
			held = s.own
		}
	}
	return held != nil && held.options == f.options && s.options == f.options && reflect.DeepEqual(held.Summaries, f.Summaries)
}

// WithoutFile removes all summaries of a file in a new snapshot.
func (s *Snapshot) WithoutFile(path string) *Snapshot {
	n := NewSnapshotWithOptions(snapshotOptions(s))
	if s != nil {
		for p, f := range s.flat().files {
			if p != path {
				n.files[p] = f
			}
		}
	}
	n.rebuild()
	return n
}

// flat returns the snapshot with its own maps, merging an overlay.
func (s *Snapshot) flat() *Snapshot {
	if s.base == nil {
		return s
	}
	n := NewSnapshotWithOptions(s.options)
	for path, file := range s.base.files {
		n.files[path] = file
	}
	n.files[s.own.Path] = s.own
	n.rebuild()
	return n
}

// Replace returns the snapshot NewSnapshotWithOptions would build from its
// files with each file of prev replaced by the file of next at the same
// position (nil removes it, as a file with other options does). Only the
// names of these files are counted again: a fixed point that summarizes a
// few files per round does not re-add the contracts of every other file.
func (s *Snapshot) Replace(prev, next []*File) *Snapshot {
	if s == nil {
		s = NewSnapshotWithOptions(Options{})
	}
	f := s.flat()
	n := &Snapshot{options: f.options, files: maps.Clone(f.files), summaries: maps.Clone(f.summaries), shared: maps.Clone(f.shared)}
	// A name left to a single file whose summary is not at hand yet.
	pending := map[string]bool{}
	count := func(name string) int {
		if _, ok := n.summaries[name]; ok || pending[name] {
			return 1
		}
		return n.shared[name]
	}
	drop := func(old *File) {
		delete(n.files, old.Path)
		for name := range old.Summaries {
			switch c := count(name); c {
			case 1:
				delete(n.summaries, name)
				delete(pending, name)
			case 2:
				delete(n.shared, name)
				pending[name] = true
			default:
				n.shared[name] = c - 1
			}
		}
	}
	for _, old := range prev {
		if old != nil && n.files[old.Path] == old {
			drop(old)
		}
	}
	for _, file := range next {
		if file == nil {
			continue
		}
		if old := n.files[file.Path]; old != nil {
			drop(old)
		}
		if file.options != n.options {
			continue
		}
		n.files[file.Path] = file
		for name, summary := range file.Summaries {
			switch c := count(name); c {
			case 0:
				n.summaries[name] = &summary
			case 1:
				delete(n.summaries, name)
				delete(pending, name)
				if n.shared == nil {
					n.shared = map[string]int{}
				}
				n.shared[name] = 2
			default:
				n.shared[name] = c + 1
			}
		}
	}
	for name := range pending {
		for _, file := range n.files {
			if summary, ok := file.Summaries[name]; ok {
				n.summaries[name] = &summary
			}
		}
	}
	return n
}

func (s *Snapshot) rebuild() {
	for _, f := range s.files {
		for name, summary := range f.Summaries {
			if _, exists := s.summaries[name]; exists || s.shared[name] > 0 {
				if s.shared == nil {
					s.shared = map[string]int{}
				}
				if exists {
					delete(s.summaries, name)
					s.shared[name] = 1
				}
				s.shared[name]++
				continue
			}
			s.summaries[name] = &summary
		}
	}
}

// sameSummary compares two contracts through pointers, so that comparing
// them copies neither.
func sameSummary(a, b *Summary) bool { return reflect.DeepEqual(a, b) }

// held is the contract a snapshot that is no overlay holds for name itself,
// as lookup answers it; nil when lookup would answer otherwise.
func (s *Snapshot) held(name string) *Summary {
	if s == nil || s.base != nil {
		return nil
	}
	return s.summaries[names.Fold(name)]
}

func (s *Snapshot) lookup(name string) (Summary, bool) {
	if s == nil {
		return Summary{}, false
	}
	name = names.Fold(name)
	if s.base == nil {
		v, ok := s.summaries[name]
		if !ok {
			if s.shared[name] == 0 && s.options.Vendor != nil {
				return s.options.Vendor.lookup(name)
			}
			return Summary{}, false
		}
		return *v, true
	}
	// A name has a contract when exactly one file declares it, counting the
	// overlaid file instead of the one it replaces.
	replaced := s.base.files[s.own.Path]
	fresh, inOwn := s.own.Summaries[name]
	declaring := s.base.shared[name]
	if _, single := s.base.summaries[name]; single {
		declaring = 1
	}
	inReplaced := false
	if replaced != nil {
		_, inReplaced = replaced.Summaries[name]
	}
	if inReplaced {
		declaring--
	}
	switch {
	case inOwn && declaring == 0:
		return fresh, true
	case !inOwn && declaring == 0 && s.options.Vendor != nil:
		return s.options.Vendor.lookup(name)
	case inOwn || declaring != 1:
		return Summary{}, false
	case !inReplaced:
		return *s.base.summaries[name], true
	}
	// The replaced file shared the name with one other file, which now
	// declares it alone.
	var remaining Summary
	for path, file := range s.base.files {
		if v, ok := file.Summaries[name]; ok && path != s.own.Path {
			remaining = v
		}
	}
	return remaining, true
}

// Equal reports whether two immutable snapshots expose the same contracts.
func (s *Snapshot) Equal(other *Snapshot) bool {
	if s == nil || other == nil {
		return s == other
	}
	return s.options == other.options && reflect.DeepEqual(s.flat().summaries, other.flat().summaries)
}

// EqualFor is Equal for snapshots built from the same files but for those
// declaring the given names: only the contracts of these names are compared.
func (s *Snapshot) EqualFor(other *Snapshot, names []string) bool {
	if s == nil || other == nil {
		return s == other
	}
	if s.options != other.options {
		return false
	}
	a, b := s.flat(), other.flat()
	for _, name := range names {
		x, inA := a.summaries[name]
		y, inB := b.summaries[name]
		if inA != inB || !reflect.DeepEqual(x, y) {
			return false
		}
	}
	return true
}

// alone reports that no file of the snapshot but the one at path declares
// the callable or class key: with that file's summaries replaced by ones
// that declare it, the snapshot answers with them.
func (s *Snapshot) alone(path, key string) bool {
	switch {
	case s == nil:
		return true
	case s.base != nil:
		return false
	}
	declaring := s.shared[key]
	if _, single := s.summaries[key]; single {
		declaring = 1
	}
	if replaced := s.files[path]; replaced != nil {
		if _, held := replaced.Summaries[key]; held {
			declaring--
		}
	}
	return declaring == 0
}

// Options returns the analysis options the snapshot's summaries assume.
func (s *Snapshot) Options() Options { return snapshotOptions(s) }

func snapshotOptions(s *Snapshot) Options {
	if s == nil {
		return Options{}
	}
	return s.options
}
