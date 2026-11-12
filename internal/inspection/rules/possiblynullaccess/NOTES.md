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
- Conditions correlating two values (`if ($a === null && $b === null)
  return; if ($b === null) $a->m();`) are not followed.
- Every access on the same nullable value is reported, even after one
  that would already have thrown.
- Builtin methods declared nullable that only return null in states the
  code rules out (`IteratorIterator::getInnerIterator()`) are reported.
