---
id: UselessReturn
group: Confusing constructs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UselessReturn

## Summary

Two pointless `return` forms: a bare `return;` as the very last statement of a
function body (the function ends there anyway), and `return $local = expr;`
where the assignment to a local variable can never be observed after the
function has returned.

## Detection

### Trailing bare return

- **D1** A function-like with a body: named function, method (non-abstract),
  or closure (`function () { … }`). Arrow functions have no body and are not
  concerned.
- **D2** The **last statement** of that body (the last direct child statement
  of the outer `{ … }`; trailing comments and doc comments are ignored) is a
  `return` without a value (`return;`).
  Only the top level of the body counts: a `return;` that is the last
  statement inside an `if`/loop block is not reported.

### Assignment in return

- **D3** A `return` statement whose value is directly (no parentheses) an
  assignment whose target is a plain variable `$v` (not a property, array
  element, static property, list/array destructuring or variable-variable),
  e.g. `return $v = expr;`.
- **D4** The `return` is inside a function-like (named function, method or
  closure); the nearest enclosing one is the *scope*. A `return` at file top
  level is ignored.
- **D5** None of the following holds (each makes the assignment observable):
  - **D5a** the scope has a by-reference parameter named `v` (`&$v`);
  - **D5b** `v` is bound by reference in a `use` list: either the scope
    itself is a closure with `use (… &$v …)`, or any closure nested anywhere
    inside the scope's body has `use (… &$v …)`;
  - **D5c** the scope's body contains (at any depth) a `static` declaration of
    `$v` (`static $v;`, `static $a, $v = 1;`);
  - **D5d** the nearest `try` statement enclosing the `return` (searching
    outwards but not past the scope boundary) has a `finally` block whose
    body mentions a variable named `v` anywhere (any depth).
  Variable-name comparisons are case-sensitive and exclude the `$`.

## Exceptions (no report)

- **E1** `return;` that is not the last top-level statement of the body, and
  `return null;` / `return <value>;` at the end.
- **E2** `return ($v = expr);` (parenthesised value).
- **E3** Assignments to by-reference parameters, by-reference `use`-bound
  variables, `static` variables, or variables read in a related `finally`.
- **E4** Abstract methods and interface methods (no body).

## Report

- D1–D2: range = the whole `return;` statement including the `;`.
  Severity: info. Message: `Redundant 'return;' at the end of the body;
  remove it.` (No fix.)
- D3–D5: range = the whole `return` statement, from `return` to the `;`
  inclusive. Severity: info. Message: `The assigned variable is never used
  after returning; return the value directly.`

## Fix

- **F1** (D3–D5 only) Replace the whole return statement with
  `return ` + source text of the assigned value + `;`.
  `return $total = $a + $b;` → `return $a + $b;`.
  The source text of the value is copied verbatim (comments and line breaks
  inside it are kept).
- The trailing bare `return;` (D1–D2) has **no** fix; the statement stays.

## Options

None.

## PHP versions

No gating.

## Examples

```php
<?php
function build(array $parts, &$out) {
    $memo = null;
    $join = function ($sep, &$acc) use (&$memo, $parts) {
        if ($memo !== null) {
            return $memo = implode($sep, $parts);
        }
        if ($acc === []) {
            return $acc = $parts;
        }
        <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $sep = strtoupper($sep);</weak_warning>
    };
    if ($out) {
        return $out = $join;
    }
    <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $parts = $join;</weak_warning>
}

function clear(array &$rows) {
    if (!$rows) { return; }
    $rows = [];
    <weak_warning descr="Redundant 'return;' at the end of the body; remove it.">return;</weak_warning>
}

function counter() {
    static $n = 0;
    return $n = $n + 1;
}

function guarded() {
    try {
        return $res = compute();
    } finally {
        release($res);
    }
}

function wrapped() {
    return ($tmp = compute());
}
```

```php
<?php
function build(array $parts, &$out) {
    $memo = null;
    $join = function ($sep, &$acc) use (&$memo, $parts) {
        if ($memo !== null) {
            return $memo = implode($sep, $parts);
        }
        if ($acc === []) {
            return $acc = $parts;
        }
        return strtoupper($sep);
    };
    if ($out) {
        return $out = $join;
    }
    return $join;
}

function clear(array &$rows) {
    if (!$rows) { return; }
    $rows = [];
    return;
}

function counter() {
    static $n = 0;
    return $n = $n + 1;
}

function guarded() {
    try {
        return $res = compute();
    } finally {
        release($res);
    }
}

function wrapped() {
    return ($tmp = compute());
}
```

## Divergences

- Upstream treats compound assignments (`return $v .= 'x';`, `return
  $v += 1;`) as assignments too, and its fix would produce `return 'x';` —
  changing the result. Recommendation: only report plain `=` (and `= &`)
  assignments; skip compound ones. No fixture covers compound forms.
- Variables imported with `global $v;` (or `$GLOBALS`) are reported upstream
  although the assignment is observable. Recommendation: also skip `v` when
  the scope body declares `global $v`. No fixture covers it.
