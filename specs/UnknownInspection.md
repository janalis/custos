---
id: UnknownInspection
group: Code style
kind: syntax
needs: []
php: { min: "", max: "" }
---

# UnknownInspection

## Summary
A `@noinspection` suppression naming an inspection that does not exist (typo,
renamed or removed rule) silently suppresses nothing. Point it out so it can be
corrected or deleted.

## Detection
- **D1** Scan block comments — both doc comments `/** ... */` and plain block
  comments `/* ... */` — for the tag `@noinspection` (exact, case-sensitive).
  Line comments (`//`, `#`) are not scanned. Each occurrence of the tag in a
  comment is analysed separately.
- **D2** Tag value: the text after `@noinspection` up to the end of that line
  (or the comment terminator `*/`, whichever comes first), trimmed.
- **D3** Candidates: every identifier-like word (`[A-Za-z_][A-Za-z0-9_]*`) in
  the tag value. Separators (commas, spaces, `-`) and trailing free text are
  allowed; e.g. `@noinspection FooInspection, PhpBar - explanation` yields
  `FooInspection`, `PhpBar`, `explanation`.
- **D4** A candidate is *relevant* only if it starts with `Php` or ends with
  `Inspection` or `Inspector` (all case-sensitive). Other words (`SqlResolve`,
  `explanation`, …) are ignored.
- **D5** A relevant candidate is *unknown* if neither the candidate itself nor
  the candidate followed by `Inspection` is in the known-inspection set:
  - every `legacyId` from `internal/meta/rules.json` (e.g.
    `UnknownInspectionInspection`) and every custos rule ID;
  - a bundled list of JetBrains PhpStorm inspection short names (names of the
    IDE's own inspections, which users legitimately suppress). It must at least
    contain: `PhpUnused`, `PhpMultipleClassDeclarationsInspection`,
    `PhpMissingParentCallCommonInspection`, `PhpUndefinedMethodInspection`,
    `PhpUndefinedFieldInspection`, `PhpUndefinedClassInspection`,
    `PhpUndefinedFunctionInspection`, `PhpUndefinedVariableInspection`,
    `PhpUndefinedConstantInspection`, `PhpUndefinedNamespaceInspection`,
    `PhpUndefinedClassConstantInspection`, `PhpUnusedParameterInspection`,
    `PhpUnusedLocalVariableInspection`, `PhpUnusedPrivateMethodInspection`,
    `PhpUnusedPrivateFieldInspection`, `PhpUnusedAliasInspection`,
    `PhpDocSignatureInspection`, `PhpDocMissingThrowsInspection`,
    `PhpDocRedundantThrowsInspection`, `PhpUnhandledExceptionInspection`,
    `PhpIncludeInspection`, `PhpDeprecationInspection`,
    `PhpComposerExtensionStubsInspection`, `PhpFullyQualifiedNameUsageInspection`,
    `PhpUnnecessaryFullyQualifiedNameInspection`, `PhpParamsInspection`,
    `PhpMethodParametersCountMismatchInspection`,
    `PhpPossiblePolymorphicInvocationInspection`,
    `PhpDynamicAsStaticMethodCallInspection`, `PhpVoidFunctionResultUsedInspection`,
    `PhpExpressionResultUnusedInspection`, `PhpStatementHasEmptyBodyInspection`,
    `PhpInconsistentReturnPointsInspection`, `PhpMissingReturnTypeInspection`,
    `PhpMissingParamTypeInspection`, `PhpMissingFieldTypeInspection`,
    `PhpTypedPropertyMightBeUninitializedInspection`, `PhpStrictTypeCheckingInspection`,
    `PhpRedundantOptionalArgumentInspection`, `PhpUnreachableStatementInspection`,
    `PhpMissingBreakStatementInspection`, `PhpAssignmentInConditionInspection`,
    `PhpIllegalArrayKeyTypeInspection`, `PhpIllegalPsrClassPathInspection`,
    `PhpConditionAlreadyCheckedInspection`, `PhpUnnecessaryLocalVariableInspection`,
    `PhpRedundantVariableDocTypeInspection`, `PhpVariableVariableInspection`,
    `PhpSameParameterValueInspection`, `PhpMissingParentConstructorInspection`,
    `PhpRedundantClosingTagInspection`, `PhpCSValidationInspection`,
    `PhpPureAttributeCanBeAddedInspection`, `PhpMissingDocCommentInspection`.
    Keep the list in a data file so it can grow.
- **D6** Report the tag when at least one relevant candidate is unknown.

## Exceptions (no report)
- **E1** All relevant candidates known (custos/EA legacy IDs, PhpStorm names).
- **E2** No relevant candidate at all (only non-`Php*`, non-`*Inspection`,
  non-`*Inspector` names such as SQL or other-language inspection names).
- **E3** `@noinspection` in a line comment or in a string.

## Report
- Range: the `@noinspection` tag text itself (13 characters, from `@` to the
  final `n`).
- Severity: info.
- Message: `Suppressed inspection not recognised: {names}.` where `{names}` is
  the list of unknown candidates joined by `, ` in source order, each listed once.
- One report per tag, even if several names are unknown.

## Fix
None.

## Options
None.

## PHP versions
None.

## Examples

```php
<?php
function audit() {
    /** <weak_warning descr="Suppressed inspection not recognised: PhpLegacyCheckInspector.">@noinspection</weak_warning> PhpLegacyCheckInspector */
    /* <weak_warning descr="Suppressed inspection not recognised: TypoNameInspection.">@noinspection</weak_warning> TypoNameInspection */
    /** <weak_warning descr="Suppressed inspection not recognised: PhpGoneAway.">@noinspection</weak_warning> PhpUnused, PhpGoneAway -- kept for BC */

    /** @noinspection PhpUndefinedMethodInspection */
    /** @noinspection UnnecessarySemicolonInspection */
    /** @noinspection UnnecessarySemicolon */
    /** @noinspection SqlNoDataSourceInspection2 */
    /** @noinspection JSUnresolvedReference */
    // @noinspection PhpWhatever
}
```

## Divergences
- Upstream compares against every inspection registered in the running IDE
  (PhpStorm built-ins and all installed plugins). custos cannot know those, so
  it uses its own catalogue plus a curated PhpStorm list (D5). Names from other
  plugins that start with `Php` or end with `Inspection`/`Inspector` may be
  reported by custos where the IDE would not. Recommendation: keep the curated
  list editable and accept user additions through configuration later.
- Upstream extracts candidates either from the IDE's parsed references or, as a
  fallback, with a "two or more capitalised words" pattern; D3 (identifier
  words) gives the same results on all fixtures and handles names with
  consecutive capitals better.
- Message order: upstream joins names in hash order; custos uses source order.
- Optional extension (not upstream): also validate IDs in `@custos-ignore`.
- **Wider PhpStorm list (custos).** Beyond the names this spec requires,
  the embedded list holds further PhpStorm short names seen in real code
  (`PhpIncompatibleReturnTypeInspection`,
  `ArrayTypeOfParameterByDefaultValueInspection`,
  `PhpNamedArgumentsWithChangedOrderInspection`,
  `AdditionOperationOnArraysInspection`, … — Craft CMS), so suppressions
  of real PhpStorm inspections are not reported as unknown.
