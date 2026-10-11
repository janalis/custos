---
id: GmpAutomaticBaseChangesDecimalInput
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GmpAutomaticBaseChangesDecimalInput

## Summary

Specify decimal base for zero-prefixed GMP input. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. gmp_init receives a valid digit-only constant string with leading zero, an omitted or zero base, and octal interpretation differs from decimal interpretation. Enabling selects decimal-input policy. Exclude all-zero strings and 0x/0b prefixes.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. A variable argument is not reported when its assigned literal may be stale at the call: the variable, or a variable whose value was copied into it, is changed by `++` or `--`, is bound by a `catch` clause, is captured by reference in a closure `use` list, or is stored by reference in an array item; or the enclosing scope calls `extract()` or one-argument `parse_str()`. A file-level variable is also a global, so a `global` declaration naming it, or any `$GLOBALS` access in the file, has the same effect.
- E3. Both @custos-ignore GmpAutomaticBaseChangesDecimalInput and @noinspection GmpAutomaticBaseChangesDecimalInput suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: false.
- Message: `Specify decimal base for zero-prefixed GMP input.`

## Fix

Append base 10 to a one-argument gmp_init call, or replace an explicit literal zero base with 10. Withhold for unpacking, unsupported named arguments, or comments inside the replacement range.

The fix is offered only when the argument expression itself fixes the digits: a string literal, a constant, or a class constant named through a fixed class. When the digits are read from a variable, or from a `static::` constant that a subclass may redefine, the finding is reported without a fix, because an explicit decimal base makes `gmp_init` throw on a value that turns out to carry a `0x` or `0b` prefix.

Use exact byte edits and preserve comments, evaluation count, argument order and supported PHP syntax. Withhold the fix whenever its prerequisites cannot be proven.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
$n = <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init('024')</warning>;
```

Fixed result:

```php
<?php
$n = gmp_init('024', 10);
```

Valid case:

```php
<?php
$n = gmp_init('024', 10);
```

Reported without a fix, since the digits come from local flow:

```php
<?php
function parse()
{
    $digits = '024';

    return <warning descr="Specify decimal base for zero-prefixed GMP input.">gmp_init($digits)</warning>;
}
```

Not reported, since the closure may have replaced the string:

```php
<?php
function read()
{
    $digits = '024';
    $refresh = function () use (&$digits) {
        $digits = '0x14';
    };
    $refresh();

    return gmp_init($digits);
}
```

## References

- [PHP API contract](https://www.php.net/manual/en/function.gmp-init.php).

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
