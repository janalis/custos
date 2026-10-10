---
id: RequestControlledSessionId
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RequestControlledSessionId

## Summary

Accepting a session identifier from request data can let a caller choose the session being used. Generate identifiers independently of untrusted input.

## Detection

- D1. Resolved session_id setter receives direct request-superglobal data, or an unmodified local carrying it. Getter with no argument excluded.
- D1e. Source provenance additionally requires an untouched lexical superglobal prefix. Exclude earlier explicit writes to or reference escapes of that superglobal and any earlier write or reference target involving GLOBALS. Reject all earlier object construction and unknown calls, including calls with no arguments, method calls and static calls; only resolved strlen, is_string, ctype_alnum, ctype_digit or in_array are supported pure-call exceptions. Traverse at most 4096 AST nodes, excluding nested variable scopes; exhaustion or unsupported source mutation discards the request-origin proof.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Generate session identifiers independently of request data.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
<warning descr="Generate session identifiers independently of request data.">session_id($_GET["sid"])</warning>;
```

Valid case:

```php
<?php
session_id();session_start();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
