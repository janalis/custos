---
id: NoContentResponseWithBody
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# NoContentResponseWithBody

## Summary

Detect nonempty output after response status is proven 204 and before any status change.

## Detection

- D1. Report nonempty output after response status is proven 204 and before any status change.
- D2. The status is proven 204 only while the last status write reaching the output on every path is `http_response_code(204)` and no `header()` call that may replace it can run in between. A `header()` call counts as a possible replacement when it sits after the status write and before the output, or anywhere in a loop that repeats the output, and any of these holds:
  - its third argument is not a known integer, or is a known integer other than `0` and `204`;
  - its line is a status line (`HTTP/...`) naming a code other than 204, or one whose code cannot be read;
  - its line sets the `Location` or `Status` field (a redirect turns 204 into 302; CGI and FPM send a `Status` field as the response status);
  - its field name cannot be read: the line is neither a string literal nor a concatenation whose leading literal contains the field name.

  Other header lines, including a concatenation such as `'Content-Type: ' . $type`, keep the proof.
- D3. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D4. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

## Exceptions (no report)

- E1. Empty output, changed or possibly changed statuses (D2), unknown status and HEAD-specific transport behavior are excluded.
- E2. Output written while an output buffer opened by `ob_start()` is known to be open is not reported: the buffer may be discarded (`ob_end_clean()`) or captured (`ob_get_clean()`) instead of sent. Output after the buffer was closed by `ob_end_clean()` or `ob_end_flush()` is checked again. Buffered output that is later flushed is a known missed case.
- E3. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: warning.
- Message: `Omit response content for status 204.`

## Fix

No automatic fix: changing the status or removing the body is an API decision.

A fix is offered only when the concrete prerequisites above are proven. Edits use exact source byte ranges, preserve comments and argument evaluation order, and introduce syntax supported by the configured PHP version. Never invent error-handling policy.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
http_response_code(204); <warning descr="Omit response content for status 204.">echo '{"done":true}';</warning>
```

Not reported, because the status no longer is 204 or the output is buffered:

```php
<?php
function fallback(): void
{
    http_response_code(204);
    header('X-Reason: fallback', true, 500);
    echo 'body';
}

function debugTrace(): void
{
    http_response_code(204);
    ob_start();
    echo 'trace';
    ob_end_clean();
}
```

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
