# Adding or changing a rule

The catalogue contains EA-modeled rules and native custos rules. Rule work
includes fixing a false positive, a missed case or a fix. The workflow is the
same either way: **spec first, then code**, done separately (see the
[clean-room process](./clean-room)).

Native rules have independently authored facts in
`internal/inspection/meta/native-rules.json`; `make extract` only regenerates
the EA-modeled facts in `rules.json`. Native rules have no legacy PhpStorm
name and no EA conformance requirement. Their own fixtures remain mandatory.

## 1. The spec

`specs/<ID>.md` (start from `specs/_TEMPLATE.md`) describes the rule's
behavior in our own words. It has YAML front matter
(`id`, `group`, `kind: syntax|semantic`, `needs`, `php: { min, max }`) and
these sections:

| Section | Content |
|---|---|
| Summary | What the rule reports and why, for users. Shown by `custos explain` and on the rule's page. |
| Detection | Numbered conditions (D1, D2…) for a report. |
| Exceptions | Numbered cases (E1, E2…) that are not reported. |
| Report | Range, severity and message. |
| Fix | What the quick-fix produces (F1…), if there is one. |
| Options | Each option, its type, default and effect. Shown to users. |
| PHP versions | Version gating. |
| Examples | New PHP examples using the [fixture markup](./fixtures#markup). The first block is shown on the rule's page, and the block right after it as the result of the fix. |
| Divergences | Intentional differences from upstream, and why. |

A spec describes observable behavior only: no Go identifiers, repository
paths, analysis budgets or advice to implementers. Those belong in the rule's
[implementation notes](#implementation-notes). For a behavior change, update
the spec first, in the same pull request.

## 2. The implementation

Rules live in `internal/inspection/rules/<lowercase-ID>/`. `rule.go` carries
the package comment and `New()`; the private implementation type lives there
or, for larger rules, in `<rule_snake>.go`:

```go
// rule.go
// Package unnecessarysemicolon implements the UnnecessarySemicolon inspection.
package unnecessarysemicolon

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessarySemicolon{} }
```

```go
// unnecessary_semicolon.go
type unnecessarySemicolon struct{}

func (unnecessarySemicolon) ID() string { return "UnnecessarySemicolon" }

func (unnecessarySemicolon) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KNop, syntax.KEcho}
}

func (r unnecessarySemicolon) Check(ctx *analysis.Context, n syntax.Node) {
	// follow the spec's D/E items in order; ctx.Report(span, msg, fixes...)
}
```

- List only the node kinds the rule needs; the engine dispatches by kind.
  Whole-file rules also implement `analysis.FileRule` (`CheckFile`).
- Semantic data is lazy: `ctx.Names()`, `ctx.Index()`, `ctx.TypeOf()` and
  `ctx.Flow()`. A rule that needs declarations from other files implements
  the `Semantic()` marker (`analysis.SemanticRule`); one that needs
  project-local callable summaries implements `Flow()` (`analysis.FlowRule`).
- Keep rule-specific helpers and companion tests in the inspection package.
  Shared helpers live in `astquery`, `flowquery` and `semanticquery`; see
  [Inspection helpers](./architecture#inspection-helpers).
- Add one constructor to `internal/inspection/catalogue/catalogue.go`.
  Preserve group execution order and ID sorting within each group;
  `make architecture` checks complete and unique registration.
- A fix is a `diagnostic.Fix{Title, Edits}` whose `Edits` function returns text
  edits on exact byte ranges. A fix must never change behavior or produce
  invalid PHP; when that cannot be guaranteed, do not offer it.
- Messages are short and imperative: *"Stray semicolon; remove it."*
- Version-gate the rule on `ctx.PHP` when the spec says so.

### Implementation notes

When a rule has something worth keeping that the spec must not say, add
`internal/inspection/rules/<lowercase-ID>/NOTES.md`: a non-obvious algorithm,
budgets and limits, reliance on shared helpers, or known gaps. Create it only
when there is such content. Start with `# <RuleID> implementation notes`,
then use free `##` sections, written in our own words.

## 3. Fixtures and checks

1. Write or extend the [fixtures](./fixtures) in
   `internal/inspection/rules/<lowercase-ID>/testdata/`.
2. `make fixtures RULE=<ID>`.
3. For EA-modeled rules, `make conformance RULE=<ID>` when you have the
   upstream checkout. A mismatch means the spec is wrong or incomplete, or the
   difference is an intentional divergence to document.
4. Touching a hot path? Add or extend a benchmark (`make bench`).
5. `make rules-doc`, then `make verify`.
6. Add a line under `## [Unreleased]` in `CHANGELOG.md`.
