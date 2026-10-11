---
id: ConnectTimeoutAssumedToBoundReads
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "5.3", max: "" }
---

# ConnectTimeoutAssumedToBoundReads

## Summary

Set a separate timeout for stream reads. This opt-in rule applies the policy described in Detection.

## Detection

- D1. Enabled policy requires explicit bounded reads. Report fgets/fread/stream_get_contents/fgetcsv on a local/direct fsockopen or pfsockopen result whose connection timeout is a positive integer or float literal, unless a checked stream_set_timeout call on that same handle (same expression, same opening call) is known to have returned true when the read runs. Unknown origins and zero, negative, non-numeric or unknown connection timeouts are excluded. This checks per-read configuration; it does not establish an overall deadline.
- D1a. A condition proves success through its polarity: the call itself when true; `!` flips the polarity; `&&`/`and` when true and `||`/`or` when false prove it if either operand does; `==`, `===`, `!=`, `!==` against a `true`/`false` literal on either side are reduced to the compared operand. Any other expression proves nothing.
- D1b. Accepted guard positions, searched from the read outwards within the same function scope: the body of an `if`/`elseif` whose condition proves success when true; an `elseif` or `else` reached only after an earlier condition of the chain that proves success when false; the right operand of `&&`/`and` (left proves success when true) or `||`/`or` (left proves success when false); the matching arm of a full ternary; and an early exit, i.e. an earlier `if` in a statement list enclosing the read whose body always leaves the path (return, throw, exit, break, continue) and whose condition proves success when false.
- D1c. A read reached only when configuration failed (inverted condition, wrong branch) is reported. A bare `stream_set_timeout($h, …);` statement whose result is discarded is not a guard: the policy asks for a checked configuration, so the following read is still reported.
- D2. Resolve builtin functions, extension classes, method receivers, constants and named arguments. Follow unchanged locals and aliases only within the same bounded lexical scope. Require a reachable path; merge branch facts only when all incoming paths establish them.
- D3. Emit one finding at the violating operation, not one per evidence statement. Rebinding, writes, by-reference escape, unresolved calls that can mutate relevant state, and exhausted analysis budgets remove proof.

## Exceptions (no report)

- E1. Exclude shadowed builtins, unresolved receivers, incomplete syntax and unknown prerequisite facts. An extension function name alone does not prove an application function is that builtin.
- E2. A dominating valid guard or repair described in Detection prevents the finding. Both legacy noinspection and custos-ignore forms suppress this ID.

## Report

- Range: the violating expression or call described in D1; stateful absence checks highlight the final unsafe operation or loop.
- Severity: warning.
- Enabled by default: false.
- Message: `Set a separate timeout for stream reads.`

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
$s=fsockopen("example.test",443,$errno,$errstr,2); if ($s!==false) { <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>; }
```

Valid case:

```php
<?php
$s=fsockopen("example.test",443,$errno,$errstr,2); if ($s!==false && stream_set_timeout($s,2)) { fgets($s); }
```

Early exit, also valid:

```php
<?php
function banner(string $host) {
    $s = fsockopen($host, 443, $errno, $errstr, 2.5);
    if ($s === false || !stream_set_timeout($s, 2)) { return null; }
    return fgets($s);
}
```

Reported, the read runs only when no read timeout could be set:

```php
<?php
function banner(string $host) {
    $s = fsockopen($host, 443, $errno, $errstr, 2);
    if (!stream_set_timeout($s, 2)) { return <warning descr="Set a separate timeout for stream reads.">fgets($s)</warning>; }
    return null;
}
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
