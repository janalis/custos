package pattern

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/types"
)

// rule is a compiled custom rule.
type rule struct {
	id          string
	message     string
	description string
	severity    diagnostic.Severity
	enabled     bool
	php         phpversion.Version // minimum target version; 0 for any
	paths       []string
	exclude     []string
	patterns    []*snippet
	kinds       []syntax.NodeKind
	where       map[string]*constraint
	inside      []*snippet // enclosing code one of which must match
	notInside   []*snippet // enclosing code none of which may match
	seq         bool       // a pattern is a sequence of statements
	fix         *template
	fixTitle    string
	semantic    bool   // a type constraint needs the project index
	fingerprint string // digest of the definition
}

// semanticRule is a rule whose type constraints need the project index.
type semanticRule struct{ *rule }

func (*semanticRule) Semantic() {}

// seqRule is a rule with a sequence pattern: it also checks the file's
// top-level statements, which no dispatched node owns.
type seqRule struct{ *rule }

func (r *seqRule) CheckFile(ctx *analysis.Context) { r.checkFile(ctx) }

// semanticSeqRule is both a semanticRule and a seqRule.
type semanticSeqRule struct{ *rule }

func (*semanticSeqRule) Semantic()                         {}
func (r *semanticSeqRule) CheckFile(ctx *analysis.Context) { r.checkFile(ctx) }

// Fingerprint identifies the rule's definition (see analysis.Fingerprinted).
func (r *rule) Fingerprint() string { return r.fingerprint }

func (r *rule) ID() string               { return r.id }
func (r *rule) Kinds() []syntax.NodeKind { return r.kinds }

// Facts returns the rule's catalogue facts.
func (r *rule) Facts() meta.Rule {
	return meta.Rule{
		Native: true, ID: r.id, Group: Group, Severity: r.severity,
		EnabledByDefault: r.enabled, HasFix: r.fix != nil,
	}
}

// Description returns the rule's documentation.
func (r *rule) Description() meta.Description {
	d := meta.Description{Summary: r.description, Fix: r.fix != nil}
	if d.Summary == "" {
		d.Summary = r.message
	}
	if r.php != 0 {
		d.PHPMin = r.php.String()
	}
	return d
}

// Check reports n when it matches one of the rule's patterns, and the
// runs of statements of n matching a sequence pattern.
func (r *rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !r.applies(ctx) {
		return
	}
	b := r.budget(ctx)
	if b.spent {
		return
	}
	for _, p := range r.patterns {
		if p.root == nil {
			if list, ok := syntax.StmtListOf(n); ok {
				r.checkSeq(ctx, p, list)
			}
			continue
		}
		if p.root.Kind() != n.Kind() {
			continue
		}
		m := &matcher{ctx: ctx, budget: b, binds: b.binds[:0]}
		matched := m.match(p.root, n, roleClass, false) && r.accepts(m, n)
		over := spent(m, n.Span())
		if matched && !over {
			r.report(m, n, n.Span())
		}
		b.binds = m.binds
		if matched || over {
			return
		}
	}
}

// budgetKey keys a rule's budget among the file's shared data (a
// pointer-sized key: looking it up allocates nothing).
type budgetKey struct{ r *rule }

// budget returns the rule's budget in the file.
func (r *rule) budget(ctx *analysis.Context) *budget {
	return ctx.Shared(budgetKey{r}, func() any { return &budget{} }).(*budget)
}

// spent reports whether the rule's budget in the file is spent, so its
// last match is not trusted; the first time, it marks the file's analysis
// incomplete at span.
func spent(m *matcher, span syntax.Span) bool {
	if !m.budget.spent {
		return false
	}
	if !m.budget.reported {
		m.budget.reported = true
		m.ctx.Gap("query-budget", span)
	}
	return true
}

// checkFile checks the sequence patterns against the top-level statements.
func (r *rule) checkFile(ctx *analysis.Context) {
	if !r.applies(ctx) {
		return
	}
	for _, p := range r.patterns {
		if p.root == nil {
			r.checkSeq(ctx, p, ctx.File.Stmts)
		}
	}
}

// checkSeq reports the runs of list matching sequence pattern p, each
// statement in one run at most.
func (r *rule) checkSeq(ctx *analysis.Context, p *snippet, list []syntax.Stmt) {
	b := r.budget(ctx)
	for i := 0; i < len(list); {
		m := &matcher{ctx: ctx, budget: b, binds: b.binds[:0]}
		n, ok := m.stmts(p.stmts, list[i:], true)
		ok = ok && r.accepts(m, list[i])
		b.binds = m.binds
		if spent(m, list[i].Span()) {
			return
		}
		if !ok {
			i++
			continue
		}
		r.report(m, list[i], syntax.Span{Start: list[i].Span().Start, End: list[i+n-1].Span().End})
		i += n
	}
}

// applies reports whether the rule runs on the file at all.
func (r *rule) applies(ctx *analysis.Context) bool {
	return ctx.PHP >= r.php && r.inScope(ctx)
}

// accepts checks a match at n against the rule's context and constraints.
func (r *rule) accepts(m *matcher, n syntax.Node) bool {
	return r.inContext(m, n) && r.satisfied(m)
}

// report records a match covering span, which starts at matched.
func (r *rule) report(m *matcher, matched syntax.Node, span syntax.Span) {
	var fixes []diagnostic.Fix
	if r.fix != nil {
		if fix, ok := r.render(m, matched, span); ok {
			fix.Title = r.fixTitle
			fixes = append(fixes, fix)
		}
	}
	m.ctx.Report(span, r.interpolate(m), fixes...)
}

// inContext reports whether the code enclosing n matches one inside
// pattern, when there are some, and no notInside pattern. Inside patterns
// share the match's metavariables.
func (r *rule) inContext(m *matcher, n syntax.Node) bool {
	if len(r.inside) > 0 && !m.enclosed(r.inside, n) {
		return false
	}
	if len(r.notInside) == 0 {
		return true
	}
	// The probe's bindings are its own: appending copies the match's.
	probe := &matcher{ctx: m.ctx, binds: slices.Clip(m.binds), budget: m.budget}
	return !probe.enclosed(r.notInside, n)
}

// inScope reports whether the file lies within the rule's paths and
// outside its excluded ones (computed once per file).
func (r *rule) inScope(ctx *analysis.Context) bool {
	if len(r.paths) == 0 && len(r.exclude) == 0 {
		return true
	}
	return ctx.Memo("scope", func() any {
		p := ctx.ProjectPath()
		for _, g := range r.exclude {
			if matchGlob(g, p) {
				return false
			}
		}
		if len(r.paths) == 0 {
			return true
		}
		for _, g := range r.paths {
			if matchGlob(g, p) {
				return true
			}
		}
		return false
	}).(bool)
}

// satisfied checks the where constraints of the bound metavariables.
func (r *rule) satisfied(m *matcher) bool {
	for name, c := range r.where {
		b := m.lookup(name)
		text := m.ctx.SpanText(b.span)
		if c.regex != nil && !c.regex.MatchString(text) {
			return false
		}
		if b.seq {
			if c.typed || len(c.not) > 0 {
				return false // types and patterns describe one expression
			}
			continue
		}
		if c.typed {
			x, ok := b.node.(syntax.Expr)
			if !ok || types.Accepts(c.typ, m.ctx.TypeOf(x), semanticquery.AcceptContext(m.ctx, false)) != types.Yes {
				return false
			}
		}
		for _, p := range c.not {
			inner := &matcher{ctx: m.ctx, budget: m.budget}
			node := b.node
			if es, ok := node.(*syntax.ExprStmt); ok && p.expr {
				node = es.Expr // `f()` excludes the statement `f();`
			}
			if inner.match(p.root, node, roleClass, false) {
				return false
			}
		}
	}
	return true
}

// maxQuoted bounds the code a message quotes for one metavariable.
const maxQuoted = 60

// interpolate replaces the metavariables of the message by the code they
// bound, on one line.
func (r *rule) interpolate(m *matcher) string {
	if !strings.Contains(r.message, "$") {
		return r.message
	}
	return metaVarRE.ReplaceAllStringFunc(r.message, func(v string) string {
		// Compile checked every pattern binds the message's metavariables.
		text := strings.Join(strings.Fields(m.ctx.SpanText(m.lookup(v[1:]).span)), " ")
		if len(text) > maxQuoted {
			cut := maxQuoted - 3
			for !utf8.RuneStart(text[cut]) {
				cut--
			}
			text = text[:cut] + "..."
		}
		return text
	})
}
