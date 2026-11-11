# DocumentedReturnType implementation notes

The check is shared with InvalidReturnType (see its NOTES): this rule
reports the mismatches of `semanticquery.ReturnTypeMismatch` that are not
proven.

## Value details

The `@return` type is parsed from the declaration's own PHPDoc, so it keeps
integer ranges, string lengths and literal values. They count only when
every int member of the value lists single values and every string member
lists its values; otherwise the tag is widened to its plain atoms
(`Type.DocString`, parsed again). Ordinary inference never tracks string
values, so in practice only integers such as `count()` of a literal array
are compared exactly.

## Corpus triage

The corpora (guzzle, composer, laravel) give seven documented findings, all
on laravel:

- Real contract breaks: `return [];` against a `@return array{…}` shape
  with required keys (three methods), a `@return array` method returning
  its string argument, a `@return int` command handler returning the null
  result of a closure, and `View::renderSections()` (`@return array`)
  returning `render()`, whose `@return string` ignores that it returns the
  callback's value.
- False positive: `TestResponse::streamedContent()` returns
  `$this->streamedContent` after `$this->sendContent()` fills it through an
  output buffer callback. The PHPDoc-aware environment keeps the `null`
  narrowing of the property across calls (the native environment drops
  it), so the value reads as null.
