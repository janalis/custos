---
id: IntlFormattingFailureUnchecked
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# IntlFormattingFailureUnchecked

## Summary

Internationalization formatting can fail and return false. Check success before consuming the formatted string.

## Detection

- D1. A proven NumberFormatter or IntlDateFormatter format result is passed to a resolved string-only consumer strtoupper/strlen without a false guard. Track unmodified local results and direct nesting.
- D2. Resolve builtin functions, classes, constants and methods; user-defined lookalikes are excluded. Follow only bounded local proof in one lexical scope.
- D3. Emit one finding per violating operation, not per supporting evidence statement. Use existing analysis budgets; exhausted proof produces no report.

## Exceptions (no report)

- E1. Do not report unresolved or shadowed builtin names, unknown values, or incomplete receiver/type provenance. Unknown calls, aliases escaping, and intervening writes invalidate local proof. Similar code outside the stated detection contract is not a finding.
- E2. Do not report a use that a success check on the same, unmodified format result dominates. The result is a string or false, so each of these proves success: the result is truthy (`$s`, `!empty($s)`), differs from false (`!== false`, `!= false`, either operand order), `is_string($s)` holds, or `is_bool($s)` does not. Negation and `&&`/`||`/`and`/`or` combine such tests, and the assignment that stores the result may be tested in place. A check dominates the use when:
  - the use sits in the arm of an `if`/`elseif`/`else` chain that the check selects, including later arms once an earlier condition has ruled failure out;
  - the use sits in the ternary arm, or the right operand of a short-circuit operator, that is only evaluated on success;
  - an earlier statement of an enclosing statement list leaves the path on failure: an `if` whose failure arm always terminates, or a `check or throw/exit` (`failed and throw/exit`) statement.

  A check on another variable, on a repeated format call, or made before the variable was reassigned proves nothing. `isset`, null comparisons and `assert` are not accepted.
- E3. Both standard suppression forms suppress this native ID.

## Report

- Range: the violating expression or call, unless D1 specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: Check formatting success before consuming the formatted text.

## Fix

None. Choosing a repair requires intent or additional runtime information.

## Options

None.

## PHP versions

Available throughout supported PHP 5.3–8.5 where the referenced builtin/extension is available. Respect syntax/version-specific receiver contracts; unavailable APIs are not this rule’s concern.

## Examples

```php
<?php
$f=new NumberFormatter("en",NumberFormatter::DECIMAL); $s=$f->format($value); echo strtoupper(<warning descr="Check formatting success before consuming the formatted text.">$s</warning>);
```

Valid case:

```php
<?php
$f=new NumberFormatter("en",NumberFormatter::DECIMAL);$s=$f->format($value);if($s!==false){echo strtoupper($s);}
```

```php
<?php
function width(NumberFormatter $f, $value): int {
    $s = $f->format($value);
    if (!$s) {
        return 0;
    }
    return strlen($s);
}
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement.
