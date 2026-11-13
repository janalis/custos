// Package flow provides bounded, file-local value and state analysis.
// Environments are confined to one analysis; snapshots contain no syntax trees.
package flow

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
)

// Context distinguishes incompatible escaping contracts.
type Context uint16

const (
	HTML Context = 1 << iota
	SQL
	Shell
	Header
	URL
	Path
	// HTMLSingleQuoted and HTMLDoubleQuoted prove conversion of the matching
	// delimiter in an ordinary quoted data attribute.
	HTMLSingleQuoted
	HTMLDoubleQuoted
	// HTMLScript proves text safe in JavaScript: a script element, or an
	// event handler attribute once its entities are decoded. HTMLRaw proves
	// text safe where markup is read again or as style: a style element or
	// attribute, a srcdoc document. HTMLURL proves that text cannot start a
	// URL attribute with a scheme, and HTMLUnquoted that it cannot end an
	// unquoted attribute value.
	HTMLScript
	HTMLRaw
	HTMLURL
	HTMLUnquoted
)

// allContexts holds every contract: what a value without request data is
// safe for.
const allContexts = HTML | SQL | Shell | Header | URL | Path | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLScript | HTMLRaw | HTMLURL | HTMLUnquoted

// Covers reports whether safe holds every contract of context.
func (safe Context) Covers(context Context) bool {
	return safe&context == context
}

// Limits bound hostile scopes and recursive wrapper queries. MaxRetained
// bounds retained block bindings, states and immutable origin cells of one
// scope (memory). MaxCopied bounds bindings copied, joined and compared while
// iterating to a fixed point (time). An ordinary long function carries a hundred
// locals through a few hundred blocks, far below both.
const (
	MaxBlocks    = 8192
	MaxTransfers = 100000
	MaxRetained  = 8 * MaxTransfers
	MaxCopied    = 64 * MaxTransfers
	MaxSources   = 64
	MaxDepth     = 8
)

// Source records an originating input, or a symbolic parameter in a summary.
type Source struct {
	Kind      string
	Path      string
	Span      syntax.Span
	Parameter int // -1 for external input
	// ValuesOnly means that only an array parameter's values contribute.
	ValuesOnly bool
	// Keys records keys moved into an element or scalar output. Such a
	// dependency cannot be made safe by projecting the original values.
	Keys bool
	// htmlDecoded prevents caller escaping from surviving a decoding transform.
	htmlDecoded bool
}

// Value is a conservative union of possible inputs. Safe is the intersection
// of proven escaping contexts. Expr and Identity require agreement on every path.
type Value struct {
	record   *resourceReturn
	literal  *constantScalar
	returned *callableValue
	targets  *callableSet
	interval *intInterval
	Sources  []Source
	Expr     syntax.Expr
	// elements holds the literal-keyed elements of an array variable that
	// are tracked on their own. Only a binding carries them: the value of an
	// expression never does.
	elements *elementSet
	// origins lists the expressions that may have produced a binding the
	// paths disagree on (Expr is then nil), each with what its own paths
	// established. It means nothing while Expr is set.
	origins *originSet
	// tested and unless belong to a binding only, never to the value of an
	// expression. tested is the short-circuit operation whose left operand
	// last read the binding, which was not written since. unless is such an
	// operation whose right operand assigns the binding and tests it in
	// place: the binding is not false (unlessFalse) or not null (unlessNull)
	// except where the last evaluation of the operation skipped that
	// operand. Between the evaluation of the operation and the branch on its
	// outcome, skippedFalse and skippedNull record that the value the
	// skipped operand left in place was not false or not null.
	tested, unless *syntax.Binary
	Length         int64
	Identity       uint32
	storedIdentity uint32
	Invalidated    uint32
	Safe           Context
	Complete       bool
	NonFalse       bool
	NonNull        bool
	LengthKnown    bool
	headerMask     uint8
	// upload marks the table of uploaded files or one of its entries as PHP
	// built it: only there are the server-chosen fields what they claim.
	upload uint8
	// scalars marks an array whose values are numbers, booleans or null
	// while its keys remain whatever the request sent.
	scalars bool
	// matched marks request data a switch found equal to one of its string
	// labels: the text of a label, not what the request sent.
	matched bool
	// unlessFalse, unlessNull, skippedFalse and skippedNull qualify unless.
	unlessFalse, unlessNull, skippedFalse, skippedNull bool
	// replaced marks a binding that code the analysis does not follow may
	// have replaced since it was last written, on some path.
	replaced bool
	// entry marks the binding of a static variable that still holds what it
	// held when its declaration was reached: nothing wrote it since.
	entry bool
	// plain marks a value that is not an object: converting it to text,
	// comparing it, iterating it or reading an offset of it runs no method.
	// deep marks one that, when it is an array, holds no object either, at
	// any depth: what is done with its elements runs none.
	plain, deep bool
	truth       uint8
}

// State is the most recent operation common to every reaching path.
type State struct {
	Operation    string
	Span         syntax.Span
	Successful   bool
	closeChecked bool // the exact fclose result was proven true
}

// Call describes a reached call and its argument values (in written order).
type Call struct {
	Node      syntax.Node
	Name      string
	Receiver  syntax.Expr
	Arguments []Value
}

// Options controls runtime assumptions used for flow proofs.
type Options struct {
	Assertions bool
	// Vendor, when set, answers contracts no file of a snapshot declares from
	// dependency files (see Vendor). Snapshots built with different providers
	// are not compatible.
	Vendor *Vendor
}

// Env shares one bounded analysis per variable scope among all inspections.
type Env struct {
	sharing     *sharingFlags
	recordSites map[recordSite]uint32
	options     Options
	file        *syntax.File
	types       *infer.Env
	scopes      map[syntax.Node]*scopeResult
	declarationCaches
	summaryResolver
	globalsState
}

// declarationCaches holds what the Env learned of the declarations the code
// reaches.
type declarationCaches struct {
	declarations declarationCache
	classKeys    map[string]string
	arrowNames   map[*syntax.ArrowFunction]captureNames
	// overridden caches, per receiver class and method, whether a class of
	// the project replaces the method (see replaced).
	overridden map[[2]string]cachedAnswer
	// extendedBy caches whether a class of the project extends a class.
	extendedBy map[string]bool
	// properties caches, per class and property name, whether reading the
	// property of the object itself runs no code (see backed).
	properties map[[2]string]cachedAnswer
	// kindOf caches what each declared type read admits (see kinds).
	kindOf map[string]uint8
	// statics caches, per scope, what its code does to each of its static
	// variables (see declared).
	statics map[syntax.Node]map[string]static
}

// summaryResolver holds what the Env needs to answer the contracts of other
// functions: the snapshot, the summaries being computed, and what each
// computation read.
type summaryResolver struct {
	invocationDepth int
	snapshot        *Snapshot
	// consulted records the contracts read while summaries are extracted.
	consulted map[string]bool
	// answers holds what those contracts gave, and summarizing the state of
	// the summaries being computed, which serve the file's own.
	answers     map[string]answer
	summarizing *summarizing
	local       *localSummaries
	// reading holds, innermost last, the contracts each result being
	// computed read meanwhile; globalReads and staticReads those the cached
	// globals and static variable facts were computed from.
	reading     []map[string]bool
	globalReads map[string]bool
	staticReads map[syntax.Node]map[string]bool
	localBodies map[string]syntax.Node
	// nested holds the functions nestedSpares is looking into.
	nested map[syntax.Node]bool
}

// globalsState holds what the Env learned of the globals the file's
// functions write.
type globalsState struct {
	// global caches the file's function-bound global names once scanned, and
	// globalTaint the request data those functions write to them.
	global       map[string]bool
	globalTaint  map[string]Value
	writers      map[globalWriter]bool
	globalsKnown bool
	globalsReady bool
}

type scopeResult struct {
	sharing *sharingFlags

	resourceInputs     map[uint32]bool
	fallsThrough       bool
	argumentSensitive  bool
	distinctReferences map[syntax.Node]bool
	callables          map[syntax.Expr]string
	boundMethods       map[syntax.Expr]*boundMethod
	allocations        map[uint32]bool
	exitStates         map[uint32]State
	closures           map[syntax.Expr]map[string]Value
	closureSinks       map[syntax.Node][]Sink
	parameterReads     map[string]uint8
	parameterObserved  map[string]uint8
	closedReceivers    map[syntax.Node]bool
	gaps               map[string]syntax.Span
	originTypes        *infer.Env
	originKeys         map[syntax.Expr]string
	originReserve      int
	originContracts    map[string]bool
	generator          bool
	values             valueStore
	// reads names the contracts the result was computed from, with those of
	// the results and facts it read; nil when Summarize did not compute it.
	reads     map[string]bool
	before    map[syntax.NodeIdentity]map[uint32]State
	calls     []Call
	returns   []Value
	callIndex map[syntax.Node]int
	complete  bool
	finished  bool
	// cut is where dynamic code (include, eval) first made the scope
	// incomplete, when nothing else had before and the control flow is
	// otherwise complete: what is known of an expression evaluated before
	// it, outside any loop around it, still holds (see Excludes).
	cut uint32
	// stale lists the assignments tested in place in a short-circuit
	// operand whose earlier result can reach the path that skips them.
	stale map[*syntax.Assign]bool
	// matching marks a scope in which a switch matched a request table
	// element with a fixed text (see frame.match).
	matching bool
	// leaves marks a scope no path of which completes normally.
	leaves bool
	// spent marks a scope that ran out of one of its memory budgets:
	// nothing more is worth computing for it.
	spent      bool
	operations int
	facts      int
	retained   int
	copies     int
	// global marks a file-level scope whose variables a function of the file
	// can replace through the globals table.
	global bool
	// holders is the request data closures of the scope write to the names
	// they capture by reference.
	holders map[string]Value
	// exit is the request data each variable may hold when the scope ends,
	// and globalWrites what the scope stores in literal globals table entries.
	exit         bindings
	globalWrites map[string]Value
	// globalNames are the names the scope binds with `global`.
	globalNames map[string]bool
	// statics are the names the scope binds with `static`, each with what a
	// call that enters the function again can do to it (see staticKept).
	statics map[string]static
	// broken names the static variables the analysis assumed never to be
	// false where the function can be entered again, and found that they may
	// be. The scope is then analysed again without the assumption for the
	// names in distrust, or for every name (distrustAll).
	broken, distrust map[string]bool
	distrustAll      bool
	// runs marks a scope that may run code of the project other than its
	// own: a call that is not known to make none, a constructor, a callback.
	runs bool
	// modern marks a scope analysed for PHP 8 or later, where writing a
	// property of a value that is no object is an error.
	modern bool
	// receivers tells, for each method call on a variable, what the variable
	// holds wherever the call is reached (see follows).
	receivers map[syntax.Node]uint8
	// escaped memoizes what each evaluated call let escape (see
	// invalidateEscaped), so a fluent chain is not walked again per link.
	escaped map[syntax.Node]escapeSet
	// labelled is where the first `goto` label of the scope is written, plus
	// one: a statement after it can be reached without running the ones
	// written before it.
	labelled uint32
	// landings is read when a rule first asks about a label of the scope.
	landings *landings
}

// static describes a static variable of a scope: what a call that enters the
// function again can do to it, and what the code of the scope gives it (see
// declared).
type static struct {
	state               uint8
	starts, plain, deep bool
}

const (
	// staticKept: the function never assigns the variable; another
	// invocation leaves it as it is.
	staticKept uint8 = iota
	// staticHeld: the function assigns it, and it is assumed not to be
	// false wherever the function, having written it, calls out, reaches
	// its declaration again or ends, which is checked.
	staticHeld
	// staticLoose: the function assigns it and nothing is assumed.
	staticLoose
)

// New creates an environment without computing scope flow eagerly.
func New(file *syntax.File, types *infer.Env, snapshot *Snapshot, options Options) *Env {
	if snapshot != nil && snapshot.options != options {
		snapshot = nil
	}
	types = types.WithOptions(infer.Options{Assertions: options.Assertions})
	return &Env{options: options, file: file, types: types, summaryResolver: summaryResolver{snapshot: snapshot}, scopes: map[syntax.Node]*scopeResult{}}
}

// written reports the value of a literal: text the program wrote.
func (v *Value) written() bool {
	switch v.Expr.(type) {
	case *syntax.Literal, *syntax.StringPart:
		return v.Complete && len(v.Sources) == 0
	}
	return false
}

// producers exports local syntax only after the whole scope was followed.
// Partial generator or budget-limited bodies cannot establish that a write ran.
func (r *scopeResult) producers(v Value) *originSet {
	if !r.complete || r.generator {
		return nil
	}
	out := origins(Value{}, v, r)
	if !r.complete {
		return nil
	}
	return out
}
