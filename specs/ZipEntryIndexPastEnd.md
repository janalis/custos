---
id: ZipEntryIndexPastEnd
group: Probable bugs
kind: flow
needs: [names, types, flow]
php: { min: "", max: "" }
---

# ZipEntryIndexPastEnd

## Summary

Use an existing ZIP entry index.

## Detection

- D1. Report a resolved ZIP indexed operation with a proven negative index, or an index directly equal to that same receiver numFiles property. Support getNameIndex, statIndex, getFromIndex, deleteIndex and getStreamIndex. Other inferred counts and unknown indices are excluded.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and code too complex to analyze remove proof.
- D4. A negative index is proven only by an integer literal, a resolved constant, or a local variable whose reaching value is such a literal. Such a variable is proven only while nothing in its scope can change it behind its plain assignments, anywhere and in any statement order: `++` or `--` on it or on one of its elements; a `catch`, `foreach` key or value binding it; a reference to it (a by-reference assignment on either side, a by-reference `foreach` over it, a by-reference array item, a by-reference capture by a nested closure, or an argument, other than the inspected argument itself, that a resolved callee binds by reference or that a callee custos cannot resolve receives); a write naming its target at run time (a variable variable, `extract()`, one-argument `parse_str()` or `mb_parse_str()`, `include`/`require`, `eval`); or a `$GLOBALS` access other than a literal entry naming another variable. A `global` or `static` declaration of it, or a by-reference capture by the closure that is the scope, removes the proof too. In top-level code, where locals are globals, a `global` declaration naming it in any function of the file, or a `$GLOBALS` access anywhere in the file, does as well. The same holds for every local whose value an assignment stores into it, whatever the assigned expression computes, directly or through further assignments. A read inside an arrow function also checks the scopes the arrow function captures from. A property or static property is never proof: its declared default says nothing about the value at the call.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.
- E3. Counter idioms that start below zero are valid: `$i = -1; while (++$i < $z->numFiles) { $z->getNameIndex($i); }`, `$i = -1; do { $i++; ... } while (...)`, and a cursor object whose position property defaults to `-1` and is advanced before each read.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Use an existing ZIP entry index.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within the supported PHP versions; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$z=new ZipArchive(); <warning descr="Use an existing ZIP entry index.">$z->getNameIndex($z->numFiles)</warning>;
```

Valid case:

```php
<?php
$z=new ZipArchive(); for ($i=0;$i<$z->numFiles;$i++) { $z->getNameIndex($i); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
