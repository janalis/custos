# DocumentedArgumentType implementation notes

The check is shared with InvalidArgumentType (see its NOTES): this rule
reports the mismatches of `semanticquery.ArgumentTypeMismatches` that are
not proven.

## Value details

The index keeps `@param` types through `Type.DocString`, which drops integer
ranges, string lengths and literal string values. `Param.DocValues` keeps
the integer ranges and a non-empty string (`Type.ValueDocString`); the
check judges an argument against it only when the argument lists exact
values (`semanticquery.exactValues`: integer and string literals), and
against the plain `DocType` otherwise. Other lengths and literal string
values are never kept. Magic `@method` parameters have no `DocValues`.

## Corpus triage

The corpora (guzzle, composer, laravel) give no proven finding. The
documented findings on laravel are PHPDoc contracts the code breaks while
PHP accepts the call (untyped parameters): an array to `@param string`,
null to `@param string` and `@param callable`, an empty array to a
required-key shape, a `RedisCluster` to `@param \Redis`.
