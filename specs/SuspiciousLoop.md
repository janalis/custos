---
id: SuspiciousLoop
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SuspiciousLoop

## Summary

Two loop mistakes: a `for` condition section with several comma-separated
expressions (only the last one decides whether the loop continues), and loop
variables that silently overwrite a parameter of the enclosing function or the
variable of an outer loop.

## Detection

### Multiple conditions

- **D1** A `for` statement whose condition section (between the two `;`)
  contains two or more comma-separated expressions
  (`for ($i = 0; $i < $n, $ok; $i++)`). Always checked, regardless of options.

### Loop variables (only when `VERIFY_VARIABLES_OVERRIDE` is true)

Applies to `for` and `foreach` statements.

- **D2** Loop variable names of a loop:
  - `for`: for each expression of the init section that is a plain `=`
    assignment whose target is a simple variable (`$i = 0`), that
    variable's name, unless the assigned value reads that variable
    (`$i = $i + 1`, custos); compound assignments (`$i += $step`) and
    other init expressions are ignored;
  - `foreach`: the names of the key and value variables (including a
    by-reference value `&$v`, and the variables of a `list(...)`/`[...]`
    destructuring value).
- **D3** Parameter clash: let `F` be the nearest enclosing function, method,
  closure or arrow function. For every loop variable name equal to one of
  `F`'s parameter names, report (one report per clashing name). No enclosing
  function (top-level code) → no parameter check.
- **D4** Outer-loop clash: walk up the ancestors of the loop until reaching a
  function/method/closure/arrow function boundary or the file. For every
  ancestor that is a `for` or `foreach`, compute its loop variable names
  (D2); for every name shared with the current loop, report (one report per
  shared name per ancestor loop). `while`/`do-while` ancestors are ignored.

## Exceptions (no report)

- **E1** `for` with zero or one condition expression.
- **E2** Loop variables that match neither a parameter nor an outer
  `for`/`foreach` variable; closure `use` variables are not parameters.
- **E3** Outer loops beyond a function/closure boundary.
- **E4** With `VERIFY_VARIABLES_OVERRIDE` false, D2–D4 are skipped entirely.
- **E5** D1 is not reported when every expression before the last is an
  assignment (plain or compound), an increment/decrement or a function,
  method or static call: such a list runs statements on each iteration and
  only means the last expression as the condition
  (`for ($x = 0; $c = $x * 2, $x++ < 75;)`).
- **E6** D4 is not reported for an outer `foreach` whose overwritten
  variable is dead: neither loop takes the variable by reference, and the
  name is not mentioned after the inner loop inside the outer one, nor
  anywhere in a loop that sits between them (its condition or later
  passes would see the inner value), nor after the outer loop or before
  it inside a loop that encloses it (a variable-variable counts as a
  mention). The next outer iteration assigns the variable again, so
  nothing is lost.

## Report

- Range: the loop keyword token only (`for` or `foreach`), for every report.
  Several reports may share the same keyword (one per clashing name / outer
  loop, plus possibly D1).
- Severity: error.
- Messages:
  - D1: `Only the last expression of the 'for' condition is evaluated as the condition; combine them with && or ||.`
  - D3: `Loop variable '${name}' overwrites a {kind} parameter.` where
    `{kind}` is `method` when `F` is a method, otherwise `function` (also for
    closures and arrow functions).
  - D4: `Loop variable '${name}' overwrites a variable of an outer loop.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `VERIFY_VARIABLES_OVERRIDE` | bool | `true` | Enables the parameter and outer-loop clash checks (D2–D4). |

## PHP versions

No gating.

## Examples

```php
<?php
<error descr="Only the last expression of the 'for' condition is evaluated as the condition; combine them with && or ||.">for</error> ($n = 0; $n < 5, $n != 3; $n++) {}

function scan(array $rows, $pos) {
    <error descr="Loop variable '$rows' overwrites a function parameter.">foreach</error> ($rows as $rows) {}
    <error descr="Loop variable '$pos' overwrites a function parameter.">for</error> ($pos = 1; $pos < 3; $pos++) {}
    foreach ($rows as $r) {}
}

class Grid {
    public function walk($cell) {
        <error descr="Loop variable '$cell' overwrites a method parameter.">foreach</error> ([1, 2] as $cell) {}
    }
}

foreach ($matrix as $row => $cols) {
    <error descr="Loop variable '$row' overwrites a variable of an outer loop.">for</error> ($row = 0, $k = 1; $row < 2; $row++) {}
    while ($cols) {
        <error descr="Loop variable '$cols' overwrites a variable of an outer loop.">foreach</error> ($cols as $cols) {}
    }
    array_map(function () {
        foreach ([] as $row) {}
    }, []);
    foreach ($cols as $c) {}
    echo $row;
}
```

## Divergences

- **Statements before the condition (custos diverges, D1).** Upstream
  reports any condition section with several expressions. custos stays
  silent when all but the last are assignments, increments or calls (E5):
  they run for their effect each iteration, as in php-src's benchmark
  loops, and joining them with `&&` would change what the loop does.

- **Loops going on from the given value (custos diverges, D2).** Upstream
  counts any assignment of the `for` init. `for ($index += $direction; …)`
  or `for ($i = $i + 1; …)` starts from the value the parameter (or the
  outer loop) holds: nothing it was given is lost, so custos does not take
  the variable as the loop's own (php-cs-fixer, amphp/hpack).
- **Dead outer variables (custos diverges, E6).** Upstream reports every
  shared name. `foreach ($chunks as $chunk) { foreach (explode("\n",
  $chunk) as $chunk) { … } }` reads the outer value once, to start the
  inner loop, and never again: the reuse loses nothing (Symfony's glob
  resource and container dumper).
- Upstream gathers loop variable names in a hash set, so the order of
  multiple reports on one keyword is unspecified. custos reports in source
  order of the clashing variables.
- A `foreach` loop whose key and value use the same name as one parameter
  yields one report per distinct name.
- Destructured `foreach` values (`foreach ($a as [$x, $y])`) are assumed to
  contribute their variables to D2; upstream behavior there is unverified
  (no fixture); custos counts them.
