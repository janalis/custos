---
id: IntlCalendarSecondsPassedAsMilliseconds
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# IntlCalendarSecondsPassedAsMilliseconds

## Summary

Convert seconds to calendar milliseconds.

## Detection

- D1. Report resolved IntlCalendar::setTime with direct time(), or an unchanged local assigned time(), as its timestamp. Exclude arbitrary numbers and arithmetic/computed timestamps. Offer a fix only for the direct call argument.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.
- D4. A local is not unchanged when it, or a variable whose value was copied into it, is captured by reference in a closure `use` of the same scope or bound by a `catch` clause of that scope. A file-level variable is also a global: a `global` declaration naming it, or any `$GLOBALS` access anywhere in the file, removes proof as well. A by-value capture does not.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report for `$t = time(); $f = function () use (&$t) { $t *= 1000; }; $f(); $c->setTime($t);`: the closure may already have converted the value.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Convert seconds to calendar milliseconds.`

## Fix

For direct resolved time() argument only, replace the exact source of the call with (original * 1000). When the argument is already parenthesized, keep those parentheses and everything between them and the call, and only append the multiplication to the call: `(/* note */ time())` becomes `(/* note */ time() * 1000)`. Evaluate once; the edit never covers a comment outside the call.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.5 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c = IntlCalendar::createInstance(); <warning descr="Convert seconds to calendar milliseconds.">$c->setTime(time())</warning>;
```

Fixed result:

```php
<?php
$c = IntlCalendar::createInstance(); $c->setTime((time() * 1000));
```

Valid case:

```php
<?php
$c = IntlCalendar::createInstance(); $c->setTime(time() * 1000);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
