---
id: UntrustedShellCommand
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedShellCommand

## Summary

Detect request-derived data reaching a shell-command sink in syntax-capable position without proven appropriate argument escaping or strict validation.

## Detection

- D1. Report request-derived data reaching a shell-command sink in syntax-capable position without proven appropriate argument escaping or strict validation.
- D1a. Shell-command sinks are the first (`command`) argument of resolved builtin `system`, `exec`, `shell_exec`, `passthru`, `popen` and `proc_open`, and the text of a backtick command expression, which PHP runs exactly like `shell_exec`. A call to a project function is a sink when its summary forwards an argument, or request data it reads itself, to one of the builtin call sinks.
- D1b. Decide per operand whether request data can become command text. Concatenation, interpolation and `??` carry any of their operands; a conditional or `match` expression carries only its result operands, never its condition or subject; a variable with a single reaching assignment carries what that assignment carries. Every other request-derived operand carries its data unless E1 or E3 applies. A variable is followed to its reaching assignment only when its scope cannot rebind it behind that assignment; a reference alias, a by-reference closure capture, `foreach` or argument, a `catch` binding or `static` declaration of the same name, `extract()`, an assignment through a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry keeps the report.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Resolved argument-vector execution, escaped arguments with validated option policy, constant commands and unknown flows are excluded. `proc_open` whose command is an array literal, a variable holding one or an array-typed expression is argument-vector execution. A command assembled only from constants selected by request data (for example `match` arms or conditional branches that are all literals) is a constant command.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.
- E4. Under E3, numeric scalar output covers: integer, float and boolean casts; the results of `-`, `*`, `/`, `%` and `**` and of unary `+` and `-`, which yield a number or fail; a resolved builtin call whose declared result on the configured PHP version consists only of integer, float, boolean and null (such as `strlen`, `count` or `strpos`); and `+` when at least one operand is a literal or such a builtin call. `+` between two operands that may both be arrays is an array union that keeps request-derived elements and is not excluded. Bitwise operators on strings produce strings and are not excluded.
- E5. A wrapper call is not reported when every request-derived argument written at the call is excluded by E1, E3 or E4.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Keep untrusted input out of shell command syntax.`

## Fix

No automatic fix: executable, option boundaries and allowed arguments require explicit policy.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Keep untrusted input out of shell command syntax.">system('viewer ' . $_GET['path'])</error>;
$listing = <error descr="Keep untrusted input out of shell command syntax.">`ls {$_GET['dir']}`</error>;
<error descr="Keep untrusted input out of shell command syntax.">proc_open('convert ' . $_GET['file'], [], $pipes)</error>;

// Not reported: the request only selects a constant, or contributes a number.
$mode = match ($_GET['mode']) { 'full' => '--full', default => '--quick' };
system('backup ' . $mode);
system('head -c ' . strlen($_POST['body']) . ' /dev/zero');
proc_open(['convert', $_GET['file']], [], $pipes);
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
