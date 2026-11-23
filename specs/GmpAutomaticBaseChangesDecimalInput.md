---
id: GmpAutomaticBaseChangesDecimalInput
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# GmpAutomaticBaseChangesDecimalInput

## Summary

Specify decimal base for zero-prefixed GMP input. This inspection is disabled by default; enabling it selects the stated policy.

## Detection

- D1. gmp_init receives a valid digit-only constant string with leading zero, an omitted or zero base, and octal interpretation differs from decimal interpretation. Enabling selects decimal-input policy. Exclude all-zero strings and 0x/0b prefixes.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. A variable argument is not reported when its assigned literal may be stale at the call. The variable is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from.
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

None.

## PHP versions

Apply on every supported PHP version only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

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

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
