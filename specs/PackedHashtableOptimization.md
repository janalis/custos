---
id: PackedHashtableOptimization
group: Performance
kind: syntax
needs: []
php: { min: "7.0", max: "" }
---

# PackedHashtableOptimization

## Summary

Since PHP 7 an array whose keys are integers inserted in ascending order can
be stored in a compact "packed" layout. An array literal that lists integer
keys out of order, or spells integer keys as numeric strings, misses that
layout. Sorting the keys or writing them as integers lets the engine use it.

## Detection

- **D1** Project PHP level is **7.0 or higher**.
- **D2** An array creation expression (`[...]` or `array(...)`) with **at
  least 3** elements. (List/destructuring targets are not array creations.)
- **D3** Not in a test context: the file path does not end with `Test.php`,
  `Spec.php` or `.phpt` and does not contain `/Fixtures/`; and the innermost
  enclosing class (if any) has an FQN that does not end with `Test` and does
  not contain `\Tests\` or `\Test\`.
- **D4** **Every** element is a `key => value` pair whose key is either
  - a string literal without interpolation (single or double quoted), or
  - a number literal, or a unary minus applied directly to a number literal
    (`-3`).
  Any other element (value-only element, spread, variable/constant/expression
  key) → no report for the whole array.
- **D5** Convert each key, in source order, to a 64-bit signed integer (the
  integer key PHP uses on 64-bit platforms):
  - string key: take its raw content (escapes not decoded). If the content is
    longer than 1 character and starts with `0` (`'00'`, `'01'`, `'0x1'`) →
    no report for the whole array. Otherwise it must be a canonical decimal
    integer (see Divergences for the sign) within
    −9223372036854775808…9223372036854775807; anything else (`'a'`, `'1.5'`,
    `' 1'`, `''`, a value beyond the 64-bit range) → no report for the whole
    array.
  - number key: an integer literal evaluated as PHP does — decimal, hex
    (`0x1F`), octal (`017`, `0o17`), binary (`0b101`), with `_` digit
    separators — negated when written under a unary minus (`- 3` → −3).
    Float literals and integer literals that overflow the 64-bit range (PHP
    makes them floats) → no report.
- **D6** Let *ascending* be true unless some key is strictly smaller than the
  key immediately before it (equal neighbours are fine). Let *has-string* be
  true when at least one key is a string literal.
  - **D6a** not ascending → report kind R (reorder).
  - **D6b** ascending and has-string → report kind I (use integer keys).
  - ascending without string keys → nothing.

## Exceptions (no report)

- **E1** PHP level below 7.0.
- **E2** Fewer than 3 elements.
- **E3** Any element without a key, or with a non-literal key.
- **E4** A non-integer-like string key (`'name'`), or one with a leading zero
  (`'007'`), anywhere in the array.
- **E5** Test context (D3).

## Report

- Range: the first token of the array creation only — the `[` of a short
  array, or the `array` keyword of a long one (not the parenthesis).
- Severity: info (weak warning).
- Messages:
  - R: `Sort the integer keys ascending so the array can be stored packed.`
  - I: `Write the keys as integers so the array can be stored packed.`

## Fix

None.

## Options

None.

## PHP versions

Reported only at PHP ≥ 7.0. The EA fixture runs at 7.1.

## Examples

```php
<?php
$grid = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>
    5 => 'e', 3 => 'c', 9 => 'i'
];
$cells = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">array</weak_warning>(
    -1 => 'a', -4 => 'b', 0 => 'c', 2 => 'd'
);
$names = <weak_warning descr="Write the keys as integers so the array can be stored packed.">[</weak_warning>
    '3' => 'x', 4 => 'y', '8' => 'z'
];
$mixed = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">array</weak_warning>('7' => 1, '2' => 2, 9 => 3);

$small = [2 => 'b', 1 => 'a'];
$sorted = [1 => 'a', 2 => 'b', 2 => 'c'];
$labels = ['3' => 'x', 'label' => 'y', 5 => 'z'];
$padded = ['07' => 'x', 8 => 'y', 9 => 'z'];
$dyn = [$k => 1, 4 => 2, 6 => 3];
$plain = ['a', 'b', 'c'];
$float = [3.5 => 'a', 2 => 'b', 1 => 'c'];
```

Integer literals in any base count, and so do 64-bit keys:

```php
<?php
$flags = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>0x10 => 'a', 0b1 => 'b', 0o7 => 'c'];
$ids = <weak_warning descr="Sort the integer keys ascending so the array can be stored packed.">[</weak_warning>8000000000 => 'a', 7000000000 => 'b', 9000000000 => 'c'];
```

## Divergences

- **Sign handling in string keys:** upstream's integer parse accepts a
  leading `+` (and `-0`), but PHP keeps `'+5'` and `'-0'` as string keys, so
  such arrays are mis-classified (e.g. `['+5' => …, 6 => …, 7 => …]` gets
  report I). Recommendation: accept only canonical decimal strings (`0` or
  an optional `-` followed by a non-zero digit and digits, no `-0`). No
  fixture covers it.
- **64-bit keys (custos diverges from upstream).** Upstream parses keys as
  32-bit integers, so any key beyond ±2147483647 stops the analysis
  although PHP on 64-bit platforms stores it as an integer key. custos uses
  the 64-bit range (D5).
- **Hex/octal/binary number keys (custos diverges from upstream).**
  Upstream parses the key text as a decimal number, so `0x10`, `017`,
  `0b1` or `1_000` stop the analysis. custos evaluates PHP integer literals
  (D5), so out-of-order keys written in those forms are reported.
- **Element count:** upstream counts the array's child nodes; comments inside
  the literal are not expected to count. Recommendation: count real elements
  only.
