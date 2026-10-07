---
id: ComparisonOperandsOrder
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# ComparisonOperandsOrder

## Summary
Enforces one consistent operand order for equality comparisons against a
literal/constant: with the project-wide comparison style set to *yoda*, the
literal must come first (`10 === $n`); with the *regular* style (default), the
variable part must come first (`$n === 10`).

## Detection
D1. Node: a binary expression whose operator is one of `==`, `!=` (including
    the `<>` spelling), `===`, `!==`. Both operands must be present.

D2. Each operand is classified as **constant-like** when it is exactly (no
    parentheses unwrapping, no casts) one of:
    - a string literal of any form (single/double quoted, with or without
      interpolation, heredoc, nowdoc);
    - a global constant reference: a bare or namespaced constant name such as
      `true`, `false`, `null` (any case), `PHP_EOL`, `\App\LIMIT`, and the
      magic constants (`__DIR__`, `__LINE__`, `__CLASS__`, …);
    - a number literal (int or float, any base), or a unary minus applied
      directly to a number literal (`-5`, `-0.5`).
    Everything else is not constant-like — notably class constants
    (`Foo::BAR`), arrays, `+5`, `(5)`, function calls, variables.

D3. Report only when exactly one operand is constant-like:
    - style **yoda**: report when the *right* operand is the constant-like one;
    - style **regular** (default, also when the style is unset): report when the
      *left* operand is the constant-like one.

## Exceptions (no report)
E1. Both operands constant-like (`0 == ''`, `PHP_OS === 'Linux'`).
E2. Neither operand constant-like (`$a == $b`, `Foo::BAR === $x`).
E3. Already in the configured order.
E4. Any other operator (`<`, `>=`, `<=>`, `instanceof`, `??`, …).

## Report
- Range: the whole binary expression, from the start of the left operand to the
  end of the right operand.
- Severity: info (fixtures tag it `weak_warning`).
- Message (yoda): `Put the constant operand on the left side of the comparison.`
- Message (regular): `Put the constant operand on the right side of the comparison.`

## Fix
F1. Swap the two operands' source text. The operator and all whitespace and
    comments between the operands and the operator stay exactly where they
    were: `$total   != -1` → `-1   != $total`.
    Available for both styles.

## Options
| Option | Type | Default | Effect |
|---|---|---|---|
| (none) | | | Driven by the global comparison style setting (`comparisonStyle`: `regular` default, or `yoda`). |

## PHP versions
None.

## Examples
Style yoda:

```php
<?php

$isZero  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$counter === 0</weak_warning>;
$isNeg   = <weak_warning descr="Put the constant operand on the left side of the comparison.">$delta != -3</weak_warning>;
$isOn    = <weak_warning descr="Put the constant operand on the left side of the comparison.">$flag == true</weak_warning>;
$isUnix  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$sep !== "/"</weak_warning>;
$fileOk  = <weak_warning descr="Put the constant operand on the left side of the comparison.">$path <> __FILE__</weak_warning>;

// no report
$same    = 'a' === $label;
$both    = PHP_INT_MAX == 99;
$classes = $mode === Mode::FAST;
$less    = $counter < 5;
```

```php
<?php

$isZero  = 0 === $counter;
$isNeg   = -3 != $delta;
$isOn    = true == $flag;
$isUnix  = "/" !== $sep;
$fileOk  = __FILE__ <> $path;

// no report
$same    = 'a' === $label;
$both    = PHP_INT_MAX == 99;
$classes = $mode === Mode::FAST;
$less    = $counter < 5;
```

Style regular (default):

```php
<?php

if (<weak_warning descr="Put the constant operand on the right side of the comparison.">null === $handler</weak_warning>) {}
if (<weak_warning descr="Put the constant operand on the right side of the comparison.">'yes' != $answer</weak_warning>) {}
if ($answer == 'no') {}
```

```php
<?php

if ($handler === null) {}
if ($answer != 'yes') {}
if ($answer == 'no') {}
```

## Divergences
None.
