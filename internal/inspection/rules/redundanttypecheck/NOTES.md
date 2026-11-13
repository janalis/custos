# RedundantTypeCheck implementation notes

## Shared check

The rule reports the type tests of `semanticquery.DecidedTypeCheck`
(`redundantTypeCheck`), cached per node through `analysis.Context.Shared`.
It shares the operand typing, class relations and exclusions of
ImpossibleTypeComparison (see its NOTES):
`predicateVerdict` judges the test on every atom of the value's type in
each environment, and all must give the same decided answer.

`lastBranchCondition` walks up through parentheses and `!` (flipping the
answer) to an `ElseIf` that is the last of a chain without `else`, or an
`if` alone in an `else` (`else if` in two words).

## Owners

- CallableParameterUseCaseInTypeContext: an always-false predicate of its
  table on a parameter of the directly enclosing function or method, when
  that rule runs. Its other cases (always-true predicates, closure
  parameters, aliases such as `is_long()`) are reported here.
- NotOptimalIfConditions: `notOptimalReports` replays that rule's operand
  split (`conditionOperands`) and its D3/D4 conditions, when it runs. Its
  `REPORT_INSTANCE_OF_FLAWS` option is not visible across rules: turned
  off, the check is still left to it.
- DomMissingAttributeCheckedAsNull: `ownedComparison` with `null`, the
  method matched through `NativeMethod` (native receiver type, subclasses
  included) as the owning rule does.

## Settings

The newest-PHP twin environment takes `analysis.Context.InferOptions`, so
`flow.assertions` holds for every environment.

## Known gaps

- Always-true tests in the last arm of a `match (true)` are reported.
- `is_numeric()`, `is_callable()`, `is_resource()` and `is_countable()`
  are not judged.
