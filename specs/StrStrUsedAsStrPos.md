---
id: StrStrUsedAsStrPos
group: Performance
kind: syntax
needs: []
php: { min: "", max: "" }
---

# StrStrUsedAsStrPos

## Summary

`strstr()` / `stristr()` build and return a substring. When the result is only
used as a yes/no answer ("does the haystack contain the needle?"), that copy is
wasted work: a position search (`strpos()` / `stripos()`) compared strictly
against `false` answers the same question without allocating.

## Detection

- **D1** A plain function call (not a method or static call) that resolves to
  the global function (names compared case-insensitively, as PHP does; `\` and
  a global `use function` import are fine, but a same-named function declared
  in the current namespace, one imported from another namespace, or a
  qualified non-global name such as `Ns\strstr` does not count): `strstr` or
  `stristr`. Call it `C`. Map `strstr` → `strpos`, `stristr` → `stripos`.
- **D2** `C` has **at least 2** arguments. Only the first two are used in the
  replacement; a 3rd argument (`before_needle`) is dropped.
- Then the two patterns below are tried in order; the first that matches
  produces the (single) report.

### Pattern A — explicit comparison with `false`

- **D3** The **direct** parent of `C` (no parentheses in between) is a binary
  expression `B` whose operator is `==`, `!=`, `===` or `!==`. `C` may be the
  left or right operand.
- **D4** The other operand of `B` is the constant `false`, written unqualified
  in any letter case (`false`, `FALSE`, `False`).
- Operator for the replacement: the operator text of `B`; if it is 2
  characters long a `=` is appended (`==` → `===`, `!=` → `!==`); 3-character
  operators are kept.
- Reported node: `B`.

If D3 holds but D4 does not (e.g. `true == strstr(...)`, `strstr(...) === ''`),
pattern A fails and pattern B is tried; B then fails because the parent is a
comparison, so nothing is reported.

### Pattern B — used as a boolean condition

Let `P` be the first ancestor of `C` that is not a parenthesized expression
(skip any number of wrapping parentheses); `W` is the node directly below `P`
on that path (`C` itself or its outermost wrapping parenthesis).

- **D5** One of:
  - `P` is an `if` or `elseif` / `else if` statement and `W` is its condition;
  - `P` is a `while` or `do … while` loop and `W` is its condition;
  - `P` is a logical-not `!`;
  - `P` is a binary expression with operator `&&`, `||`, `and` or `or`
    (either side; `xor` does **not** count);
  - `P` is a full ternary `a ? b : c` (not the short `a ?: c`) and `W` is its
    condition `a`.
  (`for` conditions, `match`, `switch`, assignments, returns, arguments… do not
  count.)
- Operator / reported node:
  - if the **direct** parent of `C` is a unary `!` (no parentheses between
    `!` and `C`): operator `===`, reported node is that `!` expression;
  - otherwise operator `!==`, reported node is `C` itself (this includes
    `!(strstr(...))`, where only the call is replaced and the `!` and the
    parentheses are kept).

## Exceptions (no report)

- **E1** Fewer than 2 arguments.
- **E2** Compared with anything other than `false` (`true`, `null`, `''`, a
  variable, …), or with a non-equality operator (`<`, `<=>`, …).
- **E3** Wrapped in parentheses and then compared: `(strstr($a, $b)) === false`
  (pattern A needs the direct parent; pattern B rejects comparisons).
- **E4** Result used as a value: assigned, returned, passed as an argument,
  concatenated, used as a `for` condition, the short-ternary subject, the
  branches of a ternary, operand of `xor`.
- **E5** Name differing in case (`StrStr`), method calls `$o->strstr(...)`.

## Report

- Range: the reported node (pattern A: whole comparison `B` from its left
  operand start to its right operand end; pattern B: the `!` expression from
  `!` to the call's `)`, or just the call from its name — including any
  namespace qualifier — to its `)`).
- Severity: warning.
- Message: `Use '{replacement}' instead; it avoids building a substring.`

## Fix

- **F1** Replace the reported node with `{replacement}`:
  - `{call}` = `{qualifier}{fn}({arg1}, {arg2})` where `{qualifier}` is the
    namespace qualifier written before `C`'s name (empty or `\`; when empty
    and an unqualified `{fn}` at that position would not reach the global
    function — a `use function` import under that name, or a same-named function declared in the current namespace — it is `\`),
    `{fn}` is `strpos` / `stripos` per D1, `{arg1}`/`{arg2}` are the verbatim
    texts of `C`'s first two arguments, separated by `, `.
  - Regular comparison style (default): `{call} {op} false`.
    Yoda style: `false {op} {call}`. `false` is always lower-case; single
    spaces around `{op}`. The produced operand order depends only on the
    comparison style, not on the original order in `B`.
  - No parentheses are added around the replacement.
  Examples (regular):
  - `false == strstr($h, 'x')` → `strpos($h, 'x') === false`
  - `!stristr($h, $n)` → `stripos($h, $n) === false`
  - `$ok && \strstr($h, $n, true)` → `$ok && \strpos($h, $n) !== false`

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| (none) | | | Operand order in the fix follows the global comparison style (`regular` default / `yoda`). |

## PHP versions

No gating. Upstream fixtures run at the test default level.

## Examples

```php
<?php
function scan($text, $word, $flag) {
    if (<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) {}
    if (<warning descr="Use 'stripos($text, '#') === false' instead; it avoids building a substring.">!stristr($text, '#')</warning>) {}
    while ($flag and <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) { $flag = false; }
    $a = <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning> ? 1 : 2;
    $b = !(<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>);
    $c = <warning descr="Use '\strpos($text, $word) === false' instead; it avoids building a substring.">FALSE == \strstr($text, $word, true)</warning>;
    $d = <warning descr="Use 'stripos($text, $word) !== false' instead; it avoids building a substring.">stristr($text, $word) != false</warning>;

    $e = strstr($text, $word);
    $f = true == strstr($text, $word);
    $g = (strstr($text, $word)) === false;
    $h = strstr($text, $word) ?: 'none';
    $i = strstr($text);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}
```

```php
<?php
function scan($text, $word, $flag) {
    if (strpos($text, $word) !== false) {}
    if (stripos($text, '#') === false) {}
    while ($flag and strpos($text, $word) !== false) { $flag = false; }
    $a = strpos($text, $word) !== false ? 1 : 2;
    $b = !(strpos($text, $word) !== false);
    $c = \strpos($text, $word) === false;
    $d = stripos($text, $word) !== false;

    $e = strstr($text, $word);
    $f = true == strstr($text, $word);
    $g = (strstr($text, $word)) === false;
    $h = strstr($text, $word) ?: 'none';
    $i = strstr($text);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}
```

Yoda style: `!strstr($text, $word)` becomes `false === strpos($text, $word)`.

## Divergences

- **`<>` operator:** upstream appends `=` to any 2-character operator, so
  `strstr($a, $b) <> false` would become the invalid `<>=`. Recommendation:
  map `<>` to `!==`. Not covered by upstream fixtures.
- **Needle `''` / non-string needles:** behaviour of `strstr` and `strpos`
  differs only in pathological cases; upstream ignores this. Keep.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `strstr`/`stristr` by the name as written (case-sensitive, any namespace
  qualifier, no resolution), so a differently cased call such as `StrStr($t,
  $w) ? 1 : 0` is missed while a namespaced or imported user function of the
  same name is reported (and rewritten) as if it were the builtin. custos
  matches case-insensitively and only calls that reach the global function.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `strpos(`/`stripos(` in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
