---
id: ConstantCanBeUsed
group: Language level migration
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ConstantCanBeUsed

## Summary

Some function calls only return a value that PHP already exposes as a
constant (`PHP_VERSION`, `PHP_SAPI`, `M_PI`, …), and comparisons of
`PHP_VERSION` via `version_compare()` or OS sniffing on `PHP_OS` have cheaper,
clearer constant-based equivalents (`PHP_VERSION_ID`, `PHP_OS_FAMILY`).

## Detection

Function names below are compared case-insensitively (`PhpVersion()`
matches), and the call must resolve to the global function: unqualified or
`\`-qualified (`\phpversion()`), not a qualified `Foo\phpversion()`, a
`use function` import from another namespace, or an unqualified call in a
namespace declaring that function. This applies to the D5/D6 calls too. Only
plain function calls are considered (not method/static calls). A name
appearing inside a `use function …;` import is not a call.

### Constant-returning calls (no version gating)

- **D1** A call with **zero** arguments to one of:

  | Call | Constant |
  |---|---|
  | `phpversion()` | `PHP_VERSION` |
  | `php_sapi_name()` | `PHP_SAPI` |
  | `get_class()` | `__CLASS__` |
  | `pi()` | `M_PI` |

  `get_class()` is reported only when the call has a class scope: walking
  up from the call, a class-like (class, anonymous class, interface, trait,
  enum) is reached before any named function declaration; closures and
  arrow functions are transparent. The other calls need no context.

### version_compare (no version gating)

- **D2** A call to `version_compare` with exactly 3 arguments where:
  - argument 1 is a constant reference that resolves to the global
    constant `PHP_VERSION`: unqualified (`PHP_VERSION`, with PHP's global
    fallback: no `use const` import of another constant under that name and
    no `PHP_VERSION` constant declared in the current namespace) or
    `\PHP_VERSION`; a qualified `Foo\PHP_VERSION` does not match;
  - argument 2 is a string literal (single or double quoted) whose raw
    contents are non-empty and fully match `M[.m[.p]]`: `M` exactly one ASCII
    digit, `m` exactly one ASCII digit, `p` one or more ASCII digits (so
    `7`, `8.1`, `7.4.33` match; `10.0`, `7.12`, `7.1.0-dev`, `7.x` do not);
  - argument 3 is a string literal whose raw contents are exactly one of the
    operators below (case-sensitive):

    | Operator(s) | Becomes |
    |---|---|
    | `<`, `lt` | `<` |
    | `<=`, `le` | `<=` |
    | `>`, `gt` | `>` |
    | `>=`, `ge` | `>=` |
    | `==`, `=`, `eq` | `===` |
    | `!=`, `<>`, `ne` | `!==` |

- **D3** The version number `V` is built as: `M`, then `m` (missing → `0`)
  left-padded to 2 digits with one `0` if it has a single digit, then `p`
  (missing → `0`) left-padded with one `0` if it has a single digit (a
  multi-digit patch is used as-is). Examples: `7` → `70000`, `8.1` → `80100`,
  `7.4.33` → `70433`, `5.6.4` → `50604`, `7.0.123` → `700123`.
- **D4** Suggested expression `E` = `PHP_VERSION_ID <op> V` (single spaces),
  with `<op>` adjusted by D4b.
- **D4c** Constant spelling: the constants inserted by D1 (`PHP_VERSION`,
  `PHP_SAPI`, `M_PI`) and D4 (`PHP_VERSION_ID`) are written with a leading
  `\` when a bare name at the call's position would not reach the global
  constant (a `use const` import under that name, or a constant of that name
  declared in the current namespace). The message uses the same spelling.
- **D4b** For a one- or two-part version (`'8'`, `'7.1'`), `<op>` `>`
  becomes `>=` and `<=` becomes `<` (`<` and `>=` are unchanged):
  `version_compare(PHP_VERSION, '7.1', '>')` → `PHP_VERSION_ID >= 70100`,
  `version_compare(PHP_VERSION, '7.1', '<=')` → `PHP_VERSION_ID < 70100`.
  `version_compare()` sorts a short version below its `.0` release
  (`'7.1'` < `'7.1.0'`), and `PHP_VERSION` always has three parts, so
  `PHP_VERSION > '7.1'` already holds on 7.1.0. Full `M.m.p` versions keep
  the operator from the table.
- **D4a** For the equality operators (`==`, `=`, `eq`, `!=`, `<>`, `ne`)
  the version must have all three parts (`M.m.p`); `'8.1'` or `'8'` with an
  equality operator is not reported.

### PHP_OS sniffing (PHP ≥ 7.2)

- **D5** A constant reference resolving to the global constant `PHP_OS`
  (same resolution as for `PHP_VERSION` in D2; `Foo\PHP_OS` does not match) that
  is a **direct argument** (any position, not parenthesised or nested in
  another expression) of a function call `F` named one of: `strpos`,
  `stripos`, `mb_strpos`, `mb_stripos`, `strncmp`, `strncasecmp`, `substr`,
  `mb_substr`.
- **D6** The *context* node `X` is `F`; except that when `F` is `substr` or
  `mb_substr` and `F` is itself a direct argument (any position) of a function
  call named `strtolower`, `mb_strtolower`, `strtoupper` or `mb_strtoupper`,
  `X` is that outer call.
- **D7** `X`'s direct parent must be a binary expression with one of the
  operators `==`, `!=`, `===`, `!==` (not `<>`, not ordering operators), `X`
  being either operand (parentheses around `X` break this).
- **D8** The other operand must be:
  - for `substr` / `mb_substr` (wrapped or not): a string literal;
  - for the other functions: an integer/float number literal, a negated number
    literal (`-1`), or the constant `false` (any letter case).

## Exceptions (no report)

- **E1** D1 calls with any argument (`phpversion('mysqli')`,
  `phpversion($ext)`); functions not in the table (`php_uname()` etc.).
- **E4** `get_class()` without a class scope (file level, plain functions,
  a function declared inside a method).
- **E5** `version_compare` with an equality operator and a one- or two-part
  version (`version_compare(PHP_VERSION, '8.1', '==')`).
- **E2** `version_compare` with 2 or 4 arguments, with a first argument other
  than the bare `PHP_VERSION` constant, with a non-literal version or
  operator, an empty version, a version not matching D2 (suffixes like
  `-dev`, two-digit major/minor), or an unknown operator string.
- **E3** D5–D8 below PHP 7.2; `PHP_OS` not a direct argument; comparisons
  against variables, other constants, `true`, `null`; ordering comparisons;
  substr compared to a number; strpos compared to a string.

## Report

- Range:
  - D1, D2: the whole call (name through closing `)`).
  - D5: the context node `X` only (the function call, or the wrapping case
    call), not the comparison.
- Severity: info (weak warning).
- Message:
  - D1: `Use the {CONST} constant instead of this call.`
  - D2: `Replace with '{E}'.`
  - D5: `Compare PHP_OS_FAMILY instead of sniffing PHP_OS.`

## Fix

- **F1** D1: replace the call with the constant name (`PHP_VERSION`,
  `PHP_SAPI`, `__CLASS__`, `M_PI`).
- **F2** D2: replace the whole `version_compare(...)` call with `E`, e.g.
  `version_compare(PHP_VERSION, '7.3', '<')` → `PHP_VERSION_ID < 70300`.
  Upstream inserts `E` without parentheses (see Divergences).
- D5: no fix.

## Options

None.

## PHP versions

- D1/D2: no gating (fixture runs at the test default level, below 7.1).
- D5–D8: only when the configured PHP level is ≥ 7.2 (`PHP_OS_FAMILY`
  availability); fixture runs at 7.3.

## Examples

```php
<?php
use function pi;

$sapi = <weak_warning descr="Use the PHP_SAPI constant instead of this call.">php_sapi_name()</weak_warning>;
$ver  = <weak_warning descr="Use the PHP_VERSION constant instead of this call.">\phpversion()</weak_warning>;
$self = get_class();
$area = 2 * <weak_warning descr="Use the M_PI constant instead of this call.">pi()</weak_warning> * $r;

$ext  = phpversion('intl');
$host = php_uname('n');

$a = <weak_warning descr="Replace with 'PHP_VERSION_ID < 80000'.">version_compare(PHP_VERSION, '8', 'lt')</weak_warning>;
$b = <weak_warning descr="Replace with 'PHP_VERSION_ID >= 70433'.">version_compare(PHP_VERSION, "7.4.33", 'ge')</weak_warning>;
$c = <weak_warning descr="Replace with 'PHP_VERSION_ID === 80200'.">version_compare(PHP_VERSION, '8.2.0', 'eq')</weak_warning>;
$d = <weak_warning descr="Replace with 'PHP_VERSION_ID !== 50604'.">version_compare(PHP_VERSION, '5.6.4', '<>')</weak_warning>;

$e = version_compare(PHP_VERSION, '8.2.0-RC1', '>=');
$f = version_compare(PHP_VERSION, '10.1', '>=');
$g = version_compare(PHP_VERSION, '8.2');
$h = version_compare($other, '8.2', '>=');
$i = version_compare(PHP_VERSION, '8.2', 'GE');
$j = version_compare(PHP_VERSION, '8.2', '==');

class Widget {
    public function name() {
        return <weak_warning descr="Use the __CLASS__ constant instead of this call.">get_class()</weak_warning>;
    }
}
```

```php
<?php
use function pi;

$sapi = PHP_SAPI;
$ver  = PHP_VERSION;
$self = get_class();
$area = 2 * M_PI * $r;

$ext  = phpversion('intl');
$host = php_uname('n');

$a = PHP_VERSION_ID < 80000;
$b = PHP_VERSION_ID >= 70433;
$c = PHP_VERSION_ID === 80200;
$d = PHP_VERSION_ID !== 50604;

$e = version_compare(PHP_VERSION, '8.2.0-RC1', '>=');
$f = version_compare(PHP_VERSION, '10.1', '>=');
$g = version_compare(PHP_VERSION, '8.2');
$h = version_compare($other, '8.2', '>=');
$i = version_compare(PHP_VERSION, '8.2', 'GE');
$j = version_compare(PHP_VERSION, '8.2', '==');

class Widget {
    public function name() {
        return __CLASS__;
    }
}
```

PHP_OS sniffing (PHP level 7.4, no fix):

```php
<?php
function os_checks() {
    $win  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">stripos(PHP_OS, 'win')</weak_warning> === 0;
    $dar  = false !== <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">mb_strpos(PHP_OS, 'Darwin')</weak_warning>;
    $lin  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">strncasecmp(PHP_OS, 'LIN', 3)</weak_warning> == 0;
    $bsd  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">mb_strtoupper(mb_substr(PHP_OS, 0, 3))</weak_warning> != 'BSD';
    $sun  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">substr(PHP_OS, 0, 3)</weak_warning> === 'Sun';

    // not reported
    $p1 = (strpos(PHP_OS, 'WIN')) === 0;
    $p2 = strpos(PHP_OS, 'WIN') > 0;
    $p3 = strpos(PHP_OS, 'WIN') === $offset;
    $p4 = substr(PHP_OS, 0, 3) === $prefix;
    $p5 = strpos(strtoupper(PHP_OS), 'WIN') === 0;
    $p6 = str_starts_with(PHP_OS, 'WIN') === true;
}
```

## Divergences

- F2 inserts `PHP_VERSION_ID <op> V` without parentheses, so in a
  higher-precedence context (`!version_compare(PHP_VERSION, '8', '<')`,
  `version_compare(...) . 'x'`) the text reparses differently
  (`!PHP_VERSION_ID < 80000`). custos wraps `E` in parentheses when the
  call's parent is an operator binding tighter than comparisons (unary `!`,
  arithmetic, concatenation, `instanceof`, casts, …). Statement/argument/
  assignment contexts (the only ones in fixtures) need none.
- custos diverges from upstream on short-version equality (D4a, E5).
  `version_compare()` orders `8.1` before `8.1.0`, so
  `version_compare(PHP_VERSION, '8.1', '==')` is false on every real PHP
  version, while upstream's `PHP_VERSION_ID === 80100` is true on 8.1.0.
  custos rewrites equality comparisons only for full `M.m.p` versions,
  where both forms agree.
- custos diverges from upstream on `get_class()` scope (D1, E4). Outside a
  class `__CLASS__` is the empty string, so the rewrite would change the
  value; custos reports `get_class()` only where it has a class scope.
- custos diverges from upstream on short-version ordering (D4b). Upstream
  maps `>`/`<=` unchanged, so `version_compare(PHP_VERSION, '7.1', '>')`
  (true on 7.1.0) becomes `PHP_VERSION_ID > 70100` (false on 7.1.0), and
  `'<='` (false on 7.1.0) becomes `<= 70100` (true on 7.1.0). custos emits
  `>= V` / `< V` for one- and two-part versions.
- The first two divergences affect the main EA case, which is listed in
  `testdata/ea-divergences.json`.
- Upstream's `$`-anchored version match would also accept a version string
  ending with a single trailing newline (`"8.1\n"` raw contents never contain
  a real newline in practice). Recommendation: require a full match without
  trailing characters.
- **Function names (custos diverges):** upstream compares the written last
  segment case-sensitively: it misses `PHP_SAPI_NAME()`, `Version_Compare(...)`
  or `StrPos(PHP_OS, ...)`, and rewrites calls to user functions such as a
  namespaced `phpversion()` (whose result is not `PHP_VERSION`). custos
  matches any case and requires every matched call to resolve to the global
  function. Constant names (`PHP_VERSION`, `PHP_OS`) stay case-sensitive.
- **Namespaced constants (custos diverges).** Upstream compares only the
  last segment of the constant name, so `version_compare(Foo\PHP_VERSION,
  …)` and `strpos(Lib\PHP_OS, 'WIN')` are reported although they read user
  constants. custos requires the reference to resolve to the global constant
  (D2, D5). It also writes the inserted constants as `\NAME` when a
  namespaced constant would capture the bare name (D4c); upstream always
  writes them bare.
