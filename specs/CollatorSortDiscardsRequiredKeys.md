---
id: CollatorSortDiscardsRequiredKeys
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# CollatorSortDiscardsRequiredKeys

## Summary

Preserve keys during collation sorting. Disabled by default. Explicit enablement establishes the policy described in Detection; names, comments and example labels never establish intent.

## Detection

- D1. Enabled policy requires preserving known noninteger record keys. Report Collator::sort applied to a local array proven to have at least one string key, including a known literal. Skip list arrays, unknown keys, and Collator::asort.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape and unresolved calls that can mutate relevant state remove proof. When the proof is inconclusive, nothing is reported.
- D4. The sorted variable is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from. The sort's own by-reference binding of its argument does not count. By-value captures, captures of other variables, and captures written in a nested function, method, closure or arrow function (which name that scope's own variable) keep the proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. No report, and therefore no fix, when D4 removes the proof: after the hidden replacement the array may be a list, and switching to asort would stop the reindexing that code relies on.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Preserve keys during collation sorting.`

## Fix

Replace resolved method token sort with asort; preserve arguments and receiver evaluation.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within the supported PHP versions; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; <warning descr="Preserve keys during collation sorting.">$c->sort($a)</warning>;
```

Fixed result:

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; $c->asort($a);
```

Valid case:

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"]; $c->asort($a);
```

Not reported, the closure may have replaced the array (D4):

```php
<?php
$c=new Collator("en_US"); $a=["part-q"=>"Quartz","part-a"=>"Amber"];
$reload=function () use (&$a) { $a=["Zinc","Agate"]; }; $reload(); $c->sort($a);
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
