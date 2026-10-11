---
id: SubStrUsedAsArrayAccess
group: Performance
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# SubStrUsedAsArrayAccess

## Summary

Extracting a single character with `substr($str, $i, 1)` costs a function call;
string offset access (`$str[$i]`) reads the same byte directly. The proven negative
position `-1` is translated to `strlen($str) - 1`.

## Detection

- **D1** A plain function call that resolves to the global function (names
  compared case-insensitively, as PHP does; `\` and a global `use function`
  import are fine, but a same-named function declared in the current
  namespace, one imported from another namespace, or a qualified non-global
  name such as `Ns\substr` does not count): `substr` (`mb_substr` is never
  reported). Call it `S`.
- **D2** `S` has **exactly 3** arguments.
- **D3** The 3rd argument is a number literal whose source text is exactly
  `1` (not `01`, `1.0`, `0x1`, a constant or variable).
- **D4** The 1st argument (`source`) is syntactically one of:
  - a simple variable (`$s`),
  - an array/offset access (`$rows[$k]`, `$m['x'][0]`),
  - a property fetch (instance `$o->name`, nullsafe `$o?->name`, or static
    `Cls::$name`).
  Anything else (calls, casts, literals, concatenations, parenthesized
  expressions) → no report.
- **D5** The inferred type of `source` contains `string` among its known
  members (unknown members are ignored first). So `string`, `?string`,
  `string|int` (any union containing `string`) and docblock-declared
  `string` all qualify; an untyped parameter (unknown type only), `int`,
  `array` do not. (`string[]` counts as `array`, not `string`.)
- **D6** The offset resolves to a known integer at least `-1`. Positive
  and zero offsets retain their written expression; a proven `-1` uses
  `strlen(source) - 1` and requires a fixed-name source variable.
- **D7** An offset written as a variable only counts as known when nothing
  but the visible assignments of its own scope can change it. See E7.

### Safety requirements

The offset must resolve to a known integer at least `-1`; an integer type
alone is insufficient. For `-1`, the source must be a fixed-name variable,
so the generated strlen expression cannot duplicate side effects or
magic property/index reads. Fixes require PHP 8.0 or newer, where an
out-of-range substr returns an empty string; earlier versions report
without a fix.

## Exceptions (no report)

- **E1** `mb_substr`; builtin name casing is otherwise ignored.
- **E2** Length other than literal `1` (`-1`, `2`, `$len`), or a 2/4-argument call.
- **E3** Source expression of another syntactic kind (`substr(trim($s), 0, 1)`,
  `substr((string) $n, 0, 1)`).
- **E4** Source not known to be a string (`int $n`, untyped `$x`).
- **E5** Any other negative offset (`-3`, `-$back`, `-(2 + $k)`): the
  computed position `strlen($s) - n` becomes negative on strings shorter than
  `n`, where offset access counts from the end (PHP ≥ 7.1) or fails, while
  `substr` clamps the start to 0. Only `-1` is eligible for a negative-offset rewrite.
- **E6** Offset not known to be an `int` (untyped, `string`, `int|string`,
  float literal): a non-integer string offset makes `$s[$i]` behave
  differently from `substr`.
- **E7** Offset variable whose assigned value may be stale at the call:
  - a top-level (global-scope) variable,
  - a variable declared `global` or `static` in the enclosing function,
  - a by-reference parameter of the enclosing function,
  - a variable imported by reference into a closure (`use (&$i)`), whether
    the closure is the enclosing scope or nested in it,
  - a variable stored by reference in an array item (`[&$i]`,
    `['k' => &$i]`),
  - a variable that is the operand of `++` or `--` anywhere in the scope.
  A call between the assignment and the `substr` can then move the offset
  below `-1`, so the value read at the assignment proves nothing.

## Report

- Range: the whole call `S` (start of its name including any namespace
  qualifier, to its closing `)`).
- Severity: warning.
- Message: `Use '{replacement}' (string offset access) instead.` where
  `{replacement}` is the R text below (message wording is ours; the hardened
  form may be shown with or without inner padding).

### Replacement text

- `{src}` = verbatim text of the 1st argument, `{off}` = verbatim text of the
  2nd argument.
- **R1** Offset `-1`: `{access}` = `{src}[strlen({src}) - 1]`.
- **R2** Otherwise: `{access}` = `{src}[{off}]`.
- **R3** Level-dependent wrapping:
  - PHP level **8.0 or above**: replacement = `({access} ?? '')` — the
    null-coalescing guard yields `''` for an out-of-range position, as
    `substr` does, instead of raising a warning.
  - PHP level **below 8.0**: substr may return false out of range; the message shows the bare
    `{access}` and no fix is offered.

## Fix

- **F1** (PHP ≥ 8.0 only) Replace `S` with the wrapped replacement text. The
  produced code has **no** padding inside the parentheses: `($s[$i] ?? '')`,
  `($s[strlen($s) - 1] ?? '')`. The inserted `strlen` is written `\strlen`
  when an unqualified call at that position would not reach the global
  function (a `use function` import under that name, or a same-named function declared in the current namespace).
- No fix below PHP 8.0.

## Options

None.

## PHP versions

- The fix exists from PHP 8.0 (R3); below that the rule reports without a
  fix.

## Examples

```php
<?php
function peek(string $buf, ?string $maybe, string|int $mixed, array $rows, int $at)
{
    $a = substr($buf, $at, 1);
    $b = <warning descr="Use '($buf[strlen($buf) - 1] ?? '')' (string offset access) instead.">substr($buf, -1, 1)</warning>;
    $c = <warning descr="Use '($maybe[0] ?? '')' (string offset access) instead.">substr($maybe, 0, 1)</warning>;
    $d = <warning descr="Use '$mixed[2]' (string offset access) instead.">substr($mixed, 2, 1)</warning>;
    return [$a, $b, $c, $d];
}
```

```php
<?php
function peek(string $buf, ?string $maybe, string|int $mixed, array $rows, int $at)
{
    $a = substr($buf, $at, 1);
    $b = ($buf[strlen($buf) - 1] ?? '');
    $c = ($maybe[0] ?? '');
    $d = substr($mixed, 2, 1);
    return [$a, $b, $c, $d];
}
```

## Divergences

- **Audit safety (custos diverges).** Unknown offsets can be negative beyond the string length, where substr
  clamps to the beginning but string offsets do not. Those offsets are skipped.
  The negative-one rewrite is restricted to plain variables because it reads
  the source twice. Before PHP 8.0, out-of-range substr returns false rather
  than an empty string, so the guarded offset rewrite has no automatic fix.

- **Shared offset variables (custos diverges).** An offset variable that
  other code can reach (global, static, by-reference parameter, closure
  import or array item) or that is incremented/decremented is not treated
  as a known value, even right after a literal assignment (E7); upstream
  rewrites such calls regardless.

- **Version gate (custos diverges):** upstream wraps the access in `?? ''`
  only below PHP 7.0, where `??` is a parse error, and emits the bare access
  from 7.0 on, where an out-of-range position raises a warning instead of
  yielding `''`. custos wraps from 8.0 on and offers no fix below 8.0 (R3,
  F1), where substr can return false out of range.
- **Negative offsets (custos diverges):** upstream turns any offset starting
  with `-` into `strlen($s) - n`. On strings shorter than `n` that position
  is negative, which offset access treats differently from `substr` (which
  clamps to the start). custos keeps only `-1` (E5) and stays silent on
  other negative offsets such as `-$offset`.
- **Offset type (custos diverges):** upstream ignores the offset's type; a
  non-integer string offset behaves differently with `$s[$i]`. custos
  requires a proven integer value at least -1 (D6), so uncertain offsets
  are no longer reported.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `substr` by the name as written (case-sensitive, any namespace qualifier, no
  resolution), so a differently cased call such as `SubStr($s, 0, 1)` is
  missed while a namespaced or imported user function of the same name is
  reported (and rewritten) as if it were the builtin. custos matches
  case-insensitively and only calls that reach the global function.
- **Builtin spelling (custos diverges).** Upstream inserts a bare `strlen(` in its fix, so a function of that name declared in (or imported into) the namespace captures the rewritten call. custos writes `\name(` in that case (F1).
- **Non-string sources (custos diverges, R3).** When the source may hold
  another scalar (`string|int $code`), the finding has no fix: an int's
  offset is null, so `($code[$i] ?? '')` would return `''` where
  `substr()` returns the digit.
