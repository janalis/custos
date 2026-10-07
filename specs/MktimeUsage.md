---
id: MktimeUsage
group: Compatibility
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# MktimeUsage

## Summary
Two legacy `mktime()` / `gmmktime()` usages:
calling them without arguments just returns the current timestamp (and
emits an `E_STRICT`/deprecation notice on PHP 5.x/7.x) — `time()` is the
intended call; and the seventh `$is_dst` argument was deprecated in PHP 5.1
and removed in PHP 7.0.

## Detection
Visit every plain function call (not method/static calls).

- **D1** The called name — last segment, compared case-insensitively as PHP
  does — is `mktime` or `gmmktime`.
- **D2** The call resolves to the **global** built-in function of that name:
  unqualified in the global namespace, `\`-qualified, imported with
  `use function mktime;`, or unqualified inside a namespace where it falls
  back to the global function. A call resolving to a same-named
  namespaced/user function (`use function Lib\mktime;`, or `function
  mktime()` declared in the current namespace), or one that cannot be
  resolved, is not reported.
- **D3** The name reference inside a `use function ...;` import statement is
  never treated as a call.

### No arguments
- **D4** Zero arguments: report the call (pattern A).

### `is_dst` argument
- **D5** Exactly seven arguments: report the seventh argument (pattern B),
  whatever expression it is (literal, unary `-1`, variable, constant, call…).

## Exceptions (no report)
- **E1** One to six arguments, or eight or more.
- **E2** Calls not resolving to the global function (D2).
- **E3** Other spellings (`MkTime()`), method calls `$x->mktime()`, static
  calls `X::mktime()`.

## Report
Pattern A
- Range: the whole call, from the function name (including a leading `\` or
  namespace qualifier as written) to the closing `)`.
- Severity: warning.
- Message: `Call time() instead; mktime()/gmmktime() without arguments is deprecated.`

Pattern B
- Range: the seventh argument expression exactly (e.g. `-1` including the
  minus sign).
- Severity: warning (rendered as deprecated/strikethrough).
- Message: `The is_dst argument is deprecated and was removed in PHP 7.0.`

## Fix
- **F1** Pattern A only: replace the whole call expression (including any
  leading `\` / qualifier) with `time()`.
  - `mktime()` → `time()`, `\gmmktime()` → `time()`, `gmmktime( )` → `time()`.
  Builtin spelling: `time` is written `\time` when an unqualified call at
  that position would not reach the global function (a `use function`
  import under that name, or a same-named function declared in the current
  namespace): `namespace Clock; function time() {…} mktime();` → `\time()`.
- Pattern B has no fix.

## Options
None.

## PHP versions
No gating: both patterns are reported at every language level (upstream
does not check the configured version; the `is_dst` report applies
regardless of whether the target is below 7.0).

## Examples

```php
<?php
namespace Billing;

use function gmmktime;

function epoch_values($flag)
{
    $now   = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">mktime()</warning>;
    $utc   = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">gmmktime()</warning>;
    $root  = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">\mktime()</warning>;
    $dst   = mktime(12, 0, 0, 6, 1, 2010, <warning descr="The is_dst argument is deprecated and was removed in PHP 7.0.">$flag</warning>);
    $dst2  = gmmktime(1, 2, 3, 4, 5, 2001, <warning descr="The is_dst argument is deprecated and was removed in PHP 7.0.">0</warning>);
    $noon  = mktime(12);
    $full  = gmmktime(1, 2, 3, 4, 5, 2001);
}
```

```php
<?php
namespace Billing;

use function gmmktime;

function epoch_values($flag)
{
    $now   = time();
    $utc   = time();
    $root  = time();
    $dst   = mktime(12, 0, 0, 6, 1, 2010, $flag);
    $dst2  = gmmktime(1, 2, 3, 4, 5, 2001, 0);
    $noon  = mktime(12);
    $full  = gmmktime(1, 2, 3, 4, 5, 2001);
}
```

## Divergences
- **Case of the name (custos diverges from upstream).** Upstream compares
  the written name case-sensitively, so `MKTIME()` is not reported. custos
  matches the name in any case (D1).
- The `is_dst` report could be gated on target `< 7.0` being irrelevant
  (on 7.0+ the call is a fatal arity error anyway); upstream reports at all
  levels. Recommendation: keep reporting at all levels (matches fixtures).
- **Builtin spelling (custos diverges).** Upstream inserts a bare `time(`,
  which a function of that name declared in (or imported into) the current
  namespace captures, so the fix would call user code. custos writes
  `\time(` in that case.
