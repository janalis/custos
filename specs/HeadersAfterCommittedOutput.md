---
id: HeadersAfterCommittedOutput
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# HeadersAfterCommittedOutput

## Summary

Detect a response-header write (header, setcookie, or http_response_code used as a setter) after demonstrably committed output on every incoming path, such as output plus flush with no active buffering.

## Detection

- D1. Report `header`, `setcookie`, or `http_response_code` called as a setter, after demonstrably committed output on every incoming path, such as output plus flush with no active buffering.
- D1a. `http_response_code` is a setter only when it receives a status argument (first positional or named `response_code`) that is not a known falsy value. A dynamic status counts as a setter.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Active output buffering, unknown server buffering, no preceding output and headers before commitment are excluded.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.
- E3. `http_response_code()` without a status argument, with a known `0` or `null` status, or with an unpacked argument list: the first two only read the current status, the last is not provably a write.
- E4. Calls that a `headers_sent()` test keeps off the committed path:
  - inside the body of `if (!headers_sent())` or of an `elseif` with that condition, or inside the `else` of `if (headers_sent())`;
  - directly after an `if (headers_sent())` statement without `else`/`elseif` whose body always leaves the block;
  - as the right operand of `headers_sent() || ...` / `headers_sent() or ...` or of `!headers_sent() && ...` / `!headers_sent() and ...`;
  - in the matching branch of a ternary on such a test.

  The test may be parenthesised, negated, compared with `true`/`false` (`===`, `==`, `!==`, `!=`), one operand of a conjunction that must hold, or one operand of a disjunction that must fail. It must target the builtin and belong to the same function scope as the call. The exception does not apply when the committing flush sits between the test and the call, because the test result is then stale.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Set response headers before committing output.`

## Fix

No automatic fix: moving headers can change control flow and response semantics.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
echo 'started'; flush(); <warning descr="Set response headers before committing output.">header('X-State: ready')</warning>;
```

Not reported: the status is only read, and the header call is guarded.

```php
<?php
echo 'started'; flush();
$status = http_response_code();
if (!headers_sent()) {
    header('X-State: ready');
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
