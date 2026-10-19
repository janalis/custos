# Inspection ownership

`analysis` (dispatch, context, suppressions), `meta` (independent rule facts
and descriptions), `catalogue` (explicit constructors), `rules/<lowercase-ID>`
(one package per inspection), `astquery`, `flowquery`, `semanticquery` and
`phpunit`. See `docs/contributing/architecture.md` and
`docs/contributing/adding-a-rule.md`.

Implement from clean-room specs only. Each rule package owns its private
implementation, companions and tests, and exposes `New()` from `rule.go`.
Never import another rule package. Add constructors explicitly in catalogue
order. Keep helpers local until shared; then place them by lexical
(`astquery`), local-flow (`flowquery`) or resolved (`semanticquery`)
behavior. Lazy edits keep exact byte spans; rules own fix safety.

Specs stay behavior-only; record a rule's non-obvious algorithm, budgets,
shared-helper reliance or known gaps in its optional `NOTES.md`
(`# <RuleID> implementation notes`, free `##` sections, own words).

Run `make architecture`, single-ID fixtures/conformance and coverage; then
`make rules-doc` when specs, fixtures or status change, then `make verify`.
