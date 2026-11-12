# ImpossibleTypeComparison implementation notes

## Shared check

The rule reports the comparisons of `semanticquery.DecidedTypeCheck`,
cached per node through `analysis.Context.Shared`; RedundantTypeCheck
reports its type tests, so a `=== null` is judged once for both. The
comparison (`impossibleTypeComparison` in `semanticquery/type_checks.go`)
goes:

1. cheap syntactic exclusions first (null literal, two literals,
   constants, `$this` in a closure);
2. the operand types, asked environment by environment and stopping at the
   first that does not decide (`operandType`): the default (PHPDoc-aware)
   environment, `infer.Env.ProvenTypeOf`, and an `infer.Env` built for
   `phpversion.Max` (its proven twin, cached per file) when the target is
   older;
3. `disjoint` on the coarse types (value facts dropped); exact values are
   compared only when both operands carry them and `exactOperand` accepts
   both (a literal, or a local of a function whose `ReachingAssignments`
   exclude its entry value and whose `PossibleValuesReaching` are all
   literals);
4. the context exclusions last: test context and `assert()`,
   `instanceofSibling`, `writtenLaterInLoop`, `versionedOperand`,
   `ownedComparison`.

Analysis budgets: an inference that hits a recursion or work bound yields
an unknown type, which step 2 rejects, so no separate gap check exists.

## Class relations

`typeChecks.classVerdict` wraps `types.Accepts` with `IndexRelations`
(strict mode) and answers Maybe for traits, unresolved hierarchies and
`dependencyHierarchy`: a class or ancestor whose file path has a `vendor`
segment. Two classes are disjoint when either direction is No.

## Owners

`comparisonOwners` lists the builtin results other rules report against a
constant; `ownedComparison` defers only when that rule runs
(`analysis.Context.RuleEnabled`). The fixtures `owners.php` (with the
owners, `"with"` sidecar key) and `owners-alone.php` (without) cover both.

## Known gaps

- Loose comparisons are never judged, even when no conversion could make
  the operands equal (an object against an int).
- Only an operand that reads a constant itself is skipped: a local
  assigned from a configuration constant whose stub or declaration gives
  `false` (`$debug = APP_DEBUG; $debug === true`) is judged by its atoms.
- The vendor test is a path heuristic: a `vendor-dir` configured under
  another name is not recognized.
