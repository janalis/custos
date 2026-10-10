---
id: ReadLoopProcessesEofFailure
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "7.0", max: "" }
---

# ReadLoopProcessesEofFailure

## Summary

An EOF-tested loop can pass a failed read to a strict string consumer. Loop on the successful read result before processing a line.

## Detection

- D1. A while condition is !feof(simpleLocalHandle); under declare(strict_types=1), its body unconditionally assigns fgets(sameHandle) to a local then passes it to resolved strlen, trim or strtoupper without a false guard, or directly nests fgets in that consumer. These consumers require strings and reject the false read result in strict mode. Highlight loop condition. Echo and coercive consumers are excluded because false can be harmless there.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Loop on successful reads before consuming line data.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Requires PHP 7.0 or later and strict scalar typing for the reported consumer.

## Examples

```php
<?php
declare(strict_types=1);$h=tmpfile();while(<warning descr="Loop on successful reads before consuming line data.">!feof($h)</warning>){strlen(fgets($h));}
$g=tmpfile();while(<warning descr="Loop on successful reads before consuming line data.">!feof($g)</warning>){$line=fgets($g);echo trim($line);}
```

Valid case:

```php
<?php
while(($line=fgets($h))!==false){echo $line;}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
