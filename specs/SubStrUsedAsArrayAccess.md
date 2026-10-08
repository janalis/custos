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
string offset access (`$str[$i]`) reads the same byte directly. Negative
positions are translated to `strlen($str) - n`.

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
- **D6** The 2nd argument (`offset`) is either
  - exactly the literal `-1` (unary minus applied to the integer literal
    `1`), or
  - an expression not starting with `-` whose inferred type is known to be
    `int` only (an integer literal, an `int` parameter, …).

## Exceptions (no report)
- **E1** `mb_substr`, other name casing (`SubStr`).
- **E2** Length other than literal `1` (`-1`, `2`, `$len`), or a 2/4-argument call.
- **E3** Source expression of another syntactic kind (`substr(trim($s), 0, 1)`,
  `substr((string) $n, 0, 1)`).
- **E4** Source not known to be a string (`int $n`, untyped `$x`).
- **E5** Any other negative offset (`-3`, `-$back`, `-(2 + $k)`): the
  computed position `strlen($s) - n` becomes negative on strings shorter than
  `n`, where offset access counts from the end (PHP ≥ 7.1) or fails, while
  `substr` clamps the start to 0. Only `-1` is safe (an empty string gives
  `''` on both sides).
- **E6** Offset not known to be an `int` (untyped, `string`, `int|string`,
  float literal): a non-integer string offset makes `$s[$i]` behave
  differently from `substr`.

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
  - PHP level **7.0 or above**: replacement = `({access} ?? '')` — the
    null-coalescing guard yields `''` for an out-of-range position, as
    `substr` does, instead of raising a warning.
  - PHP level **below 7.0**: `??` does not parse; the message shows the bare
    `{access}` and no fix is offered.

## Fix
- **F1** (PHP ≥ 7.0 only) Replace `S` with the wrapped replacement text. The
  produced code has **no** padding inside the parentheses: `($s[$i] ?? '')`,
  `($s[strlen($s) - 1] ?? '')`. The inserted `strlen` is written `\strlen`
  when an unqualified call at that position would not reach the global
  function (a `use function` import under that name, or a same-named function declared in the current namespace).
- No fix below PHP 7.0.

## Options
None.

## PHP versions
- The fix exists from PHP 7.0 (R3); below that the rule reports without a
  fix.

## Examples
At PHP 7.0+:

```php
<?php
class Token {
    /** @var string */
    public $raw = '';
}
function peek(string $buf, int $at, int $back, string $key, Token $t, $any)
{
    $a = <warning descr="Use '($buf[$at] ?? '')' (string offset access) instead.">substr($buf, $at, 1)</warning>;
    $b = <warning descr="Use '($buf[strlen($buf) - 1] ?? '')' (string offset access) instead.">\substr($buf, -1, 1)</warning>;
    $c = <warning descr="Use '($t->raw[0] ?? '')' (string offset access) instead.">substr($t->raw, 0, 1)</warning>;

    $d = mb_substr($buf, $at, 1);
    $e = substr($buf, $at, 2);
    $f = substr(strrev($buf), 0, 1);
    $g = substr($at, 0, 1);
    $h = substr($any, 0, 1);
    $i = substr($buf, -3, 1);
    $j = substr($buf, -$back, 1);
    $k = substr($buf, $key, 1);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
```

```php
<?php
class Token {
    /** @var string */
    public $raw = '';
}
function peek(string $buf, int $at, int $back, string $key, Token $t, $any)
{
    $a = ($buf[$at] ?? '');
    $b = ($buf[strlen($buf) - 1] ?? '');
    $c = ($t->raw[0] ?? '');

    $d = mb_substr($buf, $at, 1);
    $e = substr($buf, $at, 2);
    $f = substr(strrev($buf), 0, 1);
    $g = substr($at, 0, 1);
    $h = substr($any, 0, 1);
    $i = substr($buf, -3, 1);
    $j = substr($buf, -$back, 1);
    $k = substr($buf, $key, 1);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k];
}
```

Below 7.0: `substr($buf, $at, 1)` is reported as `Use '$buf[$at]' (string
offset access) instead.` with no fix.

## Divergences
- **Version gate (custos diverges):** upstream wraps the access in `?? ''`
  only below PHP 7.0, where `??` is a parse error, and emits the bare access
  from 7.0 on, where an out-of-range position raises a warning instead of
  yielding `''`. custos wraps from 7.0 on and offers no fix below 7.0 (R3,
  F1), so EA cases at the default level lose their fix.
- **Negative offsets (custos diverges):** upstream turns any offset starting
  with `-` into `strlen($s) - n`. On strings shorter than `n` that position
  is negative, which offset access treats differently from `substr` (which
  clamps to the start). custos keeps only `-1` (E5) and stays silent on
  other negative offsets such as `-$offset`.
- **Offset type (custos diverges):** upstream ignores the offset's type; a
  non-integer string offset behaves differently with `$s[$i]`. custos
  requires an offset known to be `int` (D6, E6), so untyped offsets are no
  longer reported.
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
