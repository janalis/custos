# InvalidArgumentType implementation notes

## Shared check

The rule and DocumentedArgumentType report the two halves of
`semanticquery.ArgumentTypeMismatches`, cached per call node through
`analysis.Context.Shared` so a run enabling both checks each call once.
That query:

1. binds the written arguments (`bindableArgs`, `bindParam`);
2. checks them against the nominal contract of `semanticquery.ResolveCallee`
   first, and only for an argument it rejects lists every target with
   `semanticquery.CallTargets` (the overrides walk of
   ArgumentCountMismatch), since a finding needs every target to reject;
3. judges with `types.AllRejected` on the default (PHPDoc-aware) type of
   the argument, against the native type and, for project targets, the
   `@param` type;
4. asks `infer.Env.ProvenTypeOf` only when every target's native type
   rejects the default type, and marks the finding proven when the native
   types reject the proven type too.

A proven type rejected while the default type is accepted is not reported:
it means PHPDoc contradicts the code, which is another rule's concern.

## Builtin parameter types

`semanticquery.TrustedBuiltinTypes` trusts a builtin from PHP 8.0, in the
directory of a bundled extension (`semanticquery.BundledStub`), unless it
is listed in `semanticquery/stub_type_drift.txt`. That list was generated
by dumping every stub function and method resolved at PHP 8.0, 8.1, 8.2,
8.3 and 8.5 (index `Function` / `FindMethod`, parameters filtered by
version) and comparing each native parameter type with
`ReflectionParameter::getType()` of the same PHP version, every bundled
extension loaded: a type that differs (normalized union sets), a stub type
on a parameter PHP leaves untyped or does not have, or a declaration PHP
8.5 does not know. About 5,000 declarations matched; 327 names are listed.
Parameters the stubs type per version (`LanguageLevelTypeAware`) carry no
native type in the index and are never checked. Rerun the comparison when
the stubs are updated (`make stubs`).

Stub `@param` types are not consulted, and named arguments to builtins are
not bound, since stub parameter names were not compared.

## Known gaps

- Top-level variables of a file with an unbraced `namespace` statement
  read as nullable in the native environment, so `$x = 'a'; f($x);` there
  is not reported against a nullable parameter type.
- Subtypes the index does not see (an unindexed vendor directory) may
  override a method with a wider parameter type.
