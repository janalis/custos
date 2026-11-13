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
   literals). A loose comparison goes through `looseDisjoint` instead,
   whose pair table (`looseAtomsMeet`) is the spec's D5 table, never with
   values; `plainObject` rejects interfaces and classes with a builtin
   class ancestor (identity with the `stubs.Index()` declaration);
4. the context exclusions last: test context and `assert()`,
   `notOptimalReports`, `writtenLaterInLoop`, `versionedOperand`,
   `ownedComparison`.

Analysis budgets: an inference that hits a recursion or work bound yields
an unknown type, which step 2 rejects, so no separate gap check exists.

## Class relations

`typeChecks.classVerdict` wraps `types.Accepts` with `IndexRelations`
(strict mode) and answers Maybe for traits, unresolved hierarchies and
`dependencyHierarchy`: a class or ancestor whose file lies in the vendor
directory the project layer records (`index.World.Vendor`, from
composer.json's `config.vendor-dir`; fixtures set it with the `vendorDir`
sidecar key). Two classes are disjoint when either direction is No.

## Owners

`comparisonOwners` lists the builtin results other rules report against a
constant; `ownedComparison` defers only when that rule's check of the
comparison or of the call runs (`analysis.Context.RuleChecks`). Checks
NotOptimalIfConditions reports come from `semanticquery.InstanceofFlaws`,
which that rule reports from too, and are left to it only when its
`REPORT_INSTANCE_OF_FLAWS` option is on (`analysis.Context.RuleBool`). The
fixtures `owners.php` (with the owners, `"with"` sidecar key),
`owners-flaws-off.php` and `owners-alone.php` (without) cover them.

## Known gaps

- Loose comparisons of a plain object against a string are never judged,
  even for a final class without `__toString()`; nor are an enum against
  its backing value, or exact values (`'a' == 'b'` through locals).
- Only an operand that reads a constant itself is skipped: a local
  assigned from a configuration constant whose stub or declaration gives
  `false` (`$debug = APP_DEBUG; $debug === true`) is judged by its atoms.
- Without a project (no index world), no class counts as a dependency's.
