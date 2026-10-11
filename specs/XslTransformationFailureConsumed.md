---
id: XslTransformationFailureConsumed
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# XslTransformationFailureConsumed

## Summary

Identify transformToXML result on a resolved XSLTProcessor passed directly or unchanged to file_put_contents or echo without rejecting false.

## Detection

- D1. Report transformToXML result on a resolved XSLTProcessor passed directly or unchanged to file_put_contents or echo without rejecting false.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D4. A value counts as having rejected false when the use is dominated by a test of the same unchanged value that cannot pass for false:
  - a strict comparison with `false` (`!== false` branch, or a terminating `=== false` guard);
  - a truthiness test (`if ($out)`, `if (!$out) { return; }`, `empty($out)`), since false and null are both falsy;
  - a loose comparison with `false` (`!= false` branch, terminating `== false` guard);
  - a resolved builtin `is_string()` test.
  The test may be the condition of an enclosing `if`/`elseif`/`else` branch or ternary arm, the left operand of a short-circuit `&&`/`||`/`and`/`or` whose right operand holds the use, or an earlier sibling `if` without `else`/`elseif` whose body always terminates. Negation, parentheses, conjunctions (for the passing side) and disjunctions (for the failing side) are followed. An assignment of the call inside the condition is the tested value.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E3. A guard on a different value, on an earlier transformation result that was since reassigned, or on a test that false can pass (`isset`, `!== null`, `is_string(...) || ...`) does not suppress the finding.
- E2. Both `@custos-ignore XslTransformationFailureConsumed` and `@noinspection XslTransformationFailureConsumed` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject transformation failure before using the XML.`

## Fix

None. A repair requires a semantic decision or application error-handling policy.

## Options

None.

## PHP versions

Available in PHP 5.3–8.5 where the referenced API exists. Do not apply extension contracts when the resolved API is unavailable. Gate syntax and version-specific behavior to the configured target.

## Examples

Finding (enable the rule explicitly when disabled by default):

```php
<?php
$processor = new XSLTProcessor(); file_put_contents($output, $processor->transformToXML($document));
```

Valid case:

```php
<?php
$processor = new XSLTProcessor(); $out = $processor->transformToXML($document); if ($out !== false) { file_put_contents($output, $out); }
```

```php
<?php
$processor = new XSLTProcessor(); $out = $processor->transformToXML($document); if (!$out) { throw new RuntimeException('transformation failed'); } echo $out;
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
