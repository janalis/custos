---
id: IncorrectRandomRange
group: Probable bugs
kind: semantic
needs: [names, index, stubs]
php: { min: "", max: "" }
---

# IncorrectRandomRange

## Summary

`rand()`, `mt_rand()` and `random_int()` expect `(min, max)`. Passing a
minimum greater than the maximum is an error (`random_int()` throws,
`mt_rand()` warns/returns false on older versions, `rand()` silently swaps).
Flag calls whose bounds are known numbers in the wrong order.

## Detection

Visit every function call.

- **D1** The call resolves to the global function `rand`, `mt_rand` or
  `random_int`: names compared case-insensitively, as PHP does (`MT_RAND`
  qualifies); `\rand(...)` and an unqualified call in the global namespace
  match; an unqualified call in a namespace matches only when no function of
  that name is declared in the namespace or imported with `use function`; a
  qualified `Foo\rand(...)` does not match (custos diverges).
- **D2** Exactly two arguments.
- **D3** Run *value discovery* (as defined in the `CallableMethodValidity`
  spec: parentheses stripped, ternary/`??` branches, function-local variable
  assignments and parameter defaults, property defaults, class constants,
  global constants resolved to their declared value; a variable that is also
  incremented/decremented or compound-assigned in its scope makes the result
  unknown) on the first argument;
  continue only when it yields exactly one value `F` that is a *number*.
- **D4** Same for the second argument → exactly one value `T` that is a
  number.
- **D5** A *number* is an integer or float literal token, or a unary minus
  directly applied to such a literal (`-5`; also `- 5` syntactically).
- **D6** Convert `F` and `T` to the integers PHP passes: integer literals
  are parsed as PHP does (decimal, hexadecimal `0x`, octal `0`/`0o`, binary
  `0b`, `_` digit separators), float literals are truncated toward zero
  (`2.9` → `2`), and a unary minus negates the result (whitespace after it
  is irrelevant). A built-in constant resolved from the stubs is parsed from
  its declared decimal value. If either fails (malformed literal, value out
  of the 64-bit range) → stop. Report when `T < F`.

Built-in constants come from the stubs: `PHP_INT_MAX` =
`9223372036854775807`, `PHP_INT_MIN` = `-9223372036854775808` (both
resolvable to numbers, so `random_int(PHP_INT_MAX, PHP_INT_MIN)` is
reported). `T == F` is not reported.

## Exceptions (no report)

- **E1** Other argument counts (`rand()`, `mt_rand(5)`).
- **E2** Bounds that are not a single discoverable number (unknown variables,
  function calls, expressions such as `10 - 1`, ternaries with two different
  numbers, top-level variables), and variables that are also incremented,
  decremented or compound-assigned in their scope (unknown discovery result):
  `$hi = 0; foreach ($xs as $x) { $hi++; } mt_rand(1, $hi)`.
- **E3** Bounds that do not convert to a 64-bit integer (D6), and floats
  that truncate to an ordered or equal pair (`random_int(1.7, 1)`).
- **E4** Correctly ordered or equal bounds.

## Report

- Range: the whole call expression, from the function name (including any
  leading `\`) to the closing `)`.
- Severity: error.
- Message: `Minimum is greater than maximum in this random range.`

## Fix

None.

## Options

None.

## PHP versions

None.

## Examples

```php
<?php
define('DICE_FACES', 6);

class Lottery { const LOW = 1; const HIGH = 49; }

function roll($bonus = 10)
{
    $floor = 100;
    return [
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(9, 3)</error>,
        mt_rand(3, 9),
        <error descr="Minimum is greater than maximum in this random range.">rand(DICE_FACES, 1)</error>,
        rand(1, DICE_FACES),
        <error descr="Minimum is greater than maximum in this random range.">random_int(Lottery::HIGH, Lottery::LOW)</error>,
        <error descr="Minimum is greater than maximum in this random range.">\random_int($floor, -1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">random_int($bonus, 0)</error>,
        random_int(0, PHP_INT_MAX),
        <error descr="Minimum is greater than maximum in this random range.">random_int(0, PHP_INT_MIN)</error>,
        random_int(7, 7),
        <error descr="Minimum is greater than maximum in this random range.">random_int(0x10, 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(2.5, 1)</error>,
        random_int(5, count([1])),
    ];
}
```

## Divergences

- **Function name (custos diverges).** Upstream matches the written name
  case-sensitively and without resolving it, so `MT_RAND(10, 1)` is missed while
  a namespaced user function of the same name (declared in the namespace,
  imported with `use function`, or called qualified) is checked as if it
  were the builtin. custos compares the name case-insensitively and
  requires the call to reach the global function (D1).
- **Unstable variables — custos refinement, not upstream.** Upstream value discovery ignores `++`/`--` and compound assignments, so `$n = 0; … ++$n; mt_rand(1, $n)` is reported as min > max (found on real code). custos treats such a variable as unknown (no report), per the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on it.
- **Non-decimal and float bounds (custos diverges):** upstream only parses
  plain decimal text, so `random_int(0x10, 1)`, `mt_rand(1_000, 1)`,
  `rand(2.5, 1)` or `random_int(- 5, -6)` are silently skipped. custos
  converts every numeric literal the way PHP does (D6) and reports these
  inverted ranges too.
