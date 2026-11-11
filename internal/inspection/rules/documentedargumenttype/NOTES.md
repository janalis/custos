# DocumentedArgumentType implementation notes

The check is shared with InvalidArgumentType (see its NOTES): this rule
reports the mismatches of `semanticquery.ArgumentTypeMismatches` that are
not proven.

## Value details

The index keeps `@param` types through `Type.DocString`, which drops integer
ranges, string lengths and literal string values, and literal arguments
carry no value facts, so those details never reach the check. Once the
index keeps them, they should count only against exact values (literals,
complete value sets): widen the `@param` type to its plain atoms otherwise.

## Corpus triage

The corpora (guzzle, composer, laravel) give no proven finding. The
documented findings on laravel are PHPDoc contracts the code breaks while
PHP accepts the call (untyped parameters): an array to `@param string`,
null to `@param string` and `@param callable`, an empty array to a
required-key shape, a `RedisCluster` to `@param \Redis`.
