---
id: DegradedSwitch
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# DegradedSwitch

## Summary

A `switch` with only a `default` branch, or with a single `case` (optionally
plus `default`), is just an `if` / `if-else` (or plain code) in disguise and
reads better rewritten that way.

## Detection

For every `switch` statement (brace form or alternative `switch (...): … endswitch;`
form), let *C* be the number of `case <expr>:` labels (each `case` label counts
separately, even when several labels share one body / fall through) and *D*
whether a `default:` label exists.

D1. *C* = 0 and *D*: only a default branch → report (kind "default only").
D2. *C* = 1 and no *D*: behaves like an `if` → report (kind "if").
D3. *C* = 1 and *D*: behaves like an `if`/`else` → report (kind "if-else").

Body content of the branches is irrelevant (empty bodies, bodies with only
`break`, fall-through all count the same).

## Exceptions (no report)

E1. A `switch` with no labels at all (`switch ($v) {}`).
E2. Two or more `case` labels, with or without `default` — including two
    labels sharing one body (`case 1: case 2: …`).

## Report

- Range: the `switch` keyword token only.
- Severity: info (fixture markup `weak_warning`).
- Message (per kind):
  - D1: `This switch only has a default branch; keep just its body.`
  - D2: `This switch has a single case; an 'if' is clearer.`
  - D3: `This switch has a single case and a default; an 'if'/'else' is clearer.`

## Fix

None.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function route($verb) {
    <weak_warning descr="This switch has a single case; an 'if' is clearer.">switch</weak_warning> ($verb) {
        case 'GET':
            serve();
    }
    <weak_warning descr="This switch has a single case and a default; an 'if'/'else' is clearer.">switch</weak_warning> (strtoupper($verb)) {
        default:
            reject();
            break;
        case 'POST':
            store();
            break;
    }
    <weak_warning descr="This switch only has a default branch; keep just its body.">switch</weak_warning> ($verb):
        default:
            fallback();
    endswitch;

    switch ($verb) {}
    switch ($verb) {
        case 'PUT':
        case 'PATCH':
            update();
    }
}
```

## Divergences

- A (deprecated, PHP < 7 only valid) `switch` with two `default` labels and no
  `case` is counted once as "default only"; upstream behaves the same as far
  as can be told. No action.
