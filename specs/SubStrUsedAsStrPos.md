---
id: SubStrUsedAsStrPos
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# SubStrUsedAsStrPos

## Summary
Checking a prefix by cutting it out with `substr($s, 0, …)` and comparing the
piece is the long way of asking "does `$s` start with this?". The idiomatic
form is a position search compared strictly against `0`
(`strpos($s, $prefix) === 0`), or its case-insensitive variant when the cut
piece was case-folded first.

## Detection
- **D1** A function call (not method/static) that resolves to the global
  function `substr` or `mb_substr`. Throughout this rule, function names are
  compared case-insensitively as PHP does (`SubStr`, `\MB_SUBSTR` match), and
  a call resolves to a global function when written with a single leading
  `\`, or unqualified with no function of that name declared or imported in
  the current namespace; `Ns\substr(...)` and shadowing user functions never
  match. Call it `S`.
- **D2** `S` has **3 or 4** arguments.
- **D3** The 2nd argument of `S` is a number literal whose source text is
  exactly `0` (not `00`, `0x0`, `0.0`, `-0`, a variable or a constant).
- **D4** (The 3rd argument, `length`, is checked by D7 once the compared
  operand is known.)
- **D5** Optional case-folding wrapper: if `S` is directly an argument of a
  function call `W` (not a method call) resolving to one of `strtolower`,
  `strtoupper`, `mb_strtolower`, `mb_strtoupper`, `mb_convert_case`, and `W`
  has **exactly one** argument, then the *subject* is `W` and the rule is in
  *case-insensitive* mode. Otherwise the subject is `S` itself.
  (`mb_convert_case` always needs ≥ 2 arguments in valid code, so in practice
  it never qualifies.)
- **D6** The direct parent of the subject (no parentheses in between) is a
  binary expression `B` with operator `==`, `!=`, `===` or `!==` (`<>` is
  treated as the `!=` token too — see Divergences). The subject may be the left
  or the right operand. Let `O` be the other operand of `B`.
- **D7** `length` is the length of `O`, counted in the unit of `S`:
  - a call resolving to the global `strlen` when `S`
    is `substr`, with exactly 1 argument equivalent to `O`; or `mb_strlen`
    when `S` is `mb_substr`, with 1 or 2 arguments, the first equivalent to
    `O` (structural equivalence, as elsewhere); or
  - a decimal integer literal equal to the length of `O`, when `O` is a
    string literal without interpolation (decoded value; byte length for
    `substr`; for `mb_substr` the value must be pure ASCII, so bytes and
    characters agree).
  Anything else (`$n`, `3` against a variable, `strlen($other)`, a
  byte/character mix such as `substr($s, 0, mb_strlen($p))`) → no report.
- **D8** In case-insensitive mode, `O` must be a pure-ASCII string literal
  already in the folded case: no `a`–`z` for `strtoupper`/`mb_strtoupper`,
  no `A`–`Z` for `strtolower`/`mb_strtolower`. A variable or a mixed-case
  literal → no report.

If `S` is an argument of any other function call (e.g. `trim(substr(...))`)
or of a method call, or is wrapped in parentheses, the parent check fails and
nothing is reported.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** 2nd argument other than the literal `0`.
- **E2** Length unrelated to `O` (D7): a call of another function, a
  length measuring another value, a literal that differs from the literal's
  length, a variable, or a byte/character mix.
- **E3** 2 arguments only (`substr($s, 0)`) or more than 4.
- **E4** Subject not directly an operand of an equality/identity comparison
  (e.g. assigned, passed to another function, compared with `<`, wrapped in
  parentheses).
- **E5** Case-folding wrapper with a number of arguments other than 1.
- **E6** Case-folding wrapper compared with a variable or with a literal not
  already in the folded case (D8).

## Report
- Range: the whole comparison `B` (left operand start to right operand end).
- Severity: info (weak warning).
- Message: `Use '{replacement}' instead.` with the F1 text.

## Fix
- **F1** Replace `B` with `{replacement}`, built as follows:
  - `{fn}` = `{qualifier}` + (`mb_` if `S` is `mb_substr`) + (`stripos` in
    case-insensitive mode, else `strpos`). `{qualifier}` is `\` when `S`
    was written fully qualified (`\substr`), or when an unqualified call to
    that function at the reported position would not reach the global
    function (a `use function` import under that name, or a same-named function declared in the current namespace, e.g. `namespace R; function strpos() {…}` gives
    `\strpos($u, $p) === 0`); otherwise empty. The `mb_`
    prefix depends only on `S`, not on the wrapper (`strtoupper(mb_substr(…))`
    → `mb_stripos`).
  - `{call}` = `{fn}({arg1}, {O})` where `{arg1}` is the verbatim text of
    `S`'s 1st argument and `{O}` the verbatim text of the other comparison
    operand. If `S` is `mb_substr` **and** has 4 arguments, append
    `, 0, {arg4}`: the encoding goes to `mb_strpos`'s 4th parameter, after
    an explicit offset `0` — `mb_strpos({arg1}, {O}, 0, {arg4})`. A
    4-argument plain `substr` adds nothing. Separators are `, `.
  - `{op}`: the operator text of `B`; if it is 2 characters long (`==`,
    `!=`, and `<>`) a `=` is appended (`===`, `!==`, `<>=`); 3-character
    operators are kept.
  - Regular comparison style (default): `{call} {op} 0`.
    Yoda comparison style: `0 {op} {call}`.
    (The `0` is the text of `S`'s 2nd argument, always `0`.) Single spaces
    around `{op}`.
  The original operand order of `B` does not matter: the produced order is
  driven only by the comparison style.
  Examples (regular):
  - `substr($u, 0, strlen($p)) == $p` → `strpos($u, $p) === 0`
  - `$p !== substr($u, 0, 4)` → `strpos($u, $p) !== 0`
  - `mb_strtolower(substr($u, 0, 4)) != 'get:'` → `stripos($u, 'get:') !== 0`
  - `mb_substr($u, 0, mb_strlen($p, $e), $e) == $p` → `mb_strpos($u, $p, 0, $e) === 0`

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| (none) | | | Operand order of the fix follows the global comparison style (`regular` default / `yoda`). |

## PHP versions
No gating.

## Examples

```php
<?php
function routes($uri, $base, $enc) {
    $r = [];
    $r[] = <weak_warning descr="Use 'strpos($uri, $base) === 0' instead.">substr($uri, 0, strlen($base)) == $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, '/api') !== 0' instead.">'/api' !== substr($uri, 0, 4)</weak_warning>;
    $r[] = <weak_warning descr="Use '\mb_strpos($uri, $base) === 0' instead.">\mb_substr($uri, 0, mb_strlen($base)) === $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_strpos($uri, $base, 0, $enc) !== 0' instead.">mb_substr($uri, 0, mb_strlen($base, $enc), $enc) != $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'stripos($uri, 'get:') === 0' instead.">mb_strtolower(substr($uri, 0, 4)) === 'get:'</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_stripos($uri, 'POST') === 0' instead.">strtoupper(mb_substr($uri, 0, mb_strlen('POST'))) == 'POST'</weak_warning>;

    $r[] = substr($uri, 1, strlen($base)) == $base;
    $r[] = substr($uri, 0, strpos($uri, '?')) == $base;
    $r[] = substr($uri, 0, 3) == $base;
    $r[] = substr($uri, 0, 5) == '/api';
    $r[] = mb_substr($uri, 0, strlen($base)) == $base;
    $r[] = strtoupper(substr($uri, 0, strlen($base))) == $base;
    $r[] = trim(substr($uri, 0, strlen($base))) == $base;
    $r[] = (substr($uri, 0, strlen($base))) == $base;
    $r[] = substr($uri, 0, strlen($base)) < $base;
    $r[] = substr($uri, 0) == $base;
    return $r;
}
```

```php
<?php
function routes($uri, $base, $enc) {
    $r = [];
    $r[] = strpos($uri, $base) === 0;
    $r[] = strpos($uri, '/api') !== 0;
    $r[] = \mb_strpos($uri, $base) === 0;
    $r[] = mb_strpos($uri, $base, 0, $enc) !== 0;
    $r[] = stripos($uri, 'get:') === 0;
    $r[] = mb_stripos($uri, 'POST') === 0;

    $r[] = substr($uri, 1, strlen($base)) == $base;
    $r[] = substr($uri, 0, strpos($uri, '?')) == $base;
    $r[] = substr($uri, 0, 3) == $base;
    $r[] = substr($uri, 0, 5) == '/api';
    $r[] = mb_substr($uri, 0, strlen($base)) == $base;
    $r[] = strtoupper(substr($uri, 0, strlen($base))) == $base;
    $r[] = trim(substr($uri, 0, strlen($base))) == $base;
    $r[] = (substr($uri, 0, strlen($base))) == $base;
    $r[] = substr($uri, 0, strlen($base)) < $base;
    $r[] = substr($uri, 0) == $base;
    return $r;
}
```

Yoda style (`comparisonStyle: yoda`): `substr($uri, 0, strlen($base)) == $base`
becomes `0 === strpos($uri, $base)`.

## Divergences
- **Encoding slot (custos diverges):** for a 4-argument `mb_substr` upstream
  passes the encoding as the 3rd argument of `mb_strpos`/`mb_stripos`, which
  is the offset parameter, so the produced call fails or searches from the
  wrong position. custos emits `mb_strpos({arg1}, {O}, 0, {arg4})` (F1, see
  `docs/decisions.md`).
- **`<>` operator:** upstream appends `=` to any 2-character operator, so
  `<>` would become the invalid `<>=`. Recommendation: map `<>` to `!==`.
  Not covered by upstream fixtures.
- **Length vs. compared value (custos diverges):** upstream accepts any
  non-call length, and any `strlen`/`mb_strlen` call whatever it measures,
  so `substr($s, 0, 3) == $longer` or `substr($s, 0, strlen($a)) == $b` are
  rewritten to a prefix search that answers a different question. custos
  requires the length to be that of `O`, in the substr variant's unit (D7).
- **Case-insensitive mode (custos diverges):** upstream rewrites
  `strtoupper(substr(...)) === $p` to `stripos(...) === 0`, which also
  matches when `$p` has lower-case letters (the original comparison never
  does). custos only reports when `O` is a literal already in the folded
  case (D8), so case-folded comparisons with variables are not reported.
- **Function names (custos diverges):** upstream matches `substr`,
  `mb_substr`, the case-folding wrappers and `strlen`/`mb_strlen` by exact
  spelling regardless of namespace, so `SUBSTR($u, 0, 3) == 'abc'` was
  missed while namespaced user functions with these names were reported and
  rewritten to the global `strpos`. custos matches case-insensitively and only
  calls resolving to the global functions (D1, D5, D7).
- **Builtin spelling (custos diverges).** Upstream copies the written qualifier
  and inserts the new name (`strpos`/`stripos`/`mb_…`) bare, so a function of
  that name declared in or imported into the namespace captures the
  rewritten call. custos writes `\name` in that case (F1 `{qualifier}`).
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
