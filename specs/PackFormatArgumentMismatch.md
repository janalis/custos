---
id: PackFormatArgumentMismatch
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# PackFormatArgumentMismatch

## Summary

Match pack values to the format.

## Detection

- D1. Parse a constant pack format and report missing required value arguments. Numeric directives consume their repeat counts, string directives consume one value, and padding/position directives consume none. Supported numeric directives are c, C, s, S, n, v, i, I, l, L, N, V, q, Q, J, P, f, g, G, d, e and E. String directives are a, A, Z, h and H; cursor directives are x, X and @. Star counts consume remaining arguments; unknown/unpacked arguments invalidate count proof. Do not report extra arguments here. Exclude q, Q, J and P before PHP 5.6 and g, G, e and E before PHP 7.0; exclude Z before PHP 5.5. Unknown format directives invalidate proof.
- D2. Resolve builtin functions, extension classes, constants and method receivers, including aliases and supported named arguments. Inspect reachable operations in one lexical scope with bounded local flow.
- D3. Emit one finding per violating operation. Never infer policy from names or comments. Unsupported proof and exhausted analysis budgets produce no report.

## Exceptions (no report)

- E1. Exclude shadowed or unresolved symbols, incomplete code, unknown required constants, ambiguous receiver identity and unsupported versions. Reassignment, escaping aliases, and unknown calls invalidate dependent flow facts.
- E2. Both @custos-ignore PackFormatArgumentMismatch and @noinspection PackFormatArgumentMismatch suppress this native rule.

## Report

- Range: the complete violating expression or call identified by D1; source-to-use findings highlight the final unsafe use, except where D1 specifies the producer.
- Severity: warning.
- Enabled by default: true.
- Message: `Match pack values to the format.`

## Fix

None. A safe repair requires choosing values, error handling, or application behavior.

## Options

None. Use normal rule configuration to enable or disable the inspection.

## PHP versions

Apply throughout PHP 5.3–8.5 only where the extension API and syntax exist.  Do not flag API unavailability through this rule.

## Examples

```php
<?php
$bytes = <warning descr="Match pack values to the format.">pack('N2', 11)</warning>;
```

Valid case:

```php
<?php
$bytes = pack('N2', 11, 12);
```

## Divergences

Native custos inspection, independently specified. No EA counterpart or EA fixture conformance applies.
