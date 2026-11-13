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

// Extract computes compact summaries while the project's parser tree is
// available. It does not open any other files. A supplied snapshot permits
// bounded composition with previously summarized project wrappers.
func Extract(file *syntax.File, ix *index.Index, php phpversion.Version, snapshot *Snapshot, options Options) *File {
	_, out := Summarize(file, infer.NewEnv(file, names.New(file), ix, php, infer.Options{Assertions: options.Assertions}), snapshot, options)
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
func Summarize(file *syntax.File, types *infer.Env, snapshot *Snapshot, options Options) (*Env, *File) {
	e, order := summaryInventory(file, types, snapshot, options)
	s := e.summarizing
	s.track = map[string]map[string]uint8{}
	out := s.out
	for _, n := range order {
		e.summarize(n)
	}
	finish(e, out, e.answers)
	return e, out
}

// finish publishes on out what summarizing it read: the contracts consulted
// and, from answers, which the snapshot held and which the file served.
func finish(e *Env, out *File, answers map[string]answer) {
	out.callables = e.summarizing.track
	e.summarizing = nil
	out.consulted, e.consulted = e.consulted, nil
	for key, read := range answers {
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
}

// summaryInventory builds declaration metadata without following callable bodies.
func summaryInventory(file *syntax.File, types *infer.Env, snapshot *Snapshot, options Options) (*Env, []syntax.Node) {
	e := New(file, types, snapshot, options)
	e.consulted = map[string]bool{}
	out := &File{options: e.options, Path: file.Path, Summaries: map[string]Summary{}}
	s := &summarizing{out: out, base: e.snapshot, declared: map[string][]syntax.Node{}, names: map[syntax.Node]string{}, state: map[syntax.Node]uint8{}}
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
	// Reuse the declaration inventory instead of walking the file again on
	// the first local producer or nested-contract query.
	e.localBodies = make(map[string]syntax.Node, len(s.declared))
	for key, declarations := range s.declared {
		for _, declaration := range declarations {
			switch declaration.(type) {
			case *syntax.Function, *syntax.Method:
				if _, duplicate := e.localBodies[key]; duplicate {
					e.localBodies[key] = nil
				} else {
					e.localBodies[key] = declaration
				}
			}
		}
	}
	e.summarizing = s
	return e, order
}

// summarizing is the state of Summarize: the summaries made so far, each
// callable of the file under its key, and whether its summary is begun
// (summaryBegun) or made (summaryMade).
type summarizing struct {
	base         *Snapshot
	cached       map[string]Summary
	borrowed     map[syntax.Node]*scopeResult
	dependencies map[syntax.Node]map[string]bool
	out          *File
	declared     map[string][]syntax.Node
	names        map[syntax.Node]string
	state        map[syntax.Node]uint8
	// track records what each callable read (see File.callables); stack
	// holds the callables being summarized, innermost last.
	track map[string]map[string]uint8
	stack []string
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
	if s.track != nil {
		defer e.tracking(s, s.names[n])()
	}
	if s.dependencies != nil {
		reads := e.begin()
		defer func() { e.end(reads); s.dependencies[n] = reads }()
	}
	if prior := s.borrowed[n]; prior != nil && prior.finished {
		e.scopes[n] = prior
	}
	r := e.scope(n)
	sum := Summary{Return: Value{Complete: r.complete}, argumentSensitive: r.argumentSensitive}
	body := syntax.VariableScopeBody(n)
	sum.Generator = r.generator
	for i, v := range r.returns {
		v = e.constrainCallbackReturn(n, scalarSnapshot(v))
		if i == 0 {
			sum.Return = scalarSnapshot(v)
		} else {
			sum.Return = merge(sum.Return, scalarSnapshot(v))
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
			left := r.exit.value(received(p.Var.Name))
			if p.Variadic {
				// The arguments keep what they held, and may hold
				// what was written to an element of the array.
				left = merge(left, r.exit.value(p.Var.Name))
			}
			if sum.ByRef == nil {
				sum.ByRef = map[int]Value{}
			}
			sum.ByRef[i] = Value{Complete: left.Complete, Sources: left.Sources, Safe: left.Safe, scalars: left.scalars}
		}
	}
	sum.Return = e.completedReturn(n, r, sum.Return)
	sum.forward = e.summarizeForward(n, r, sum.Return)
	if sum.Fixed {
		sum.pure = e.compilePure(n, r, nil, 0)
		sum.defaults = e.literalDefaults(n)
	}
	sum.resource = e.summarizeResource(n, r, sum)
	sum.Return = summarizeRecord(n, r, sum)
	sum.effects = summarizeResourceEffects(n, r, sum)
	if sum.Fixed && !byReference(n) {
		sum.returned = e.summarizeReturned(sum.Return, r, len(syntax.VariableScopeParams(n)))
	}
	sum.Return.returned = nil
	sum.Return.targets = nil
	sum.Return.Expr = nil
	shapeFuel := MaxSources
	sum.Return.elements = compactElements(sum.Return.elements, MaxDepth, &shapeFuel)
	sum.Return.interval = nil
	sum.Return.origins = nil
	sum.Return.Identity, sum.Return.storedIdentity = 0, 0
	sum.Return.record = nil
	if !sum.Generator {
		for _, c := range r.calls {
			sum.Sinks = appendSinks(sum.Sinks, e.nativeSinks(c))
			sum.Sinks = appendSinks(sum.Sinks, e.Sinks(c.Node))
		}
		sum.Sinks = appendSinks(sum.Sinks, e.commands(body, r))
	}
	s.out.Summaries[s.names[n]] = sum
	s.state[n] = summaryMade
}

// appendSinks appends to sinks those of more it does not list yet. A
// callable calling itself reads its own sinks back from its summary: they
// count once, so its summary settles.
func appendSinks(sinks, more []Sink) []Sink {
	for _, s := range more {
		if !slices.ContainsFunc(sinks, func(o Sink) bool {
			return o.Span == s.Span && o.Name == s.Name && o.Context == s.Context && o.Path == s.Path && reflect.DeepEqual(o.Argument, s.Argument)
		}) {
			sinks = append(sinks, s)
		}
	}
	return sinks
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
	if e.local != nil {
		return NewLocal(e.file, e.types, s, e.options, e.local.guard)
	}
	differ := changedAnswers(e.answers, s.lookup)
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

func changedAnswers(answers map[string]answer, lookup func(string) (Summary, bool)) map[string]bool {
	differ := map[string]bool{}
	for key, read := range answers {
		summary, ok := lookup(key)
		if read.own && read.held || read.ok != ok || !reflect.DeepEqual(read.summary, summary) {
			differ[key] = true
		}
	}
	return differ
}
