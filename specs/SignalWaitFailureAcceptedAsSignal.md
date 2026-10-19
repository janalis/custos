---
id: SignalWaitFailureAcceptedAsSignal
group: Probable bugs
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# SignalWaitFailureAcceptedAsSignal

## Summary

Identify truthiness test of pcntl_sigwaitinfo result followed by passing that unchanged result to a call in true branch. Failure -1 is truthy; explicit >0 or !== -1 excludes.

## Detection

- D1. Report truthiness test of pcntl_sigwaitinfo result followed by passing that unchanged result to a call in true branch. Failure -1 is truthy; explicit >0 or !== -1 excludes. The tested operand is either a variable holding the result or an assignment of the result written in the condition itself.
- D1a. When the condition assigns the result to a plain variable, the receiving function call may be anywhere in the true branch, provided nothing earlier in the branch can change the variable: no write (assignment, compound assignment, `++`/`--`, `unset`, destructuring, `foreach` target, `global`/`static`, `catch`), no element write, no reference assignment, reference array item or by-reference closure capture, no pass to a method call, static call or constructor, and no variable-variable, `extract`, `parse_str`, `eval` or include. Reads, by-value captures and plain function calls do not end the proof. When the condition assigns to any other target (a property, an element), only a call that is the first statement of the branch counts.
- D1b. A receiving call is not counted when -1 cannot reach it. That is proven by an integer comparison of the same value (the call, a local holding its result, or a plain copy of that local) made after the wait call returned, in one of these positions: the condition of an enclosing `if`/`elseif` whose taken branch holds the call (the `else` side uses the negation of the `if` and of every earlier `elseif`); the condition of an enclosing ternary; the left operand of an enclosing `&&`/`and` or `||`/`or`; the condition of an enclosing `while` loop; or an earlier sibling `if` of any enclosing statement list whose body always leaves (return, throw, exit, break, continue). Comparisons `==`, `===`, `!=`, `!==`, `<`, `<=`, `>`, `>=` against a known integer are evaluated for -1, in either operand order, through `!`, and through `&&` (when the condition holds) or `||` (when it fails). The finding stays when at least one receiving call is unguarded. A guard written before the wait call, a guard whose body does not leave the block, a comparison of another value or against a non-integer, and a comparison that -1 satisfies on the path to the call prove nothing; the finding stays.
- D2. Resolve builtin functions, classes, constants and methods, including imports and supported named arguments. Follow only bounded local facts within one lexical scope. Receiver and argument provenance must establish the specific API contract; variable names and comments never prove intent.
- D3. Report one finding per violating operation. A reachable path, unchanged value identity and applicable API behavior are required. Intervening mutation, unknown calls involving tracked values, escaped aliases or exhausted analysis budgets invalidate proof.

## Exceptions (no report)

- E1. Exclude unresolved required symbols, user-defined lookalikes, incomplete syntax, unknown required values and unsupported versions. Cases outside D1 receive no finding.
- E2. Both `@custos-ignore SignalWaitFailureAcceptedAsSignal` and `@noinspection SignalWaitFailureAcceptedAsSignal` suppress this native rule.

## Report

- Range: the complete violating call or expression. Follow D1 if it specifies a narrower range.
- Severity: warning.
- Enabled by default: true.
- Message: `Reject signal-wait failure before handling the result.`

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
if ($signal = pcntl_sigwaitinfo([SIGTERM], $info)) { handleSignal($signal); }
```

Valid case:

```php
<?php
$signal = pcntl_sigwaitinfo([SIGTERM], $info); if ($signal > 0) { handleSignal($signal); }
```

Also a finding (the call is not the first statement, the variable is untouched before it):

```php
<?php
if ($signal = pcntl_sigwaitinfo([SIGTERM], $info)) { logWakeup(); handleSignal($signal); }
```

Also valid (failure leaves before the truthiness test):

```php
<?php
$signal = pcntl_sigwaitinfo([SIGTERM], $info); if ($signal === -1) { return; } if ($signal) { handleSignal($signal); }
```

## Divergences

Native custos inspection, independently specified; no EA counterpart or EA conformance requirement. Conservative proof intentionally excludes unknown runtime state and intent.
