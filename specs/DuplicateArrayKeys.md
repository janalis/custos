---
id: DuplicateArrayKeys
group: Probable bugs
kind: syntax
needs: []
php: { min: "", max: "" }
---

# DuplicateArrayKeys

## Summary
In an array literal a repeated string key silently overwrites the earlier
entry. Either the earlier entry is dead (different value) or the whole entry
is a redundant copy (same value).

## Detection
Visit every array literal (`[...]` and `array(...)`, including short-list
destructuring targets, which share the same syntax).

Walk its `key => value` elements in source order, keeping a map from the
normalised key to the **most recently seen** value for that key. Elements
without an explicit key and spread elements are skipped.

- **D1** Only literal keys take part, normalised to the key PHP stores:
  - string literals (single- or double-quoted) **without interpolation** are
    decoded with PHP's escape rules for their quote style (`'id'`, `"id"`
    and `"\x69d"` are the same key; so are `'a\'b'` and `"a'b"`);
  - a decoded string that is a canonical decimal integer (no leading zero,
    `+` or whitespace, fits a 64-bit integer, not `"-0"`) becomes that
    integer key (`'7'` and `7` are the same key; `'07'` and `'1.0'` stay
    strings);
  - integer literals in any base, optionally negated (`0x1F`, `0b11`,
    `-3`), are integer keys.
- **D2** If the key text was already seen in this array literal:
  - **D2a (duplicate pair)** if the current value is **not** an array literal
    and is *equivalent* to the most recently stored value for that key →
    report the whole element (pair variant);
  - **D2b (duplicate key)** otherwise (different value, or the value is an
    array literal — array values are never considered equal, not even
    `[] => []`) → report the key (key variant).
- **D3** After each keyed string-literal element, store/overwrite the map
  entry with the current value (so a third occurrence is compared with the
  second one, not the first).

*Equivalent* values: same node kind and structurally identical, ignoring
whitespace and comments (e.g. `$a+1` vs `$a + 1`), or identical source text.
Two variables are equivalent when their names are equal. Different literal
spellings are not equivalent (`'x'` vs `"x"`, `1` vs `1.0`, `0x1` vs `1`).
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** Non-literal keys and floats: constants, class constants,
  variables, expressions, float literals (`K => …, K => …` is not reported).
- **E2** Interpolated string keys (`"k$n"`).
- **E3** The first occurrence of any key.
- **E4** Duplicates across different (nested) array literals.

## Report
- Range:
  - D2a: the whole element from the start of the key to the end of the
    value (`'k' => 'v'`, no trailing comma).
  - D2b: the key literal only, including quotes.
- Severity: warning for both (the pair variant is additionally rendered as
  "unused" in editors; this is presentation only).
- Messages:
  - D2a: `Same key and value already present; remove this entry.`
  - D2b: `Key already used earlier; the earlier entry is overwritten.`

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
$settings = [
    7 => 'a',
    <warning descr="Same key and value already present; remove this entry.">'7' => 'a'</warning>,
    'retries' => 3,
    <warning descr="Same key and value already present; remove this entry.">"retries" => 3</warning>,
    'hosts' => ['db1'],
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'hosts'</warning> => ['db1'],
    'timeout' => 10,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'timeout'</warning> => 20,
    <warning descr="Key already used earlier; the earlier entry is overwritten.">'timeout'</warning> => 10,
    'label' => $prefix . 'x',
    <warning descr="Same key and value already present; remove this entry.">'label' => $prefix.'x'</warning>,
    "slot-$n" => 1,
    "slot-$n" => 1,
    'mode' => 'fast',
    'nested' => ['mode' => 'slow'],
];
```

## Divergences
- custos diverges: upstream compares the raw text between the quotes and
  ignores integer keys, so keys PHP stores identically but spelled
  differently (`"\x41"` vs `'A'`, `'1'` vs `1`, `0x10` vs `16`, or simply
  `3 => …, 3 => …`) are missed. custos compares the normalised key value
  (D1), so these duplicates are reported too.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
