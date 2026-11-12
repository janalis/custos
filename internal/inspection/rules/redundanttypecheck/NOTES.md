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
- NotOptimalIfConditions: `instanceofSibling`, whatever its options.
- DomMissingAttributeCheckedAsNull: `ownedComparison` with `null`.

## Known gaps

- Always-true tests in the last arm of a `match (true)` are reported.
- `is_numeric()`, `is_callable()`, `is_resource()` and `is_countable()`
  are not judged.
