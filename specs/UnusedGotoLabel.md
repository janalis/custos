---
id: UnusedGotoLabel
group: Unused
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnusedGotoLabel

## Summary

A `goto` label that no `goto` statement targets is dead code; it only
suggests control flow that does not exist. Remove it.

## Detection

Visit every goto label statement (`name:`).

- **D1** Find the nearest enclosing function-like scope of the label:
  function, method or closure (an arrow function cannot contain
  statements). If there is none (label in top-level code, in a namespace
  body, etc.) → no report.
- **D2** That scope has a `{ … }` body (abstract/interface methods have none
  and cannot contain labels anyway).
- **D3** Collect every `goto X;` statement anywhere inside that body, at any
  depth, **except** inside nested function-likes (closures, arrow
  functions, nested function declarations) and class bodies (anonymous
  classes included): a `goto` can only jump within its own function.
- **D4** Report the label when none of those `goto` statements names it.
  Name comparison is exact and case-sensitive (`goto End;` does not use
  label `end:`).

## Exceptions (no report)

- **E1** Labels outside any function/method/closure.
- **E2** Labels targeted by at least one `goto` in the same scope body
  (before or after the label, any nesting depth of blocks/loops, but not
  from a nested function-like).

## Report

- Range: the label statement — the identifier and its colon (`unusedLabel:`),
  from the first character of the name to the `:` inclusive.
- Severity: info (rendered deprecated/strikethrough; fixture markup
  `weak_warning`).
- Message: `Label '{name}' is never targeted by a goto; remove it.`

## Fix

- **F1** Delete the label statement (`name:`). Also delete the whitespace
  run immediately preceding it (so the line disappears rather than leaving
  an indented blank line); text after the label is untouched.
  - `a:\n    b:\n\n    if …` (with `b` unused) → `a:\n\n    if …`.

(Output is compared whitespace-collapsed, so deleting either the preceding
or the following whitespace run is acceptable.)

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
function retry_job($job)
{
    $attempts = 0;
    again:
    <weak_warning descr="Label 'done' is never targeted by a goto; remove it.">done:</weak_warning>
    <weak_warning descr="Label 'Again' is never targeted by a goto; remove it.">Again:</weak_warning>
    if (!$job->run() && ++$attempts < 3) {
        goto again;
    }

    $cleanup = function () {
        finish:
        return true;
        goto finish;
    };
}

start:
echo "top level";
```

```php
<?php
function retry_job($job)
{
    $attempts = 0;
    again:
    if (!$job->run() && ++$attempts < 3) {
        goto again;
    }

    $cleanup = function () {
        finish:
        return true;
        goto finish;
    };
}

start:
echo "top level";
```

## Divergences

- **Nested functions (custos diverges).** Upstream searches the whole
  function body, nested closures included, so an outer label counts as used
  when only a `goto` inside a closure names it — a jump PHP cannot perform
  (it would fail to compile). custos stops the search at nested
  function-likes and classes, so such a label is reported (D3).
