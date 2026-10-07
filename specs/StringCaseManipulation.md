---
id: StringCaseManipulation
group: Control flow
kind: syntax
needs: []
php: { min: "", max: "" }
---

# StringCaseManipulation

## Summary
Lower- or upper-casing the haystack and/or needle only to make a substring
search case-insensitive costs extra string copies; PHP already ships
case-insensitive variants of the position functions (`stripos`, `strripos`,
`mb_stripos`, `mb_strripos`). Use them on the raw strings instead.

## Detection
- **D1** A plain function call (not a method or static call) that resolves
  to one of the global "search" functions below, the name compared
  case-insensitively as PHP compares function names (`STRPOS`, `\StrRPos`
  match). It resolves to the global function when written with a single
  leading `\`, or unqualified with no function of that name declared or
  imported (`use function`) in the current namespace; `Ns\strpos(...)` and
  shadowing user functions are not matched:

  | search call   | case-insensitive variant |
  |---------------|--------------------------|
  | `strpos`      | `stripos`                |
  | `mb_strpos`   | `mb_stripos`             |
  | `strrpos`     | `strripos`               |
  | `mb_strrpos`  | `mb_strripos`            |

- **D2** The call has **exactly two** arguments (haystack, needle).
- **D3** For each argument, its *subject* is computed:
  - if the argument is itself a plain function call (not a method/static
    call, not wrapped in parentheses) to one of the global functions
    `strtolower`, `mb_strtolower`, `strtoupper`, `mb_strtoupper` (resolved
    and compared like D1) **and** that call has exactly one argument, the subject is that
    single inner argument;
  - otherwise the argument has no subject (it is kept as is).
- **D4** Report when at least one of the two arguments has a subject (D3).
  Case functions may be mixed: one side lower-cased and the other
  upper-cased still qualifies; only one side being wrapped also qualifies.

## Exceptions (no report)
- **E1** Search calls with one, three or more arguments (e.g. an explicit
  offset as third argument).
- **E2** Case-conversion wrappers with more than one argument (e.g. an
  encoding passed to `mb_strtolower`) are not unwrapped; if neither side
  unwraps, nothing is reported.
- **E3** Method or static calls named like these functions (`$o->strpos(...)`,
  `$o->strtolower(...)`), for either the outer or the inner call.
- **E4** Other case functions (`ucfirst`, `lcfirst`, `ucwords`,
  `mb_convert_case`) are not unwrapped.
- **E5** Names in a different letter case (`STRPOS`) are not matched.

## Report
- Range: the whole outer search call, from the start of its name (including
  any leading `\` / namespace qualifier) to its closing `)`.
- Severity: info (weak warning).
- Message: `Use '{replacement}' instead of changing the case.` where
  `{replacement}` is the F1 text.

## Fix
- **F1** Replace the whole outer call with
  `{variant}({first}, {second})`, where `{variant}` is the unqualified
  case-insensitive function name from the D1 table (any namespace qualifier
  on the original call is dropped), `{first}`/`{second}` are the source text
  of the subject of each argument when it has one, otherwise the argument's
  own source text. Arguments are always joined with `, ` (one comma, one
  space) regardless of the original spacing.
  - `strrpos(mb_strtoupper($t), 'x')` → `strripos($t, 'x')`
  - `\mb_strpos($t, strtolower($n))` → `mb_stripos($t, $n)`
  `{variant}` gets a leading `\` when an unqualified call to it at the
  reported position would not reach the global function (a `use function` import under that name, or a same-named function declared in the current namespace):
  in `namespace S; function stripos() {…}`, `strpos(strtolower($t), $n)` →
  `\stripos($t, $n)`.

## Options
None.

## PHP versions
No gating. (Upstream fixture runs at the IDE test default level, below 7.1;
nothing in the rule is version-dependent.)

## Examples

```php
<?php
function lookup(string $text, string $term, array $row) {
    $a = <weak_warning descr="Use 'stripos($text, $term)' instead of changing the case.">strpos(strtoupper($text), $term)</weak_warning>;
    $b = <weak_warning descr="Use 'strripos($row['k'], 'abc')' instead of changing the case.">strrpos($row['k'], strtolower('abc'))</weak_warning>;
    $c = <weak_warning descr="Use 'mb_stripos($text, $term)' instead of changing the case.">\mb_strpos(mb_strtolower($text), mb_strtoupper($term))</weak_warning>;
    $d = <weak_warning descr="Use 'mb_strripos(trim($text), $term)' instead of changing the case.">mb_strrpos(strtoupper(trim($text)), mb_strtolower($term))</weak_warning>;

    $e = strpos(strtolower($text), $term, 3);
    $f = mb_strpos(mb_strtolower($text, 'UTF-8'), $term);
    $g = strpos(ucfirst($text), $term);
    $h = $row->strpos(strtolower($text), $term);
    $i = strpos($text, $row->strtolower($term));
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}
```

```php
<?php
function lookup(string $text, string $term, array $row) {
    $a = stripos($text, $term);
    $b = strripos($row['k'], 'abc');
    $c = mb_stripos($text, $term);
    $d = mb_strripos(trim($text), $term);

    $e = strpos(strtolower($text), $term, 3);
    $f = mb_strpos(mb_strtolower($text, 'UTF-8'), $term);
    $g = strpos(ucfirst($text), $term);
    $h = $row->strpos(strtolower($text), $term);
    $i = strpos($text, $row->strtolower($term));
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}
```

## Divergences
- **Function names (custos diverges):** upstream matches the bare name
  case-sensitively and ignores namespaces, so `STRPOS(strtolower($t), $n)`
  was missed while a namespaced user `strpos`/`strtolower` (`\App\strpos(...)`,
  or an unqualified call where the namespace declares one) was reported and
  rewritten to the global `stripos`. custos matches case-insensitively and
  only calls resolving to the global functions (D1, D3).
- Upstream does not skip a parenthesised inner call (`strpos((strtolower($t)), $n)`)
  — it simply does not unwrap it. Same here (no change needed).
- **Builtin spelling (custos diverges).** Upstream always inserts the bare
  variant name, which a namespaced or imported function of that name
  captures. custos writes `\variant(` in that case (F1).
