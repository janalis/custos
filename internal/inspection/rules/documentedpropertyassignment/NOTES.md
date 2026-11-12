# DocumentedPropertyAssignment implementation notes

The check is shared with InvalidPropertyAssignment (see its NOTES): this
rule reports the mismatches of `semanticquery.PropertyAssignmentMismatch`
that are not proven.

## Value details

The index keeps `@var` types through `Type.DocString`, which drops integer
ranges, string lengths and literal string values; `Property.DocValues`
keeps the integer ranges and a non-empty string, judged only against exact
values as for DocumentedArgumentType (see its NOTES; magic `@property`
tags have none). `DocType` turns a template nested in a generic type into
`mixed` (`array<int, T>` reads `array<int,mixed>`), which still rejects a
non-array. A bare template
(`@var T`) keeps its template marker and is skipped.

## Corpus triage

The corpora (guzzle, composer, laravel) give no proven finding. The six
documented findings on laravel all store `null` in an untyped property
whose `@var` omits null: `Model::$resolver`, `Blueprint::$after`,
`Cloud\Queue::$processingJobStartedAt` and `DateFactory::$callable`
(three resets). PHP accepts them; the tags are incomplete.
