---
id: RuleID
group: Code style
kind: syntax            # syntax | semantic
needs: []               # any of: names, index, hierarchy, types, flow, stubs, composer
php: { min: "", max: "" }
---

# RuleID

## Summary

One or two sentences, in our own words: what problem this rule points out and why it matters.

## Detection

Precise, implementation-neutral conditions (node kinds, argument shapes, contexts).
Number them so the implementation and tests can reference them (D1, D2, …).

## Exceptions (no report)

Cases that look similar but must not be reported (E1, E2, …).

## Report

- Range: which node/token is highlighted (exact start/end).
- Severity: default from `internal/meta/rules.json`.
- Message: our own wording (may use placeholders).

## Fix

What the code becomes (F1, F2, …), when the fix is unavailable, how
whitespace/comments are preserved. "None" if the rule has no fix.

## Options

| Option | Type | Default | Effect |

## PHP versions

Gating and version-dependent behaviour.

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

Intentional differences from upstream behaviour, with reason. Empty if none.
