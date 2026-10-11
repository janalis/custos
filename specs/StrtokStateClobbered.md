---
id: StrtokStateClobbered
group: Probable bugs
kind: syntax
needs: [names]
php: { min: "", max: "" }
---

# StrtokStateClobbered

## Summary

Detect starting a second two-argument strtok sequence between an initial sequence call and its one-argument continuation on the same reachable path.

## Detection

- D1. Report starting a second two-argument strtok sequence between an initial sequence call and its one-argument continuation on the same reachable path.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. The path is searched backwards from the continuation over the statements that precede it in the same statement list. The newest start is the closest statement that is a two-argument strtok call, bare or as the assigned value; the initial start is the next such statement before it. Statements in between need not be adjacent to either call.
- D4. A statement crossed by the search must be unable to start a sequence: it contains no function, method or static call, object creation, include, eval, yield, closure or arrow function. Between the newest start and the continuation, one-argument strtok calls are also accepted, since they only advance the current sequence. Any other statement ends the search without a report.
- D5. When the statement list is the body of a `while`, `foreach` or `for` loop, or the main branch of an `if`, the search continues with the statements before that construct, provided its header (the condition, the iterated expression, or the `for` initialisers and conditions) satisfies D4. This covers the nested tokenizer shape, where an inner sequence is started and consumed inside the loop that continues the outer one. Other enclosing constructs (`else`/`elseif` branches, `do`, `switch`, `try`, function bodies) end the search.

## Exceptions (no report)

- E1. Completion/restart of the same intended sequence, no later continuation, or a wrapper known not to invoke strtok are excluded. Two starts on the same string expression are a restart.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. The continuation belongs to the newest sequence: all three delimiter arguments are plain string literals, the continuation's delimiter equals the newest start's and differs from the initial start's. The initial call then only took a leading token (`$scheme = strtok($url, ':'); $name = strtok($file, '.'); $ext = strtok('.');`). When a delimiter is not a literal, or both starts share the delimiter, the owner is ambiguous and the report is kept.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Keep strtok sequences separate.`

## Fix

No automatic fix: independent tokenizer state requires restructuring.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
$first=strtok($a,','); strtok($b,':'); $next=<warning descr="Keep strtok sequences separate.">strtok(',')</warning>;
```

```php
<?php
function cells($csv) {
    $row = strtok($csv, "\n");
    while ($row !== false) {
        $cell = strtok($row, ',');
        while ($cell !== false) {
            echo $cell;
            $cell = strtok(',');
        }
        $row = <warning descr="Keep strtok sequences separate.">strtok("\n")</warning>;
    }
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
