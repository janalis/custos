package flow

import (
	"reflect"
	"slices"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
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
}

// Reads returns the summary keys (as File.Summaries lists them) of the
// contracts the snapshot supplied while the summaries were computed: within
// a recursion, the summaries depend on what the snapshot held for them.
func (f *File) Reads() []string { return f.held }

// Current reports that the summaries, extracted under the earlier snapshot,
// would come out the same under s: no contract they were computed from
// differs between the two.
func (f *File) Current(earlier, s *Snapshot) bool {
	if f.options != snapshotOptions(earlier) || f.options != snapshotOptions(s) {
		return false
	}
	for name := range f.consulted {
		before, had := earlier.lookup(name)
		now, has := s.lookup(name)
		if had != has || !reflect.DeepEqual(before, now) {
			return false
		}
	}
	return true
}

// Snapshot is an immutable project-local summary collection. WithFile and
// WithoutFile return copies; readers never observe mutations or need locks.
//
// Replacing one file is what every analysed file does to see its own buffer,
// so it does not rebuild the collection: the result overlays the file on its
// base and answers lookups from both.
type Snapshot struct {
	options   Options
	files     map[string]*File
	summaries map[string]Summary
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
	s := &Snapshot{options: options, files: map[string]*File{}, summaries: map[string]Summary{}}
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
			s.summaries[name] = summary
		}
	}
}

func (s *Snapshot) lookup(name string) (Summary, bool) {
	if s == nil {
		return Summary{}, false
	}
	name = strings.ToLower(name)
	if s.base == nil {
		v, ok := s.summaries[name]
		return v, ok
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
	case inOwn || declaring != 1:
		return Summary{}, false
	case !inReplaced:
		return s.base.summaries[name], true
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

// Extract computes compact summaries while the project's parser tree is
// available. It does not open any other files. A supplied snapshot permits
// bounded composition with previously summarized project wrappers.
func Extract(file *syntax.File, ix *index.Index, php phpversion.Version, snapshot *Snapshot, options ...Options) *File {
	var resolved Options
	if len(options) > 0 {
		resolved = options[0]
	}
	_, out := Summarize(file, infer.NewEnv(file, names.New(file), ix, php, infer.Options{Assertions: resolved.Assertions}), snapshot, resolved)
	return out
}

// Summarize computes the summaries of a file and returns them with the
// environment that computed them, which answers for the file under the same
// snapshot without analysing its scopes again.
//
// A callable of the file that the snapshot, with these summaries in place of
// the file's, would answer for gets its summary from the file: one declared
// once is summarized before the first scope that reads it, unless that scope
// is part of its own summary, and the classes of the file are described
// first. Where no callable of the file reaches itself through the others,
// the environment then read what the snapshot with the file's summaries
// reads (see Rebase).
func Summarize(file *syntax.File, types *infer.Env, snapshot *Snapshot, options ...Options) (*Env, *File) {
	e := New(file, types, snapshot, options...)
	e.consulted = map[string]bool{}
	out := &File{options: e.options, Path: file.Path, Summaries: map[string]Summary{}}
	s := &summarizing{out: out, declared: map[string][]syntax.Node{}, names: map[syntax.Node]string{}, state: map[syntax.Node]uint8{}}
	var order []syntax.Node
	syntax.InspectFile(file, func(n syntax.Node) bool {
		name := ""
		switch n := n.(type) {
		case *syntax.ClassLike:
			name := types.ClassFQN(n)
			key := sinkClassKey(name)
			if _, duplicate := out.Summaries[key]; duplicate {
				out.Summaries[key] = Summary{}
			} else {
				out.Summaries[key] = Summary{dispatch: describeSinkClass(n, types)}
			}
		case *syntax.Function:
			if n.Name != nil {
				name, _ = types.Names.Function(n.Name.Value, n.Span().Start)
			}
		case *syntax.Method:
			if n.Name != nil {
				name = types.ClassFQN(syntax.EnclosingClass(n)) + "::" + n.Name.Value
			}
		}
		if name != "" {
			key := strings.ToLower(name)
			order = append(order, n)
			s.names[n] = key
			s.declared[key] = append(s.declared[key], n)
		}
		return true
	})
	e.summarizing = s
	for _, n := range order {
		e.summarize(n)
	}
	e.summarizing = nil
	out.consulted, e.consulted = e.consulted, nil
	for key, read := range e.answers {
		if read.held {
			out.held = append(out.held, key)
		}
	}
	slices.Sort(out.held)
	return e, out
}

// summarizing is the state of Summarize: the summaries made so far, each
// callable of the file under its key, and whether its summary is begun
// (summaryBegun) or made (summaryMade).
type summarizing struct {
	out      *File
	declared map[string][]syntax.Node
	names    map[syntax.Node]string
	state    map[syntax.Node]uint8
}

const (
	summaryBegun uint8 = iota + 1
	summaryMade
)

// summarize makes the summary of a callable of the file being summarized.
func (e *Env) summarize(n syntax.Node) {
	s := e.summarizing
	if s.state[n] != 0 {
		return
	}
	s.state[n] = summaryBegun
	r := e.scope(n)
	sum := Summary{Return: Value{Complete: r.complete}}
	body := syntax.VariableScopeBody(n)
	sum.Generator = r.generator
	for i, v := range r.returns {
		if i == 0 {
			sum.Return = v
		} else {
			sum.Return = merge(sum.Return, v)
		}
	}
	sum.Return.Complete = sum.Return.Complete && r.complete
	sum.Never = r.complete && r.leaves && !sum.Generator && body != nil && unconditional(n)
	sum.Fixed = !sum.Generator && unconditional(n)
	// The body of a generator runs when it is iterated, not when
	// it is called.
	sum.Quiet = unconditional(n) && body != nil && (sum.Generator || r.complete && !r.runs)
	sum.Spares = e.sparesTables(n, body)
	for i, p := range syntax.VariableScopeParams(n) {
		if p.ByRef && p.Var != nil && r.complete {
			// The argument holds what the name wrote to it while it
			// stood for it.
			left := r.exit[received(p.Var.Name)]
			if p.Variadic {
				// The arguments keep what they held, and may hold
				// what was written to an element of the array.
				left = merge(left, r.exit[p.Var.Name])
			}
			if sum.ByRef == nil {
				sum.ByRef = map[int]Value{}
			}
			sum.ByRef[i] = Value{Complete: left.Complete, Sources: left.Sources, Safe: left.Safe, scalars: left.scalars}
		}
	}
	sum.Return.Expr = nil
	sum.Return.interval = nil
	sum.Return.origins = nil
	sum.Return.Identity = 0
	if !sum.Generator {
		for _, c := range r.calls {
			sum.Sinks = append(sum.Sinks, e.nativeSinks(c)...)
			sum.Sinks = append(sum.Sinks, e.Sinks(c.Node)...)
		}
		sum.Sinks = append(sum.Sinks, e.commands(body, r)...)
	}
	s.out.Summaries[s.names[n]] = sum
	s.state[n] = summaryMade
}

// unconditional reports a callable whose body is the one every call of its
// name runs: a function declared at the top level of the file or of a
// namespace, or a method of a class declared there. A declaration inside a
// condition may not be the one in effect, and a trait method runs in each
// class that uses it, where the methods it calls can differ.
func unconditional(n syntax.Node) bool {
	if _, method := n.(*syntax.Method); method {
		class := syntax.EnclosingClass(n)
		if class == nil || class.ClassKind == syntax.KindTrait {
			return false
		}
		n = class
	}
	switch n.Parent().(type) {
	case nil, *syntax.File, *syntax.Namespace:
		return true
	}
	return false
}

// Rebase returns the environment for s that the one Summarize returned
// becomes: itself, reading any further contract from s, when every contract
// read meanwhile gave one answer, the one s gives. Otherwise a new one keeps
// the results computed from none of those that differ.
func (e *Env) Rebase(s *Snapshot) *Env {
	differ := map[string]bool{}
	for key, read := range e.answers {
		summary, ok := s.lookup(key)
		if read.own && read.held || read.ok != ok || !reflect.DeepEqual(read.summary, summary) {
			differ[key] = true
		}
	}
	if len(differ) == 0 && snapshotOptions(s) == e.options {
		e.snapshot = s
		return e
	}
	n := New(e.file, e.types, s, e.options)
	if snapshotOptions(s) != e.options {
		return n
	}
	for scope, r := range e.scopes {
		if r.reads != nil && !readsAny(r.reads, differ) {
			n.scopes[scope] = r
		}
	}
	return n
}

// begin starts recording the contracts a result computed while summaries
// are made reads; nil when they are not being made.
func (e *Env) begin() map[string]bool {
	if e.consulted == nil {
		return nil
	}
	reads := map[string]bool{}
	e.reading = append(e.reading, reads)
	return reads
}

// end stops recording for reads, which the result computing it, if any,
// read as well.
func (e *Env) end(reads map[string]bool) {
	if reads == nil {
		return
	}
	e.reading = e.reading[:len(e.reading)-1]
	e.depend(reads)
}

// depend records that the result being computed read one computed from
// reads.
func (e *Env) depend(reads map[string]bool) {
	if len(e.reading) > 0 {
		addReads(e.reading[len(e.reading)-1], reads)
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
	if s == nil || !e.snapshot.alone(s.out.Path, key) {
		return Summary{}, false, false
	}
	if strings.HasPrefix(key, sinkClassKey("")) {
		summary, ok = s.out.Summaries[key]
		return summary, ok, ok
	}
	if declared := s.declared[key]; len(declared) == 1 && s.state[declared[0]] != summaryBegun {
		// The reader depends on the summary, which Rebase compares, not on
		// what computing it read.
		reading := e.reading
		e.reading = nil
		e.summarize(declared[0])
		e.reading = reading
		return s.out.Summaries[key], true, true
	}
	return Summary{}, false, false
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

// nativeSinks recognizes resolved native sinks and binds named arguments.
func (e *Env) nativeSinks(c Call) []Sink {
	name, builtin := e.callName(c.Node)
	if !builtin {
		return nil
	}
	args, known := e.bind(c.Node, c.Arguments)
	if !known {
		// A reference argument, which a command fills with its output,
		// leaves the first parameter what it was given.
		args, known = e.leading(c)
	}
	if !known || len(args) == 0 {
		return nil
	}
	context := Context(0)
	position := 0
	switch c.Node.(type) {
	case *syntax.MethodCall:
		_, qualified, resolved := e.declaration(c.Node)
		if resolved && strings.HasPrefix(qualified, "pdo::") && (name == "query" || name == "exec" || name == "prepare") {
			context = SQL
		}
	case *syntax.FuncCall:
		switch name {
		case "system", "exec", "shell_exec", "passthru", "popen", "proc_open":
			context = Shell
		case "header":
			context = Header
		case "readfile", "unlink", "file_get_contents", "fopen", "file_put_contents", "rmdir":
			context = Path
		case "curl_init":
			context = URL
		case "curl_setopt":
			if len(args) > 2 {
				if constant, ok := args[1].Expr.(*syntax.ConstFetch); ok && strings.EqualFold(strings.TrimPrefix(constant.Name.Value, `\`), "CURLOPT_URL") {
					context = URL
					position = 2
				}
			}
		}
	}
	if context == 0 {
		return nil
	}
	if call, function := c.Node.(*syntax.FuncCall); function && name == "proc_open" && e.vector(call, args[0]) {
		// An array is an argument vector: no shell reads it.
		return nil
	}
	value := args[position]
	value.Expr = nil
	value.interval = nil
	value.origins = nil
	value.Identity = 0
	sinks := []Sink{{Name: name, Context: context, Argument: value, Path: e.file.Path, Span: c.Node.Span()}}
	if name == "header" {
		if binary, ok := args[0].Expr.(*syntax.Binary); ok && binary.Op.Kind == syntax.TDot {
			if prefix, known := literalText(binary.Left); known && strings.HasPrefix(strings.ToLower(prefix), "location:") {
				sinks = append(sinks, Sink{Name: name, Context: URL, Argument: value, Path: e.file.Path, Span: c.Node.Span()})
			}
		}
	}
	return sinks
}

// vector reports a `proc_open` command that is an array: an array literal, a
// variable holding one on every path, or an expression typed as an array.
func (e *Env) vector(call *syntax.FuncCall, command Value) bool {
	if _, literal := syntax.UnwrapParens(command.Expr).(*syntax.Array); literal {
		return true
	}
	var few [8]int
	var written syntax.Expr
	params, _, _ := e.declaration(call)
	for i, at := range parameters(call.Args, params, few[:0]) {
		if arg, ok := call.Args.Args[i].(*syntax.Arg); ok && at == 0 {
			written = arg.Value
		}
	}
	return e.types.TypeOf(written).IsArrayLike()
}

// leading returns the value given to the first parameter of a call whose
// arguments do not all bind by value.
func (e *Env) leading(c Call) ([]Value, bool) {
	call, ok := c.Node.(*syntax.FuncCall)
	params, _, known := e.declaration(c.Node)
	if !ok || !known || call.Args == nil {
		return nil, false
	}
	var few [8]int
	for i, at := range parameters(call.Args, params, few[:0]) {
		if arg, written := call.Args.Args[i].(*syntax.Arg); written && at == 0 && !arg.Unpack && !arg.ByRef {
			return c.Arguments[i : i+1], true
		}
	}
	return nil, false
}

// commands lists the backtick strings a body runs, outside the functions
// declared in it: each is a command given to the shell.
func (e *Env) commands(body syntax.Node, r *scopeResult) []Sink {
	var sinks []Sink
	syntax.Inspect(body, func(n syntax.Node) bool {
		if n != body && syntax.IsVariableScope(n) {
			return false
		}
		if command, ok := n.(*syntax.InterpolatedString); ok && command.Backtick {
			if value, reached := r.values[command]; reached {
				value.Expr, value.origins, value.Identity = nil, nil, 0
				sinks = append(sinks, Sink{Name: "shell_exec", Context: Shell, Argument: value, Path: e.file.Path, Span: command.Span()})
			}
		}
		return true
	})
	return sinks
}

// Sinks expands known project wrapper sink contracts at this call site.
// The argument sources and escaping contexts come from this call's flow.
func (e *Env) Sinks(at syntax.Node) []Sink {
	name := e.sinkSummaryName(at)
	summary, ok := e.contract(name)
	if !ok {
		return nil
	}
	args, known := e.BoundArguments(at)
	if !known {
		return nil
	}
	out := make([]Sink, 0, len(summary.Sinks))
	for _, s := range summary.Sinks {
		v := substitute(s.Argument, args, true)
		s.Argument = v
		out = append(out, s)
	}
	return out
}

func snapshotOptions(s *Snapshot) Options {
	if s == nil {
		return Options{}
	}
	return s.options
}
