---
id: TypeUnsafeArraySearch
group: Type compatibility
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# TypeUnsafeArraySearch

## Summary

`in_array()` and `array_search()` compare loosely unless their third argument
is `true`, so `'1abc'`, `1`, `true` and `'1'` can all "match". Asking for the
third argument forces the author to decide whether loose matching is really
wanted.

## Detection

Visit every function call.

- **D1** The called name (last name segment, compared case-insensitively as
  PHP does) is `in_array` or `array_search`, and the call resolves to that **global**
  function under PHP's rules (imports, a function of the same name in the
  current namespace first, then the global fallback). `App\in_array(...)`,
  or an unqualified `in_array(...)` inside `namespace App;` where
  `App\in_array` is declared, is not matched.
- **D2** The call has exactly **two** arguments (named or positional; a third
  argument of any value, `true` or `false`, ends the check), neither of them
  a spread (`...$args`): a spread supplies an unknown number of arguments and
  may already carry the strict flag. Let `N` be argument 0 (needle) and `H`
  argument 1 (haystack), as written.
- **D3** Not excluded by E1/E2 → report.

## Exceptions (no report)

- **E1** *Literal string haystack*: `H` is an array literal (`[...]` or
  `array(...)`, not wrapped in parentheses) with at least one element, and
  **every** element is a plain value element (no `key =>`, no spread) whose
  value is a class name written `Name::class` (custos, see Divergences) or
  a string literal (any quoting) whose contents, after trimming
  surrounding whitespace, are non-empty and are not a numeric string (PHP's
  grammar: optional sign, digits with an optional decimal part or a bare
  `.5` fraction, optional exponent — `10`, `-1`, `1.5`, `.5`, `1e3`, `+2`).
  Contents are taken raw (escape sequences not decoded).
  - `['draft', 'final']` → skipped.
  - `['10', '20']`, `['-1', '1.5']` (numeric), `['']` / `[' ']` (empty after trim),
    `['k' => 'v']` (keyed), `['a', 5]` (non-string), `['a', $x]` → not
    skipped → reported.
  - An empty array literal `[]` is not skipped by E1.
  - (custos) An item may also be a class constant fetched through a class
    name (`self::USER`, `Webform::PAGE`, not `static::`, which a subclass
    may redefine) whose initializer is such a string literal.
- **E2** *Matching types*: the inferred type of `N` (unknown parts removed)
  consists of exactly one type `T`, the inferred type of `H` (unknown parts
  removed) consists of exactly one type `A`, and `A` equals `T` with a leading
  `\` ensured plus `[]` (e.g. `T = \int`, `A = \int[]`; `T = \Order`,
  `A = \Order[]`). Types come from declarations, `@var` annotations (including
  inline `/* @var type $name */` and `/** @var type $name */` comments placed
  before the statement; the nearest preceding annotation for a variable
  applies), parameter/return types, and literal inference (an array literal
  of integers is `\int[]`, an empty array literal is `\array`, a string
  literal is `\string`).
  - needle `@var int`, haystack `@var int[]` → skipped.
  - needle `@var int`, haystack `@var array` → reported.
  - (custos) `A` may also be an array whose keys and values are all known,
    with at least one entry, every value of type `T` alone: a class
    constant `const MODES = ['HTML', 'Markdown'];` searched with a string
    needle → skipped; `const MIXED = ['HTML', 1];` → reported.
  - needle `'1'` (`\string`), haystack `[]` (`\array`) → reported.
- **E3** Calls with 0, 1 or 3+ arguments, and calls where a written argument
  is a spread (`in_array($needle, ...$rest)`, `in_array(...$args)`,
  `in_array(...$a, ...$b)`).
- **E4** Calls resolving to a user function that is not the global one
  (`\App\in_array($a, $b)`).

## Report

- Range: the whole call expression, from the start of the function name
  (including any leading `\` or qualifier) to the closing `)`.
- Severity: info (weak warning).
- Message: `Pass a third argument to say whether this search must be type-strict.`

## Fix

No automatic fix: adding the strict flag can change the result (see
Divergences).

## Options

None.

## PHP versions

None. Upstream fixture runs at the PhpStorm test default level (5.6–7.0).

## Examples

```php
<?php
function pick($code, array $allowed) {
    $r = [];
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, $allowed)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\array_search(7, [])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">array_search($code, ['404', '500'])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['  '])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['x' => 'yes'])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['on', 1])</weak_warning>;

    /** @var string $color */
    /** @var array $palette */
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($color, $palette)</weak_warning>;

    /** @var string $tone */
    /** @var string[] $tones */
    $r[] = in_array($tone, $tones);

    $r[] = in_array($code, ['draft', "final"]);
    $r[] = array_search($code, array('v1', 'v2'));
    $r[] = in_array($code, $allowed, false);
    $r[] = array_search($code, $allowed, true);
    $r[] = in_array($code);
    return $r;
}
```

```php
<?php
function pick($code, array $allowed) {
    $r = [];
    $r[] = in_array($code, $allowed, true);
    $r[] = \array_search(7, [], true);
    $r[] = array_search($code, ['404', '500'], true);
    $r[] = in_array($code, ['  '], true);
    $r[] = in_array($code, ['x' => 'yes'], true);
    $r[] = in_array($code, ['on', 1], true);

    /** @var string $color */
    /** @var array $palette */
    $r[] = in_array($color, $palette, true);

    /** @var string $tone */
    /** @var string[] $tones */
    $r[] = in_array($tone, $tones);

    $r[] = in_array($code, ['draft', "final"]);
    $r[] = array_search($code, array('v1', 'v2'));
    $r[] = in_array($code, $allowed, false);
    $r[] = array_search($code, $allowed, true);
    $r[] = in_array($code);
    return $r;
}
```

## Divergences

- **Constant arrays of one type (custos diverges, E2).** Upstream types a
  class constant `['HTML', 'Markdown']` as a plain array, so a string
  needle searched in it is reported. Every value is a string like the
  needle, so loose and strict comparison agree: custos does not report it
  (monolog, pheanstalk). A static property a trait declares for itself,
  read as `self::$values`, also keeps its declared type there.
- **Class names (custos diverges, E1).** Upstream only accepts quoted
  strings in a literal haystack, so `in_array($type, ['string',
  Header::class])` is reported. A `Name::class` is a class name, never a
  numeric string: a loose search compares it as a strict one does.

- **Case of the name (custos diverges from upstream).** Upstream compares
  the written name case-sensitively, so `In_Array($a, $b)` is not reported.
  custos matches any case; the fix keeps the name as written.
- **D1 — custos diverges from upstream.** Upstream matches the last name
  segment only, so a user function `App\in_array($a, $b)` with its own
  semantics is reported and "fixed" with a third argument it may not accept.
  custos resolves the call and matches only the global functions.
- **Spread arguments (custos diverges from upstream).** Two written
  arguments are not two passed arguments when one is `...$rest`:
  `in_array($needle, ...$rest)` called with `[[1, 2], true]` is already
  strict. custos does not report a call with a spread argument.
- E2 depends on PhpStorm-style type inference (inline `@var` comments,
  array-literal element typing). Where custos cannot infer a type, the call is
  reported (safe side). The upstream fixture only needs: inline `@var` with
  `int`/`int[]`/`array`, literal `'1'` vs `[]`.
- **Numeric strings in E1 (custos diverges).** Upstream only treats
  digit-only strings as numeric, so a haystack such as `['-1', '1.5']` or
  `['1e3']` is skipped as "plain strings" although PHP compares numeric
  strings by value under loose matching (`'1.50'` finds `'1.5'`). custos
  uses the full numeric-string grammar and reports such haystacks.
- **No fix (custos diverges).** Adding `true` changes the result whenever
  loose and strict equality differ (`'1' == 1`, `null == ''`,
  `'1e1' == '10'`): PrestaShop searched a configuration string among int
  constants (`in_array(Configuration::get('PS_ROUND_TYPE'), [Order::ROUND_ITEM, …])`),
  which the fix would have made always false. The cases known to be
  equivalent (same type on both sides) are already exempt (E2), so the
  rule reports without a fix. EA case affected: `strict-array-search.php`
  (listed divergence).

Recorded in `internal/testing/conformance/testdata/ea-divergences.json`.
