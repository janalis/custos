# PossiblyNullAccess implementation notes

## Shared check

The rule reports `semanticquery.PossiblyNullAccess`, cached per access node
through `analysis.Context.Shared`. PossiblyNullArgument shares its value
test (`possiblyNull`), which asks, cheapest first:

1. the default (PHPDoc-aware) type: no null there means PHPDoc ruled it
   out (`@var`, `@phpstan-assert`), or the value is unknown or `mixed`;
2. `infer.Env.ProvenTypeOf`: null must come from native declarations;
3. an environment with `infer.Options{Assertions: true,
   PropertyFactsSurviveCalls: true}`, cached per file: `assert()` calls
   count whatever the configuration says, and a property check survives
   the calls that end it in the proven environment;
4. `checkedCall`: an identical call checked earlier
   (`NativeEarlierCallProves`), or a local whose reaching values are such
   calls or never null.

## Repeated accesses

`repeatedAccess` (`possibly_null_repeat.go`) sets `NullAccess.Repeated`;
the rule skips such accesses, but NullPointerException still defers on
them, so a value is reported once. Per scope and value (`valueKey`), the
accesses on it and its other occurrences are listed once
(`valueAccesses`); an access scans at most `maxRepeatScan` earlier
accesses, nearest first, and stops at the first one separated from it by
another occurrence. `runsBefore` is syntactic dominance: `evaluated` for
the steps up from the earlier access, `evaluatedBefore` at the first
common ancestor. `usedBetween` widens the window to a loop around the
later access unless the loop runs the earlier one on each iteration
(`repeatsWith`: not a `foreach` subject or `for` initializer). Calls do
not count as uses of a `$this` property, matching
`PropertyFactsSurviveCalls` in `possiblyNull` (spec E6). A scope holding
a `goto` has no repeated access. The earlier access must be reported: `repeatedAccess`
asks `analysis.Context.Suppressed` (the engine's suppression comments,
built once per file and reused when filtering findings) and, when it is
suppressed, whether it repeats another itself. The same question covers
the baseline: the adapters (the CLI, the language server, MCP snapshots)
hand the engine `baseline.Covers` through `analysis.Engine.WithRecorded`
before analysis, and `--generate-baseline` hands it a predicate recording
everything, so a generated baseline holds every access and a later added
access is new. `Covers` consumes no count (see the spec's known limit).

## NullPointerException

NullPointerException asks `semanticquery.PossiblyNullAccess` or
`PossiblyNullArguments` before each report, only when
`analysis.Context.RuleChecks` says the matching rule's check of that access
or call runs: the deferral only happens when this rule or
PossiblyNullArgument runs in the same engine, was not skipped on the file
after crashing there, and its check of that node was not skipped for lack
of shared data (the engine then analyzes the file again with RuleChecks
answering false for that node only). It defers only where
the shared query returns a finding, the same one the rule reports, so a
`@noinspection PossiblyNullAccess` or a baseline entry silences the
location; the other way round, `suppressedBy` in `native-rules.json` makes
NullPointerException's suppressions apply to both rules, and the baseline
(`internal/project/baseline`) lets an entry of either side of a
`suppressedBy` link cover a finding of the other at the same location. The fixture `nullpointerexception/testdata/possibly-null-rules.php` runs
the three rules together (`"with"` sidecar key).

## Known gaps

- A chain after a nullsafe access (`$a?->b()->c()`) is never reported,
  even when `b()` itself may return null.
- Correlated conditions are followed only as `infer` does it
  (`correlated` in `narrow.go`): an `&&` early exit of at most
  `maxCorrelatedTests` operands all of whose operands but one test one
  local alone, proven true by the guards at the use (each proof narrows
  that local once). Longer exits, or two operands left unproven (a call,
  a test reading two values), are not followed.
- Builtin methods declared nullable that only return null in states the
  code rules out are reported, except the short list in
  `neverNullReturns` (`getInnerIterator()` of the SPL wrappers sharing
  IteratorIterator's implementation), checked through `checkedCall`.
  `NeverNullCall` reads the receiver's proven type and requires every
  class of it, and every class extending one (`index.Extending`, stubs
  included, so `IteratorIterator` fails through `AppendIterator`), to
  resolve the method to such an entry. The
  stubs carry no such fact, so a new entry needs the PHP source behind it.
