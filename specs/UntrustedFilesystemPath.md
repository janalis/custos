---
id: UntrustedFilesystemPath
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UntrustedFilesystemPath

## Summary

Detect request-derived path components reaching file read/write/delete or include sinks without established canonical containment or finite filename selection.

## Detection

- D1. Report request-derived path components reaching file read/write/delete or include sinks without established canonical containment or finite filename selection.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Proven boundary-aware containment with explicit symlink policy, fixed paths, validated finite names and unknown flows are excluded.
- E1a. Canonical containment is proven for a variable only when all of the following hold. Anything weaker keeps the report.
  - The variable's single reaching value at the sink is a call to builtin `realpath()`, so symlinks and `..` are already resolved.
  - A prefix test on that same value (same reaching assignment) has succeeded on every path to the sink: the sink is in the body of an `if` whose condition being true implies the test, or an earlier sibling `if` whose body always leaves (return, throw, exit, …) runs when the test fails. Implication looks through parentheses, `!`, any operand of `&&` (true side) and any operand of `||` (false side); an `else` branch or a non-leaving body proves nothing.
  - The test is `str_starts_with($p, PREFIX)`, `strpos($p, PREFIX) === 0` without an offset (or `!== 0` on the rejecting side, operands in either order), or `strncmp($p, PREFIX, N) === 0` with an integer `N` not smaller than the byte length of a literal `PREFIX`. Loose comparison is not accepted: `strpos() == 0` is also true for `false`.
  - `PREFIX` is a string literal longer than one byte that starts and ends with `/`, or `$base . '/'` / `$base . DIRECTORY_SEPARATOR` where `$base` resolves to a builtin `realpath()` call on a path without request data and is proven not `false` at the test. An unchecked base may be `false`, which turns the prefix into `/` and accepts every absolute path. A prefix without the trailing separator also matches sibling directories.
  - Nothing in the scope can rebind or alter the variable (or the `$base` variable) behind the analysis: no closure capturing it by reference, no `catch` binding or `static` declaration of the same name, no reference alias, no `extract()` call, no `global` declaration of it and no `$GLOBALS` use; and, located after the `realpath()` assignment, no other assignment to it (compound, destructuring slot, offset write, variable variable), no `++`/`--`, no `foreach` key or value variable and no argument bound to a by-reference parameter of a resolved callee. A write that ends before the `realpath()` assignment is superseded by it. The check is positional, so a later reuse of the variable name keeps the report. Where the shared flow analysis itself sees the rebinding happen before the sink, the value is unknown rather than request-derived and E1 applies instead: after a call that may run a closure capturing the variable by reference, after a `catch` clause bound to it, or after a call to a function of the same file that writes it through `global` or `$GLOBALS`.
  - An explicit `=== false` rejection of the canonical value is not required: each accepted test already fails for `false`.
- E1b. A value produced by a `match` expression whose arm results all carry no request data is a finite server-chosen name, whatever the subject is. One request-derived arm result keeps the report. The shared flow analysis provides this: a `match` carries the request data of its arm results only.
- E1c. A concatenation is excluded when every operand is clean or excluded by E1a/E1b.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Validate filesystem containment before using this path.`

## Fix

No automatic fix: canonicalization, existence and symlink races require deliberate caller policy.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Validate filesystem containment before using this path.">readfile('/srv/downloads/' . $_GET['file'])</error>;
```

Canonical containment and finite selection are not reported:

```php
<?php
function download(): void
{
    $path = realpath('/srv/downloads/' . $_GET['file']);
    if ($path === false || !str_starts_with($path, '/srv/downloads/')) {
        exit;
    }
    readfile($path);
}

function page(): void
{
    $page = match ($_GET['page']) { 'terms' => 'terms.html', default => 'home.html' };
    readfile('/srv/pages/' . $page);
}
```

A prefix without its trailing separator still admits a sibling directory:

```php
<?php
function download(): void
{
    $path = realpath('/srv/downloads/' . $_GET['file']);
    if ($path === false || !str_starts_with($path, '/srv/downloads')) {
        exit;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
