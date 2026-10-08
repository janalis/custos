---
id: IncrementDecrementOperationEquivalent
group: Code style
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# IncrementDecrementOperationEquivalent

## Summary

Adding or subtracting the literal `1` to a variable and storing it back
(`$n += 1`, `$n = $n - 1`, …) is more idiomatically written with the increment
or decrement operator.

## Detection

Let *T* be the assignment target (left-hand side) and "the literal `1`" mean an
operand whose exact source text is `1` (not `1.0`, `0x1`, `(1)`, `+1`, `true`).

D1. Compound assignment `T += 1` → increment candidate.
D2. Compound assignment `T -= 1` → decrement candidate.
D3. Plain assignment `T = A + B` (the right-hand side is directly a binary `+`
    expression, not parenthesized) where either
    `A` is the literal `1` and `B` is equivalent to `T`, or
    `B` is the literal `1` and `A` is equivalent to `T` → increment candidate.
D4. Plain assignment `T = A - B` (directly a binary `-`) where `B` is the
    literal `1` and `A` is equivalent to `T` → decrement candidate.
    (Operand order matters: `T = 1 - T` is not a candidate.)

"Equivalent" means structurally identical expressions, ignoring whitespace and
comments (for simple variables: same variable name), or identical source text.

D5. Type guard — applies to every candidate: when *T* is an array element
    access `C[k]`, the candidate is kept only when the inferred type of the
    container `C` (unknown parts ignored) contains `array` (or any `X[]`
    typed-array form) **and** does not contain `string`. In every other case
    (container type unknown/empty, string, `ArrayAccess` object, mixed without
    array, …) array-element targets are skipped. Non-array-access targets
    (variables, properties, static properties) are not type-checked.

- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)

E1. Step other than the literal `1`: `$n += 2`, `$n = $n + $step`.
E2. `T = 1 - T`, `T = T - 2`, `T = 2 - T`, `T = T + 2`, `T = 2 + T`.
E3. Right operand is not the same expression as the target: `$a = $b + 1`.
E4. Array element writes whose container is not provably an array (string
    offsets, `ArrayAccess` objects, unknown types).
E5. Other compound operators (`*=`, `.=`, `|=`, …).
E6. Compound assignments are only checked by D1/D2 (`$n += $n + 1` is not a
    candidate).

## Report

- Range: the whole assignment expression, from the start of the target to the
  end of the right-hand side (trailing `;` excluded).
- Severity: info (fixtures tag it `weak_warning`).
- Message: `Use '{replacement}' instead.` where `{replacement}` is the fix text
  below.

## Fix

F1. Replace the entire assignment expression with the target's verbatim source
    text combined with `++` (increment) or `--` (decrement):
    - option `PREFER_PREFIX_STYLE = true` (default): `++{T}` / `--{T}`;
    - otherwise: `{T}++` / `{T}--`.
    Only `PREFER_PREFIX_STYLE` decides; `PREFER_SUFFIX_STYLE` is the UI
    counterpart and is implied by `PREFER_PREFIX_STYLE = false`.
    Also applied inside `for (...)` headers and any other expression context.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| PREFER_PREFIX_STYLE | bool | true | Fix produces `++$x` / `--$x`. |
| PREFER_SUFFIX_STYLE | bool | false | Mutually exclusive with the above (radio choice); fix produces `$x++` / `$x--` whenever `PREFER_PREFIX_STYLE` is false. |

## PHP versions

None.

## Examples

Options `PREFER_PREFIX_STYLE = false`, `PREFER_SUFFIX_STYLE = true`:

```php
<?php

<weak_warning descr="Use '$page++' instead.">$page += 1</weak_warning>;
<weak_warning descr="Use '$this->left--' instead.">$this->left -= 1</weak_warning>;
<weak_warning descr="Use '$total++' instead.">$total = 1 + $total</weak_warning>;
<weak_warning descr="Use '$total++' instead.">$total = $total + 1</weak_warning>;
<weak_warning descr="Use 'self::$depth--' instead.">self::$depth = self::$depth - 1</weak_warning>;
$total = 1 - $total;
$total = $total + 1.0;

$slots = [0, 0];
<weak_warning descr="Use '$slots[1]++' instead.">$slots[1] += 1</weak_warning>;

while (true) { <weak_warning descr="Use '$n--' instead.">$n = $n - 1</weak_warning>; }
```

```php
<?php

$page++;
$this->left--;
$total++;
$total++;
self::$depth--;
$total = 1 - $total;
$total = $total + 1.0;

$slots = [0, 0];
$slots[1]++;

while (true) { $n--; }
```

Default options (prefix style), no report cases:

```php
<?php

/** @var string $word */
$word[2] += 1;
/** @var \ArrayAccess $bag */
$bag['k'] -= 1;
$unknown[0] += 1;
$count = $count - 3;
<weak_warning descr="Use '++$count' instead.">$count += 1</weak_warning>;
```

## Divergences

None.

- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
- **String and bool operands (custos diverges).** When the operand's
  inferred type includes `string` or `bool`, `T + 1` is not `++T`: for a
  string `+ 1` is arithmetic (TypeError for `"abc"`) while `++` increments
  alphanumerically (`"abc"` → `"abd"`, `""` → `"1"`); for a bool `+ 1`
  gives an int while `++` leaves it unchanged. Such candidates are not
  reported. Unknown types are still reported (upstream behaviour).
