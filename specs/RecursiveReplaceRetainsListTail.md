---
id: RecursiveReplaceRetainsListTail
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# RecursiveReplaceRetainsListTail

## Summary

Recursive array replacement can leave entries from the end of an older list. Replace the list explicitly when the shorter list should remove that tail. This inspection is disabled by default because the intended policy or use can vary.

## Detection

- D1. Resolved array_replace_recursive receives literal maps whose same key holds lists, where replacement list is shorter and a retained tail index is subsequently read from the result. Highlight call; require known literal map/list shape.
- D1a. The initial supported downstream proof is a single immediately following echo of the retained nested list index.
- D1b. A call may pass further positional replacement arrays. They are applied in order, so every replacement must be a known literal map. A replacement without the key leaves the list untouched; each one that has the key must hold a list too short to reach the read index, and at least one must do so.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. When the proof is inconclusive, nothing is reported.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E1a. Do not report when any replacement array supplies the read index, replaces the list with a non-list value, or is not a known literal map: the value read is then not a retained tail, or cannot be proven to be one. A named argument after the first two positions also ends the proof.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: false.
- Message: Replace the list explicitly to remove its old tail.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$c=<warning descr="Replace the list explicitly to remove its old tail.">array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]])</warning>; echo $c["hosts"][1];
```

Valid case:

```php
<?php
$c=array_replace_recursive(["port"=>80],["port"=>443]);
```

A later replacement that writes the index is also valid:

```php
<?php
$c=array_replace_recursive(["hosts"=>["a","b"]],["hosts"=>["c"]],["hosts"=>[1=>"z"]]); echo $c["hosts"][1];
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
