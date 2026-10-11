---
id: SwitchContinuationInLoop
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SwitchContinuationInLoop

## Summary

PHP treats `switch` as a loop structure for `continue`, so a bare `continue`
inside a `switch` only leaves the `switch` (like `break`) instead of starting
the next iteration of the surrounding loop. `continue 2` is what was meant.

## Detection

- **D1** A `continue` statement without a level argument (`continue;`).
  Any argument, including `continue 1;`, disables the check.
- **D2** Walk up its ancestors:
  - reaching the file, or a function/method/closure/arrow function boundary,
    first → no report;
  - remember whether a `switch` has been passed;
  - at the first loop ancestor (`for`, `foreach`, `while`, `do … while`): if a
    `switch` was passed on the way → report; otherwise no report. The walk
    stops there in both cases.

## Exceptions (no report)

- **E1** `continue` whose nearest enclosing loop lies inside the `switch`
  (the loop is met before the `switch`).
- **E2** `continue N;` with any explicit level.
- **E3** `switch` with no enclosing loop inside the current function (no
  report).

## Report

- Range: the whole `continue` statement, from `continue` through its `;`
  inclusive.
- Severity: error.
- Message: `Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.`

## Fix

- **F1** Insert the level right after the `continue` keyword
  (`continue;` → `continue 2;`). Nothing else is rewritten, so the statement
  terminator (`;` or a closing tag `?>`) and any comment before it are kept
  as written.
- **F2** No fix (report only) when a `finally` block lies between the
  `continue` and the loop: PHP rejects `continue N` jumping out of `finally`
  at compile time, while the bare `continue` is legal there.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function dispatch(array $events) {
    while ($e = array_shift($events)) {
        switch ($e->type) {
            case 'skip':
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            case 'retry':
                if ($e->tries > 3) {
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
                }
                continue 2;
            case 'batch':
                for ($i = 0; $i < 2; $i++) {
                    continue;
                }
                break;
            case 'later':
                $cb = function () { continue; };
                break;
        }
    }
}
```

```php
<?php
function dispatch(array $events) {
    while ($e = array_shift($events)) {
        switch ($e->type) {
            case 'skip':
                continue 2;
            case 'retry':
                if ($e->tries > 3) {
                    continue 2;
                }
                continue 2;
            case 'batch':
                for ($i = 0; $i < 2; $i++) {
                    continue;
                }
                break;
            case 'later':
                $cb = function () { continue; };
                break;
        }
    }
}
```

## Divergences

- With nested switches inside a loop (`loop { switch { switch { continue; } } }`)
  upstream still reports and suggests `continue 2`, which then only targets
  the outer `switch`. Recommendation: count the `switch` levels passed and use
  `continue {switches + 1}` in the fix (message unchanged); for a single
  switch this equals upstream. Not covered by upstream fixtures.
- **No fix when statements follow the switch (custos diverges).** `continue`
  in a switch runs the statements after the switch; `continue 2` skips
  them. When the switch (or a statement enclosing it inside the loop) is
  followed by further statements in the loop body, the finding has no
  quick-fix (SuiteCRM's SQL parser keeps `++$i;` after the switch in step
  with the foreach position).
- **No fix out of a `finally` block (custos diverges).** A `continue` in a
  switch inside `finally { }` whose loop is outside the `try` is reported
  without a quick-fix: `continue 2` there is a compile-time fatal error
  ("jump out of a finally block"), whereas the original code runs.
- **The fix only inserts the level (custos diverges).** Rewriting the whole
  statement as `continue 2;` breaks a statement terminated by a closing tag
  (`continue ?>` would lose the tag and leave the following `<?php` inside
  PHP code) and drops comments placed before the `;`. custos inserts ` 2`
  after the keyword instead.
