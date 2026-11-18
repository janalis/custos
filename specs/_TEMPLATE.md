---
id: RuleID
group: Code style
kind: syntax            # syntax | semantic (project symbol index) | flow (cross-file callable summaries)
needs: []               # any of: names, index, hierarchy, types, flow, stubs, composer
php: { min: "", max: "" }   # set only when the whole rule is inactive outside the range
---

# RuleID

<!-- Behavior only: no Go identifiers, repository paths, analysis budgets or
advice to implementers. Those go in internal/inspection/rules/<lowercase-ID>/NOTES.md. -->

## Summary

One or two sentences, in our own words: what problem this rule points out and why it matters.

## Detection

Precise, implementation-neutral conditions (node kinds, argument shapes, contexts).
Number them so the implementation and tests can reference them (D1, D2, …).

## Exceptions (no report)

Cases that look similar but must not be reported (E1, E2, …).

## Report

- Range: which node/token is highlighted (exact start/end).
- Severity: default from `internal/inspection/meta/rules.json` (EA-modeled)
  or `native-rules.json` (native).
- Message: our own wording (may use placeholders).

## Fix

What the code becomes (F1, F2, …), when the fix is unavailable, how
whitespace/comments are preserved. "None" if the rule has no fix.

## Options

`None.` when the rule has no option; otherwise one row per option, its
effect described in words (no D/E/F labels):

| Option | Type | Default | Effect |
|---|---|---|---|
| NAME | bool | true | What turning it off changes. |

## PHP versions

Gating and version-dependent behavior; `No gating.` when there is none.

## Examples

New, minimal PHP snippets written for this spec (never copied from EA),
in the fixture markup format:

```php
<?php
// before
```

```php
<?php
// after fix
```

## Divergences

Intentional differences from upstream behavior, with reason. A native rule
starts with: Native custos inspection, independently specified; no EA
counterpart or EA conformance requirement.
