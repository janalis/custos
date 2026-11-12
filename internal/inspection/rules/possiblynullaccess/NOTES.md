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
common ancestor.

## NullPointerException

NullPointerException asks `semanticquery.PossiblyNullAccess` or
`PossiblyNullArguments` before each report, only when
`analysis.Context.RuleEnabled` says the matching rule runs: the deferral only
happens when this rule or PossiblyNullArgument runs in the same engine.
The fixture `nullpointerexception/testdata/possibly-null-rules.php` runs
the three rules together (`"with"` sidecar key).

## Known gaps

- A chain after a nullsafe access (`$a?->b()->c()`) is never reported,
  even when `b()` itself may return null.
- Correlated conditions are followed only as `infer` does it
  (`correlated` in `narrow.go`): a two-operand `&&` early exit whose
  other operand tests one local alone, proven true by the guards at the
  use. Exits on three operands, or on tests reading two values, are not.
- Builtin methods declared nullable that only return null in states the
  code rules out are reported, except the short list in
  `neverNullReturns` (`getInnerIterator()` of the SPL wrappers sharing
  IteratorIterator's implementation), checked through `checkedCall`. The
  stubs carry no such fact, so a new entry needs the PHP source behind it.
