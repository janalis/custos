---
id: ZipSourceRemovedBeforeArchiveClose
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ZipSourceRemovedBeforeArchiveClose

## Summary

Keep ZIP source files until the archive closes.

## Detection

- D1. Report unlink of a proven literal source path within the successful true branch of resolved ZipArchive::addFile using default flags, when no preceding close on that receiver exists. The matching literal path establishes source identity. Unknown paths, explicit flags, unguarded additions and conditional failed additions are excluded.
- D1a. A close on the same archive object counts as preceding when it is written before the unlink in the same function scope and is evaluated on every ordinary path that reaches the unlink. Besides a plain statement this covers a close used as an `if` or `elseif` condition (the code in either branch of that condition, or after an `if` guard that leaves), the left operand of a short-circuit operator, a ternary condition, a close in a `try` block whose `finally` block or following code holds the unlink, and a close in a `finally` block. When the unlink sits in the `finally` block, or after a `try` that has catch blocks, the close must belong to the first statement of the `try` block: an earlier protected statement could throw and skip it. A close that may be skipped does not count: one inside a branch body, loop body, ternary or match arm, the right operand of a short-circuit operator or of `??=`, a catch block or a callback; nor does a `try`-block close when the unlink sits in a catch block of that `try`.
- D1b. The same archive object means the same proven flow identity. An archive without one, such as a property or the result of a factory call, is matched by its receiver spelling (a variable, or a property chain rooted in a variable or a static property), provided neither the receiver nor an object it is read from is assigned between the addFile guard and the close.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

Only the successful branch establishes the deferred source-file dependency; a failed addition’s else branch is excluded.

- E0. An unlink written inside a closure or arrow function nested in the successful branch is not reported: it runs when the callback is invoked, which the rule does not track.
- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: true.
- Message: `Keep ZIP source files until the archive closes.`

## Fix

No automatic fix. The required repair depends on error handling, data selection, lifecycle ordering or application policy.

All edits use exact byte ranges and preserve comments and evaluation order. Offer no fix when its stated prerequisites cannot be proven or the replacement is unsupported by the configured PHP version.

## Options

None.

## PHP versions

Requires PHP 5.3 or later and the referenced extension API. Respect method-specific introduction versions within PHP 5.3–8.5; unsupported optional APIs cannot establish facts. Named arguments require PHP 8.0 or later.

## Examples

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; $z->close(); }
```

Valid case:

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { $z->close(); unlink("source.txt"); }
```

Checked close:

```php
<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { if (!$z->close()) { return; } unlink("source.txt"); }
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
