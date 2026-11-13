# InvalidPropertyAssignment implementation notes

## Shared check

The rule and DocumentedPropertyAssignment report the two outcomes of
`semanticquery.PropertyWriteMismatches`, cached per node (assignment,
foreach, `++`/`--`) through `analysis.Context.Shared`. Every write form
becomes a `propertyStore` (target, writing node, reported range) judged by
`judgeStore` with a function typing the stored value in a given
environment, so the PHPDoc-aware and proven passes share it. For a plain
assignment, the query:

1. keeps assignments whose target is a named property fetch
   (`propertyTarget`) before typing anything, so local assignments cost no
   inference;
2. types the stored value with the default (PHPDoc-aware) environment: the
   right-hand side for `=` and `??=`, the assignment itself (the compound
   result) otherwise;
3. resolves the receiver with `infer.Env.ProvenTypeOf`, so a receiver typed
   only by PHPDoc is never trusted, and lists the declarations with
   `propertyWrite.targets`: `index.FindProperty` on each class, then on
   every indexed subclass (`ChildrenAll`) unless the property is private or
   final or the class final; every class must be declared once with a
   resolved hierarchy and free of `__set`;
4. judges each declaration with `types.AllRejected` against the native type
   and then the `@var` type (`argumentCheck.docTypeRejects`, shared with
   the argument rules, so templates and stale tags are skipped alike), in
   the file's strict mode with `Internal` unset: property writes coerce as
   user-function parameters do;
5. asks `ProvenTypeOf` on the stored value only when every native type
   rejects the default type, and marks the finding proven when they reject
   the proven type too.

## Other write forms

`property_write_forms.go` holds them, apart from the assignment code:

- destructuring walks the pattern and types each item from the
  destructured value's type (`destructuredItem`); the local-variable rules
  of inference are not reused, as they ignore the null a missing key or a
  scalar source stores;
- foreach keys and values use `infer.Env.ForeachKeyType` and
  `ForeachValueType`, the types local foreach targets get, except that an
  iterator with an undeclared key type gives an unknown key rather than
  `int|string`;
- `++`/`--` use `infer.Env.IncDecStoredType`, the value prefix forms
  evaluate to; only `null` incremented is ever rejected in practice;
- a by-reference assignment is judged on the referenced local's current
  type: PHP verifies it at binding (`zend_verify_prop_assignable_by_ref`),
  coercing it in weak mode.

## Readonly properties

A readonly write is checked only as `$this->p = …` in the constructor of
the class owning the property, when it is the only `$this->p` write there
(`storeTargets` counts every form) and the property is not promoted, and
the write does not read the property or bind a reference (compound,
`++`/`--`, `=&`): PHP then fails on the uninitialized or readonly property
first. Any other readonly write may be a
re-initialization, whose error PHP raises before the type check.

## Builtin properties

Properties PHP declares (`Exception::$line`) are skipped: their stub types
were never compared with reflection, unlike builtin parameters
(`stub_type_drift.txt`).

## Corpus triage

The corpora (guzzle, composer, laravel) give no finding, through any write
form.
