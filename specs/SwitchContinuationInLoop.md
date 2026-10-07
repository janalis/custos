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
- **F1** Replace the whole statement with `continue 2;`.

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
