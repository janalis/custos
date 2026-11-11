# InvalidReturnType implementation notes

## Shared check

The rule and DocumentedReturnType report the two halves of
`semanticquery.ReturnTypeMismatch`, cached per `return` (or arrow function)
through `analysis.Context.Shared`. The return contract of each function-like
(`infer.Env.ReturnContract`: the native type and the own `@return` type,
names resolved, `self`/`parent`/`static` bound to the declaring class,
unknown in a trait) is prepared once per function and shared by its
returns. That query:

1. judges the default (PHPDoc-aware) type of the value with
   `types.AllRejected` against the native type, in the file's
   `strict_types` mode;
2. asks `infer.Env.ProvenTypeOf` only when the native type rejects it, and
   marks the finding proven when the native type rejects the proven type
   too;
3. otherwise judges the value against the `@return` type, unless the
   native type rejects that tag (stale) or it names a template.

Binding `static` to the declaring class is sound only for rejections:
`static` accepts a subset of what `self` accepts, so the check never asks
whether a value is accepted.

Generators are recognized by a `yield` in the body (nested function-likes
and classes aside); their `return` gives the completion value.

## Corpus triage

The corpora (guzzle, composer, laravel) give no proven finding. See
DocumentedReturnType's NOTES for the documented ones.
