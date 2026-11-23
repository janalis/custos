package pattern

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/types"
)

// Group is the rule group of every custom rule.
const Group = "Custom"

var (
	idRE      = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	metaVarRE = regexp.MustCompile(`\$[A-Z][A-Z0-9_]*`)
)

// snippetPrefix opens the PHP source a pattern is parsed from.
const snippetPrefix = "<?php "

// snippet is a parsed pattern or fix template.
type snippet struct {
	text  string
	file  *syntax.File
	root  syntax.Node   // the expression of an expression statement, else the statement; nil for a sequence
	stmts []syntax.Stmt // the statements of a sequence of several
	expr  bool          // root is an expression
	vars  []string      // metavariables, in order of first appearance
}

// nodes returns the snippet's root, or the statements of a sequence.
func (s *snippet) nodes() []syntax.Node {
	if s.root != nil {
		return []syntax.Node{s.root}
	}
	out := make([]syntax.Node, len(s.stmts))
	for i, st := range s.stmts {
		out[i] = st
	}
	return out
}

// constraint is a compiled Where.
type constraint struct {
	typ   types.Type
	typed bool
	regex *regexp.Regexp
	not   []*snippet
}

// Compile checks d and returns the rule it declares.
func Compile(d Definition) (analysis.Rule, error) {
	r, err := compile(d)
	if err != nil {
		if d.ID == "" {
			return nil, fmt.Errorf("custom rule: %w", err)
		}
		return nil, fmt.Errorf("custom rule %s: %w", d.ID, err)
	}
	switch {
	case r.semantic && r.seq:
		return &semanticSeqRule{r}, nil
	case r.semantic:
		return &semanticRule{r}, nil
	case r.seq:
		return &seqRule{r}, nil
	}
	return r, nil
}

func compile(d Definition) (*rule, error) {
	if err := CheckID(d.ID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(d.Message) == "" {
		return nil, fmt.Errorf("message is required")
	}
	if len(d.Match) == 0 {
		return nil, fmt.Errorf("match is required")
	}
	canonical, _ := json.Marshal(d) // plain data: cannot fail
	sum := sha256.Sum256(canonical)
	r := &rule{
		fingerprint: hex.EncodeToString(sum[:]),
		id:          d.ID, message: d.Message, severity: diagnostic.SeverityWarning, enabled: true,
		fixTitle: d.FixTitle, description: d.Description,
	}
	switch d.Severity {
	case "":
	case "error", "warning", "info":
		r.severity = diagnostic.Severity(d.Severity)
	default:
		return nil, fmt.Errorf("invalid severity %q", d.Severity)
	}
	if d.Enabled != nil {
		r.enabled = *d.Enabled
	}
	if d.PHP != "" {
		v, err := phpversion.Parse(d.PHP)
		if err != nil {
			return nil, err
		}
		r.php = v
	}
	var err error
	if r.paths, err = globs(d.Paths); err != nil {
		return nil, err
	}
	if r.exclude, err = globs(d.Exclude); err != nil {
		return nil, err
	}
	for _, text := range d.Match {
		p, err := parseSnippet(text)
		if err != nil {
			return nil, fmt.Errorf("match %q: %w", text, err)
		}
		kinds := []syntax.NodeKind{p.Kind()}
		switch {
		case p.root == nil:
			// A run of statements, found in every statement list; the
			// search slides, so leading `$_;` add nothing.
			for len(p.stmts) > 0 && stmtMeta(p.stmts[0]) == wildcard {
				p.stmts = p.stmts[1:]
			}
			p.stmts = mergeWildcards(p.stmts)
			if !slices.ContainsFunc(p.stmts, func(s syntax.Stmt) bool { return stmtMeta(s) == "" }) {
				return nil, fmt.Errorf("match %q: a sequence needs a statement that is not a lone metavariable", text)
			}
			r.seq, kinds = true, []syntax.NodeKind{syntax.KBlock, syntax.KCase, syntax.KNamespace}
		case isMetaVar(p.root):
			return nil, fmt.Errorf("match %q: a pattern must not be a lone metavariable", text)
		}
		r.patterns = append(r.patterns, p)
		r.semantic = r.semantic || resolvesNames(p)
		for _, k := range kinds {
			if !slices.Contains(r.kinds, k) {
				r.kinds = append(r.kinds, k)
			}
		}
	}
	for _, list := range []struct {
		name  string
		texts Patterns
		out   *[]*snippet
	}{{"inside", d.Inside, &r.inside}, {"notInside", d.NotInside, &r.notInside}} {
		for _, text := range list.texts {
			p, err := single(text)
			if err != nil {
				return nil, fmt.Errorf("%s %q: %w", list.name, text, err)
			}
			*list.out = append(*list.out, p)
			r.semantic = r.semantic || resolvesNames(p)
		}
	}
	// Every metavariable used outside the patterns must be bound by each.
	bound := func(where, name string) error {
		for _, p := range r.patterns {
			if !slices.Contains(p.vars, name) {
				return fmt.Errorf("%s uses %s, which pattern %q does not bind", where, name, p.text)
			}
		}
		return nil
	}
	for _, name := range metaVarRE.FindAllString(d.Message, -1) {
		if err := bound("message", name[1:]); err != nil {
			return nil, err
		}
	}
	if len(d.Where) > 0 {
		r.where = map[string]*constraint{}
	}
	for key, w := range d.Where {
		name := strings.TrimPrefix(key, "$")
		if !metaName(name) {
			return nil, fmt.Errorf("where key %q is not a metavariable", key)
		}
		if err := bound("where", name); err != nil {
			return nil, err
		}
		c := &constraint{}
		if w.Type != "" {
			c.typ = types.FromDoc(w.Type, nil)
			if c.typ.IsUnknown() {
				return nil, fmt.Errorf("where %s: invalid type %q", key, w.Type)
			}
			c.typed = true
			r.semantic = true
		}
		if w.Regex != "" {
			re, err := regexp.Compile(w.Regex)
			if err != nil {
				return nil, fmt.Errorf("where %s: %w", key, err)
			}
			c.regex = re
		}
		for _, text := range w.Not {
			p, err := single(text)
			if err != nil {
				return nil, fmt.Errorf("where %s: not %q: %w", key, text, err)
			}
			c.not = append(c.not, p)
			r.semantic = r.semantic || resolvesNames(p)
		}
		r.where[name] = c
	}
	r.semantic = r.semantic || r.resolvesCallees()
	if d.Fix != nil {
		t, err := parseTemplate(*d.Fix)
		if err != nil {
			return nil, fmt.Errorf("fix %q: %w", *d.Fix, err)
		}
		for _, name := range t.vars {
			if name == wildcard {
				return nil, fmt.Errorf("fix %q: $_ binds nothing", *d.Fix)
			}
			if err := bound("fix", name); err != nil {
				return nil, err
			}
		}
		for _, p := range r.patterns {
			// An expression template replacing statements stands as an
			// expression statement; statements cannot replace an expression.
			if t.snip != nil && !t.snip.expr && p.expr {
				return nil, fmt.Errorf("fix %q: statements cannot replace an expression", *d.Fix)
			}
		}
		r.fix = t
		if r.fixTitle == "" {
			r.fixTitle = "Apply the " + d.ID + " fix"
		}
	}
	return r, nil
}

// CheckID checks the ID of a rule outside the catalogue: an uppercase
// letter then letters and digits, naming no built-in rule.
func CheckID(id string) error {
	if !idRE.MatchString(id) {
		return fmt.Errorf("id %q must start with an uppercase letter and hold only letters and digits", id)
	}
	if _, ok := meta.Lookup(id); ok {
		return fmt.Errorf("id %s names a built-in rule", id)
	}
	return nil
}

// globs checks and normalizes path globs.
func globs(list []string) ([]string, error) {
	var out []string
	for _, g := range list {
		if err := checkGlob(g); err != nil {
			return nil, err
		}
		out = append(out, normalizeGlob(g))
	}
	return out, nil
}

// resolvesNames reports a pattern naming a function or constant: an
// unqualified call or constant in a namespace reaches the namespace's own
// declaration when the project declares one, which only the project index
// knows.
func resolvesNames(p *snippet) bool {
	found := false
	for _, root := range p.nodes() {
		syntax.Inspect(root, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.FuncCall:
				_, named := n.Name.(*syntax.Name)
				found = found || named
			case *syntax.ConstFetch:
				switch strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)) {
				case "true", "false", "null":
				default:
					found = true
				}
			}
			return !found
		})
	}
	return found
}

// resolvesCallees reports a metavariable standing for a called function's
// name whose resolved name counts: constrained by a regex or by names, or
// repeated. An unqualified call in a namespace resolves to the namespace's
// function when the project declares one, which only the project index
// knows.
func (r *rule) resolvesCallees() bool {
	snippets := slices.Concat(r.patterns, r.inside)
	for _, p := range snippets {
		for _, root := range p.nodes() {
			found := false
			syntax.Inspect(root, func(n syntax.Node) bool {
				call, ok := n.(*syntax.FuncCall)
				if !ok {
					return true
				}
				name := metaVarName(call.Name)
				if name == "" || name == wildcard {
					return true
				}
				c := r.where[name]
				found = found || c != nil && (c.regex != nil || len(c.not) > 0) || occurrences(snippets, name) > 1
				return !found
			})
			if found {
				return true
			}
		}
	}
	return false
}

// occurrences counts the uses of metavariable name in snippets.
func occurrences(snippets []*snippet, name string) int {
	count := 0
	for _, p := range snippets {
		for _, root := range p.nodes() {
			syntax.Inspect(root, func(n syntax.Node) bool {
				if metaVarName(n) == name {
					count++
				}
				return true
			})
		}
	}
	return count
}

// mergeWildcards keeps one `$_;` of each run: two in a row match what one
// does, with more ways to split the statements.
func mergeWildcards(stmts []syntax.Stmt) []syntax.Stmt {
	return slices.CompactFunc(stmts, func(a, b syntax.Stmt) bool {
		return stmtMeta(a) == wildcard && stmtMeta(b) == wildcard
	})
}

// single parses a pattern that must be one expression or statement.
func single(text string) (*snippet, error) {
	p, err := parseSnippet(text)
	if err == nil && p.root == nil {
		err = fmt.Errorf("want one expression or statement, got %d statements", len(p.stmts))
	}
	return p, err
}

// Kind returns the kind of the snippet's root (that of its first
// statement for a sequence).
func (s *snippet) Kind() syntax.NodeKind { return s.nodes()[0].Kind() }

// stmtMeta returns the metavariable a statement is alone (`$S;`), or "".
func stmtMeta(s syntax.Node) string {
	if es, ok := s.(*syntax.ExprStmt); ok {
		return metaVarName(es.Expr)
	}
	return ""
}

// wildcard is the metavariable that matches anything and binds nothing.
const wildcard = "_"

// metaVarName returns the metavariable n is ($X, $_), or "".
func metaVarName(n syntax.Node) string {
	v, ok := n.(*syntax.Variable)
	if !ok || v.NameExpr != nil {
		return ""
	}
	if v.Name == wildcard || metaName(v.Name) && !syntax.Superglobal(v.Name) {
		return v.Name
	}
	return ""
}

// metaName reports a metavariable name: an uppercase letter, then
// uppercase letters, digits and underscores.
func metaName(name string) bool {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'A' && c <= 'Z' || i > 0 && (c >= '0' && c <= '9' || c == '_')) {
			return false
		}
	}
	return name != ""
}

func isMetaVar(n syntax.Node) bool { return metaVarName(n) != "" }

// parseSnippet parses a PHP expression, statement or sequence of
// statements.
func parseSnippet(text string) (*snippet, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty pattern")
	}
	var f *syntax.File
	for _, src := range []string{snippetPrefix + text + ";", snippetPrefix + text} {
		f = syntax.Parse("pattern.php", []byte(src), syntax.Options{Version: phpversion.Max})
		if !f.Recovered() {
			break
		}
	}
	// A compile error (`new C(...)`) leaves the tree whole: a pattern uses
	// `...` as a wildcard where PHP would not build a closure.
	if e, recovered := f.RecoveryError(); recovered {
		return nil, fmt.Errorf("not valid PHP: %s", e.Msg)
	}
	stmts := slices.DeleteFunc(slices.Clone(f.Stmts), func(s syntax.Stmt) bool { return s.Kind() == syntax.KNop })
	if len(stmts) == 0 {
		return nil, fmt.Errorf("want an expression or statements")
	}
	s := &snippet{text: text, file: f, root: stmts[0]}
	if len(stmts) > 1 {
		s.root, s.stmts = nil, stmts
	} else if es, ok := stmts[0].(*syntax.ExprStmt); ok {
		// The expression alone: a fix must not carry the `;` written after it.
		s.root, s.expr = es.Expr, true
		s.text = string(f.Src[len(snippetPrefix):es.Expr.Span().End])
	}
	for _, root := range s.nodes() {
		syntax.Inspect(root, func(n syntax.Node) bool {
			if name := metaVarName(n); name != "" && !slices.Contains(s.vars, name) {
				s.vars = append(s.vars, name)
			}
			return true
		})
	}
	return s, nil
}
