---
id: IntlCalendarSecondsPassedAsMilliseconds
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "5.5", max: "" }
---

# IntlCalendarSecondsPassedAsMilliseconds

## Summary

Convert seconds to calendar milliseconds.

## Detection

- D1. Report resolved IntlCalendar::setTime with direct time(), or an unchanged local assigned time(), as its timestamp. Exclude arbitrary numbers and arithmetic/computed timestamps. Offer a fix only for the direct call argument.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and code too complex to analyze remove proof.
- D4. A local timestamp is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it removes the proof too; a by-reference capture by the closure that is the scope keeps it, since only the closure's caller can rebind the variable, between calls. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, removes the proof as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from. A by-value capture keeps the proof.

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

Requires PHP 5.5 or later and the referenced extension API. Respect method-specific introduction versions within the supported PHP versions; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

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

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
