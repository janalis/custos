---
id: GeneratorReturnBeforeCompletion
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# GeneratorReturnBeforeCompletion

## Summary

A generator exposes its return value only after it finishes. Complete iteration before calling getReturn.

## Detection

- D1. A local assigned a call to a literal closure containing yield is immediately used as receiver of getReturn before any resume/iteration/next/send/throw operation. Report the method call; unknown calls or escape discard proof.
- D1d. Require an unconditional directly executed nondelegating yield in the generator body prefix before any terminating or ambiguous operation. Merely finding a yield somewhere is insufficient. Exclude dead or conditional yields, nested-scope yields, prior unconditional returns, and yield-from delegation, including delegation to an empty iterable.
- D1e. The first statement of the literal closure body must be an expression statement whose expression is a direct nondelegating yield. Preceding assignments, conditionals or other statements do not establish this initial suspended state; yield-from is excluded.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Finish the generator before reading its return value.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 7.0 or later.

## Examples

```php
<?php
$g=(function(){yield 1;return 2;})(); <warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>;
```

Valid case:

```php
<?php
$g=(function(){yield 1;return 2;})(); foreach($g as $v){} $v=$g->getReturn();
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
