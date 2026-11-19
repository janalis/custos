# DocumentedArgumentType implementation notes

The check is shared with InvalidArgumentType (see its NOTES): this rule
reports the mismatches of `semanticquery.ArgumentTypeMismatches` that are
not proven.

## Value details

The index keeps `@param` types through `Type.DocString`, which drops integer
ranges, string lengths and literal string values. `Param.DocValues` keeps
the integer ranges, a non-empty string and listed literal string values
(`Type.ValueDocString`); the check judges an argument against it only when
the argument lists exact values (`semanticquery.exactValues`: integer and
string literals), and against the plain `DocType` otherwise. Other lengths
and string exclusions are never kept, nor literal strings with a
backslash, a quote or a `$`. Magic `@method` parameters have no
`DocValues`. When only the values form rejects the argument, the message
prints the value (`semanticquery.valueName`).

`#[ExpectedValues]` lists that name constants keep them in `DocValues` as
written references (`\Foo::BAR`, `\Foo::*`, `int-mask-of<\Foo::*>`, and
`\::NAME` for a global constant), since their values are known only once
every file is indexed. `docTypeRejects` reads any doc type naming constants
through `index.ConstantValues`, which `index.ResolveDocType` shares its
wildcard matching with, and only against an argument of exact values.

## Generic arguments, shapes and keys

`types.Accepts` compares generic arguments (by template variance, mapped to
an ancestor's templates through `indexRelations.TypeArgsAs`, which binds
them as `infer.Env.ClassArgBindings` does), array and object shapes,
callable signatures and array keys (int keys where only string keys are
documented). `declaredTypeArgs` drops the generic arguments of a value
unless a declaration states them (`typeArgsDeclared`), since inference may
fill them from a template's bound.

## Corpus triage

The corpora (guzzle, composer, laravel) give no proven finding. The
documented findings on laravel are PHPDoc contracts the code breaks while
PHP accepts the call (untyped parameters): an array to `@param string`,
null to `@param string` and `@param callable`, an empty array to a
required-key shape, a `RedisCluster` to `@param \Redis`.
