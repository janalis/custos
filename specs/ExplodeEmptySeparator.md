---
id: ExplodeEmptySeparator
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# ExplodeEmptySeparator

## Summary

Detect explode with a separator statically equal to the empty string.

## Detection

- D1. Report explode with a separator statically equal to the empty string.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases or recursion cannot establish definite facts. When the proof is inconclusive, nothing is reported.
- D4. A separator read from a variable is definite only when plain assignments are the sole way that variable can change in the scope of the call. See E3 for the writes that make it indefinite.

## Exceptions (no report)

- E1. Unknown or nonempty separators and shadowed application functions are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. No report when a variable of the separator expression, or a variable whose value was copied into it by assignment, can be rewritten behind the tracked assignments:
  - it is captured by reference in a closure `use` list in the scope (including the closure that is the scope itself);
  - it is declared `global` in the scope;
  - it is the operand of `++` or `--` (incrementing an empty string yields `"1"`);
  - it is bound by reference as an array item (`['k' => &$sep]`);
  - the scope calls the builtin `extract()`, or `parse_str()` with a single argument, either of which may define any local;
  - the call is at file level, where locals are globals, and the file declares the variable `global` in any function or mentions `$GLOBALS` anywhere.

  The check is lexical and deliberately conservative: it does not prove that the rewriting code runs between the assignment and the call.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Supply a nonempty separator.`

## Fix

No automatic fix: byte splitting, code-point splitting and grapheme splitting are different contracts.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$parts=<error descr="Supply a nonempty separator.">explode('',$text)</error>;
```

A local that only plain assignments can change is still definite:

```php
<?php
function cells(string $row) {
    $glue = '';
    return <error descr="Supply a nonempty separator.">explode($glue, $row)</error>;
}
```

Not reported, because the closure may rewrite the captured variable before the call:

```php
<?php
function cells(string $row) {
    $glue = '';
    $configure = function () use (&$glue) { $glue = '|'; };
    $configure();
    return explode($glue, $row);
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
