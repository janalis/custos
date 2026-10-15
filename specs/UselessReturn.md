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

- **D3** A `return` statement whose value is directly (no parentheses) a
  plain by-value assignment (`=`; neither a compound operator nor `= &`)
  whose target is a plain variable `$v` (not a property, array element,
  static property, list/array destructuring or variable-variable), e.g.
  `return $v = expr;`.
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
  - **D5d** any `try` statement enclosing the `return` (every one met while
    searching outwards, up to but not past the scope boundary) has a
    `finally` block whose body mentions a variable named `v` anywhere (any
    depth). All of those `finally` blocks run when the function returns;
  - **D5e** the scope's body binds `v` by reference somewhere (at any depth,
    before or after the `return`), so that writing to it writes through to
    another storage location:
    - a by-reference `foreach` value: `foreach ($xs as &$v)`,
      `foreach ($xs as $k => &$v)`;
    - a reference assignment with `$v` on either side: `$v = &expr`,
      `$other = &$v`;
    - a by-reference array or destructuring item whose value is `$v`:
      `[&$v] = $pair`, `list('k' => &$v) = $row`, `$refs = [&$v]`.
  Variable-name comparisons are case-sensitive and exclude the `$`.

## Exceptions (no report)

- Reference-returning property getters keep assignment returns: replacing an
  assignment expression with its right-hand variable can change what is
  returned by reference. Ordinary value-returning getter bodies use the local
  assignment checks.

- **E1** `return;` that is not the last top-level statement of the body, and
  `return null;` / `return <value>;` at the end.
- **E2** `return ($v = expr);` (parenthesised value).
- **E3** Assignments to by-reference parameters, by-reference `use`-bound
  variables, `static` variables, variables read in the `finally` of any
  enclosing `try`, or variables bound by reference in the scope (D5e).
- **E5** Reference assignments: `return $v = &$store['k'];`. Taking a
  reference creates a missing array element or property silently, which
  reading the right-hand side does not.
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
  changing the result. Only plain `=` assignments are reported; compound
  ones are skipped. No fixture covers compound forms.
- Reference assignments (`return $v = &expr;`) are not reported: binding a
  reference creates a missing array element or property without a notice,
  so returning `expr` directly is not equivalent.
- A variable that the scope binds by reference (by-reference `foreach`
  value, `$v = &expr`, `$other = &$v`, `&$v` array/destructuring item) is
  not reported: the assignment writes through the reference and stays
  observable after the return.
- Every enclosing `try` with a `finally` mentioning the variable suppresses
  the report, not only the nearest `try`: all of them run on return.
- Variables imported with `global $v;` (or `$GLOBALS`) are reported upstream
  although the assignment is observable. `v` is skipped when the scope body
  declares `global $v`. No fixture covers it.
