# DocumentedReturnType implementation notes

The check is shared with InvalidReturnType (see its NOTES): this rule
reports the mismatches of `semanticquery.ReturnTypeMismatch` that are not
proven.

## Value details

The `@return` type is parsed from the declaration's own PHPDoc, so it keeps
integer ranges, string lengths and literal values. They count only when
every int member of the value lists single values and every string member
lists its values or single byte lengths (`semanticquery.exactValues`);
otherwise the tag is widened to its plain atoms (`Type.DocString`, parsed
again). Inference gives an integer literal its value, a string literal its
length and, up to 64 bytes, its value, and `count()` of a literal array its
size.

## Corpus triage

The corpora (guzzle, composer, laravel) give six documented findings, all
on laravel, all real contract breaks: `return [];` against a
`@return array{…}` shape with required keys (three methods), a
`@return array` method returning its string argument, a `@return int`
command handler returning the null result of a closure, and
`View::renderSections()` (`@return array`) returning `render()`, whose
`@return string` ignores that it returns the callback's value.

`TestResponse::streamedContent()` returns `$this->streamedContent` after
`$this->sendContent()` fills it through an output buffer callback: a
condition on a property ends at a call that may run project code, so the
value is not read as null there.
