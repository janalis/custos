# Adding or changing a rule

Every one of the 178 rules is implemented. Most rule work is now fixing a
false positive, a missed case or a fix. The workflow is the same either
way: **spec first, then code**, done separately (see the
[clean-room process](./clean-room)).

## 1. The spec

`specs/<ID>.md` (start from `specs/_TEMPLATE.md`) describes the rule in our
own words. It has YAML front matter
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

For a behaviour change, update the spec first, in the same pull request.

## 2. The implementation

Rules live in `internal/inspection/rules/<lowercase-ID>/`. The implementation
type stays private; `New()` constructs it:

```go
type unnecessarySemicolon struct{}

func New() analysis.Rule { return unnecessarySemicolon{} }

func (unnecessarySemicolon) ID() string { return "UnnecessarySemicolon" }

func (unnecessarySemicolon) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KNop, syntax.KEcho}
}

func (r unnecessarySemicolon) Check(ctx *analysis.Context, n syntax.Node) {
	// follow the spec's D/E items in order; ctx.Report(span, msg, fixes...)
}
```

- List only the node kinds the rule needs; the engine dispatches by kind.
- Use `ctx.Names()`, `ctx.Index()` and `ctx.TypeOf()` for semantic
  information; they are computed lazily.
- Keep rule-specific helpers and companion tests in the inspection package.
  Shared lexical, flow and resolved helpers belong in `inspection/astquery`,
  `inspection/flowquery` and `inspection/semanticquery`, respectively.
- Add one constructor to `inspection/catalogue/catalogue.go`. Preserve group
  execution order and ID sorting within each group; `make architecture` checks
  complete and unique registration.
- A fix is a `diagnostic.Fix{Title, Edits}` whose `Edits` function returns text
  edits on exact byte ranges. A fix must never change behaviour or produce
  invalid PHP; when that cannot be guaranteed, do not offer it.
- Messages are short and imperative: *"Stray semicolon; remove it."*
- Version-gate the rule on `ctx.PHP` when the spec says so.

## 3. Fixtures and checks

1. Write or extend the [fixtures](./fixtures) in `testdata/rules/<ID>/`:
   positives, false positives, `.fixed.php` for fixes, a sidecar per option
   or PHP version.
2. `make fixtures RULE=<ID>`.
3. `make conformance RULE=<ID>` when you have the upstream checkout. A
   mismatch means the spec is wrong or incomplete, or the difference is an
   intentional divergence to document.
4. Touching a hot path? Add or extend a benchmark (`make bench`).
5. `make rules-doc`, then `make verify`.
6. Add a line under `## [Unreleased]` in `CHANGELOG.md`.
