# DocumentedPropertyAssignment implementation notes

The check is shared with InvalidPropertyAssignment (see its NOTES): this
rule reports the mismatches of `semanticquery.PropertyWriteMismatches`
that are not proven.

## Value details

The index keeps `@var` types through `Type.DocString`, which drops integer
ranges, string lengths and literal string values; `Property.DocValues`
keeps the integer ranges, a non-empty string and literal string values,
judged only against exact
values as for DocumentedArgumentType (see its NOTES; magic `@property`
tags have none). `DocType` turns a template nested in a generic type into
`mixed` (`array<int, T>` reads `array<int,mixed>`), which still rejects a
non-array. A bare template
(`@var T`) keeps its template marker and is skipped.

## Corpus triage

The corpora (guzzle, composer, laravel) give no finding, the write forms
beyond assignments included. The five findings laravel gave before
untyped properties without a default counted as null-initialized all
stored `null` in such a property whose `@var` omits null
(`Model::$resolver`, `Blueprint::$after`, `DateFactory::$callable` three
times); PHP accepts them.
