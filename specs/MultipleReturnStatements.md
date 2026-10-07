---
id: MultipleReturnStatements
group: Architecture
kind: syntax
needs: []
php: { min: "", max: "" }
---

# MultipleReturnStatements

## Summary
A method with many exit points is harder to follow than one with a single
result path. The rule counts `return` statements per method and reports it
above two configurable thresholds (a normal one and a severe one).

## Detection
- **D1** Every method declaration with a body: methods of classes, abstract
  classes, traits, enums and anonymous classes. Abstract methods and
  interface methods are skipped. Plain functions and closures are **not**
  inspected (only methods).
- **D2** `R` = number of `return` statements belonging to the method itself:
  every `return` (with or without a value) anywhere in the method body,
  excluding those inside nested closures, arrow functions, nested function
  declarations and anonymous-class methods (those belong to their own
  function-like; anonymous-class methods are inspected on their own).
- **D3** If `R >= SCREAM_THRESHOLD` → report with severity **error**.
  Else if `R >= COMPLAIN_THRESHOLD` → report with severity **warning**.

## Exceptions (no report)
- **E1** Methods with fewer than `COMPLAIN_THRESHOLD` returns.
- **E2** Abstract / interface methods.
- **E3** Functions and closures outside classes, and returns of closures
  nested in a method (counted for nobody unless the closure is a method).

## Report
- Range: the method name identifier.
- Severity: `error` when `R >= SCREAM_THRESHOLD`, otherwise `warning` (the
  catalogue default).
- Message: `{R} return statements in this method; try to funnel them into a
  single exit.`

## Fix
None.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| `COMPLAIN_THRESHOLD` | int | 3 | Return count (inclusive) from which a warning is raised. |
| `SCREAM_THRESHOLD` | int | 5 | Return count (inclusive) from which an error is raised instead. |

## PHP versions
No gating.

## Examples
Defaults (3 / 5):

```php
<?php
class Grader {
    public function <warning descr="3 return statements in this method; try to funnel them into a single exit.">letter</warning>(int $score) {
        if ($score > 90) {
            return 'A';
        }
        if ($score > 75) {
            return 'B';
        }
        return 'C';
    }

    public function <error descr="6 return statements in this method; try to funnel them into a single exit.">bucket</error>(int $n) {
        switch ($n) {
            case 1: return 'one';
            case 2: return 'two';
            case 3: return 'three';
            case 4: return 'four';
            case 5: return 'five';
        }
        return 'many';
    }

    public function sorter() {
        return function ($a, $b) {
            if ($a < $b) { return -1; }
            if ($a > $b) { return 1; }
            return 0;
        };
    }

    public function widget() {
        return new class {
            public function <warning descr="3 return statements in this method; try to funnel them into a single exit.">kind</warning>($v) {
                if (is_int($v)) { return 'int'; }
                if (is_string($v)) { return 'string'; }
                return 'other';
            }
        };
    }
}

function plain($x) {
    if ($x) { return 1; }
    if (!$x) { return 2; }
    return 3;
}
```

## Divergences
- Upstream counts the `return` instructions that flow directly into the
  method's exit in its control-flow graph rather than return statements in
  the source. Consequences not covered by fixtures: a `return` inside a
  `try` (or `catch`) that has a `finally` block probably flows through the
  finally block and is then not counted; returns in unreachable code may
  still be counted. Recommendation: count every `return` statement of the
  method as in D2 (simpler and stable); revisit if a conformance case shows a
  difference.
