# PHP and extension rule expansion audit

This expansion adds 100 native inspections, bringing the catalogue to 478 rules,
including 300 native inspections. Separate agents authored specifications and
implemented them from those specifications. No upstream implementation or fixture
was used to design this expansion. Native rules have no EA counterpart, so upstream
conformance is not applicable.

## Supported proofs and safe fixes

Detection depends on the explicit contract in each linked rule page. Unknown
values, escaped identities, unsupported control flow and exhausted proof budgets
cannot establish a finding. Intent-sensitive rules are disabled by default;
enabling one declares the policy in its specification.

Response proofs count bytes and discard unknown output or transformations. Numeric
proofs use exact decimal arithmetic where needed and bounded binary format parsers.
Extension rules resolve the applicable builtins and locally proven object identity.
ZIP budget validation requires both a configured entry-count bound and a full
metadata traversal that rejects failures, negative sizes and remaining-budget
overshoot before accumulation. Retrieving metadata alone does not prove a bound.
Fixes preserve evaluation count and require an unambiguous repair; policy choices
and runtime-dependent repairs remain manual.

## Corrections found during review

- HTTP checks use the correct context argument for each stream API and discard
  context facts after intervening calls. Compression findings require emitted
  output and exclude prior buffering or unknown transformation operations. Redirect
  and CGI status headers invalidate partial-response proofs because PHP can change
  the status implicitly.
- Header filename fixes escape both HTTP quoted values and the enclosing PHP
  string. Exact expected outputs cover spaces, apostrophes and dollar signs.
- Numeric proofs reject malformed signed zero strings and use exact rational
  arithmetic for configured fractional precision. Binary format directives and
  offset arguments follow their applicable PHP version thresholds.
- Builtin class provenance comes from the bundled stubs. Attribute flags are
  interpreted for the selected PHP version. Never-return proofs handle resolved
  method and static delegation and exclude unsupported compound termination.
- Memory and temporary stream wrappers change effective opening modes; checks
  use their verified effective modes rather than assuming the literal mode.
- DOM ownership, internal XML entity declarations, result finalization and
  multiple SQLite row consumers have dedicated safety regressions. SQLite mode
  fixes are withheld when another consumer needs numeric or mixed keys.
- ZIP metadata retrieval alone was insufficient. The final check requires both
  count and overflow-safe accumulated-size guards; opaque external validators
  cannot establish a definite violation.

## Follow-up safety audit

The [false-positive and quick-fix audit](./php-extension-safety-audit.md) records
subsequent mutation, branch, precedence, namespace and version corrections. Five
intent-sensitive rules are now opt-in. See that audit for current validation and
limits; the figures below describe the initial implementation verification.

## Verification

- `make verify` passes: architecture, pinned linters, full tests, own fixtures,
  both 100% statement-coverage gates and the clean-room scan.
- The binary builds and the VitePress documentation site builds. CLI analysis
  confirms a new filename finding and its available fix.
- All 28 expected fixed PHP fixtures pass PHP 8.5 syntax linting.
- The complete 440-file expansion fixture corpus accepts 713 individual fixes
  and combined fixes on 272 files without increasing syntax errors.
- A separate 2613-file incomplete-input corpus produces no internal rule errors.
- New shared-query allocation benchmarks were run. They measure bounded local
  proof and response scans; benchmark timings include engine work and are not
  independent hard latency guarantees.

## Rule matrix

| Rule | Default | Fix |
| --- | --- | --- |
| [AbstractClassInstantiation](/rules/probable-bugs/AbstractClassInstantiation) | on | none |
| [TraitInstantiation](/rules/probable-bugs/TraitInstantiation) | on | none |
| [EnumInstantiation](/rules/probable-bugs/EnumInstantiation) | on | none |
| [OverrideAttributeWithoutTarget](/rules/probable-bugs/OverrideAttributeWithoutTarget) | on | none |
| [AttributeTargetMismatch](/rules/probable-bugs/AttributeTargetMismatch) | on | none |
| [NonRepeatableAttributeRepeated](/rules/probable-bugs/NonRepeatableAttributeRepeated) | on | none |
| [NeverFunctionFallsThrough](/rules/probable-bugs/NeverFunctionFallsThrough) | on | none |
| [FailedClosureBindingInvoked](/rules/probable-bugs/FailedClosureBindingInvoked) | on | none |
| [ReflectionNamedArgumentMismatch](/rules/probable-bugs/ReflectionNamedArgumentMismatch) | on | none |
| [ReflectionCompositeTypeAssumedNamed](/rules/probable-bugs/ReflectionCompositeTypeAssumedNamed) | on | none |
| [HeapIterationConsumesCollection](/rules/probable-bugs/HeapIterationConsumesCollection) | off | none |
| [PriorityQueueExtractionShapeMismatch](/rules/probable-bugs/PriorityQueueExtractionShapeMismatch) | on | none |
| [PriorityQueueMinOrderAssumption](/rules/probable-bugs/PriorityQueueMinOrderAssumption) | off | none |
| [LinkedListDeleteIterationUnexpected](/rules/probable-bugs/LinkedListDeleteIterationUnexpected) | off | none |
| [EmptySplCollectionExtraction](/rules/probable-bugs/EmptySplCollectionExtraction) | on | none |
| [SplFixedArrayIndexOutOfBounds](/rules/probable-bugs/SplFixedArrayIndexOutOfBounds) | on | none |
| [SplFixedArrayShrinkDiscardsValues](/rules/probable-bugs/SplFixedArrayShrinkDiscardsValues) | off | none |
| [SplObjectStorageIndirectArrayMutation](/rules/probable-bugs/SplObjectStorageIndirectArrayMutation) | on | none |
| [WeakMapValueRetainsKey](/rules/probable-bugs/WeakMapValueRetainsKey) | off | none |
| [RecursiveTraversalOmitsDirectories](/rules/probable-bugs/RecursiveTraversalOmitsDirectories) | off | none |
| [SeekSuccessCheckedAsTruthy](/rules/probable-bugs/SeekSuccessCheckedAsTruthy) | off | none |
| [AppendModeSeekUsedForOverwrite](/rules/probable-bugs/AppendModeSeekUsedForOverwrite) | on | none |
| [TruncationAssumedToRewind](/rules/probable-bugs/TruncationAssumedToRewind) | on | none |
| [StreamOutputReturnUsedAsContent](/rules/probable-bugs/StreamOutputReturnUsedAsContent) | on | available |
| [ReadFromWriteOnlyStream](/rules/probable-bugs/ReadFromWriteOnlyStream) | on | none |
| [WriteToReadOnlyStream](/rules/probable-bugs/WriteToReadOnlyStream) | on | none |
| [CsvDelimiterInvalidByteLength](/rules/probable-bugs/CsvDelimiterInvalidByteLength) | on | none |
| [CsvEscapeDefaultDependency](/rules/probable-bugs/CsvEscapeDefaultDependency) | on | available |
| [CsvBlankRecordShapeMismatch](/rules/probable-bugs/CsvBlankRecordShapeMismatch) | on | none |
| [ReaddirFalsyFilenameLoss](/rules/probable-bugs/ReaddirFalsyFilenameLoss) | on | available |
| [ContentLengthKnownBodyMismatch](/rules/probable-bugs/ContentLengthKnownBodyMismatch) | on | none |
| [ContentEncodingBodyMismatch](/rules/probable-bugs/ContentEncodingBodyMismatch) | on | none |
| [ConflictingContentLengthHeaders](/rules/probable-bugs/ConflictingContentLengthHeaders) | on | none |
| [TransferEncodingWithContentLength](/rules/probable-bugs/TransferEncodingWithContentLength) | on | none |
| [ContentRangeLengthMismatch](/rules/probable-bugs/ContentRangeLengthMismatch) | on | none |
| [CorsOriginListInvalid](/rules/probable-bugs/CorsOriginListInvalid) | on | none |
| [CorsSubdomainWildcardInvalid](/rules/probable-bugs/CorsSubdomainWildcardInvalid) | on | none |
| [VaryHeaderOverwritesEarlierDimensions](/rules/probable-bugs/VaryHeaderOverwritesEarlierDimensions) | off | none |
| [ContentDispositionFilenameNeedsQuoting](/rules/probable-bugs/ContentDispositionFilenameNeedsQuoting) | on | available |
| [HttpErrorBodyReadWithoutIgnoreErrors](/rules/probable-bugs/HttpErrorBodyReadWithoutIgnoreErrors) | off | none |
| [BcMathFloatOperand](/rules/probable-bugs/BcMathFloatOperand) | off | none |
| [BcMathExponentNotation](/rules/probable-bugs/BcMathExponentNotation) | on | none |
| [BcMathScaleDiscardsRequiredFraction](/rules/probable-bugs/BcMathScaleDiscardsRequiredFraction) | off | none |
| [GmpAutomaticBaseChangesDecimalInput](/rules/probable-bugs/GmpAutomaticBaseChangesDecimalInput) | off | available |
| [GmpDivisionPairUsedAsNumber](/rules/probable-bugs/GmpDivisionPairUsedAsNumber) | on | none |
| [GmpDivisionByKnownZero](/rules/probable-bugs/GmpDivisionByKnownZero) | on | none |
| [LogarithmInvalidRealDomain](/rules/probable-bugs/LogarithmInvalidRealDomain) | on | none |
| [InverseTrigInvalidRealDomain](/rules/probable-bugs/InverseTrigInvalidRealDomain) | on | none |
| [PackFormatArgumentMismatch](/rules/probable-bugs/PackFormatArgumentMismatch) | on | none |
| [UnpackBufferTooShort](/rules/probable-bugs/UnpackBufferTooShort) | on | none |
| [DomCrossDocumentAppend](/rules/probable-bugs/DomCrossDocumentAppend) | on | none |
| [DomAppendMovesNodeUnexpectedly](/rules/probable-bugs/DomAppendMovesNodeUnexpectedly) | off | none |
| [DomShallowImportDropsDescendants](/rules/probable-bugs/DomShallowImportDropsDescendants) | off | available |
| [DomElementValueContainsBareAmpersand](/rules/probable-bugs/DomElementValueContainsBareAmpersand) | on | none |
| [DomMissingAttributeCheckedAsNull](/rules/probable-bugs/DomMissingAttributeCheckedAsNull) | on | none |
| [DomTextAssignmentDestroysChildren](/rules/probable-bugs/DomTextAssignmentDestroysChildren) | off | none |
| [DomLiveNodeListRemovalSkipsNodes](/rules/probable-bugs/DomLiveNodeListRemovalSkipsNodes) | on | none |
| [XPathQueryFailureDereferenced](/rules/probable-bugs/XPathQueryFailureDereferenced) | on | none |
| [XmlReaderAttributeCursorNotRestored](/rules/probable-bugs/XmlReaderAttributeCursorNotRestored) | off | none |
| [UntrustedXmlWriterRawContent](/rules/security/UntrustedXmlWriterRawContent) | on | none |
| [SqliteBindingIndexZero](/rules/probable-bugs/SqliteBindingIndexZero) | on | none |
| [SqliteClearWithoutRequiredReset](/rules/probable-bugs/SqliteClearWithoutRequiredReset) | on | none |
| [SqliteNullBindingDiscardsValue](/rules/probable-bugs/SqliteNullBindingDiscardsValue) | on | none |
| [SqliteFetchBothLeaksDuplicateColumns](/rules/probable-bugs/SqliteFetchBothLeaksDuplicateColumns) | off | available |
| [SqliteResultUsedAfterFinalize](/rules/probable-bugs/SqliteResultUsedAfterFinalize) | on | none |
| [PgEscapedLiteralQuotedAgain](/rules/probable-bugs/PgEscapedLiteralQuotedAgain) | on | none |
| [PgIdentifierEscapedAsLiteral](/rules/probable-bugs/PgIdentifierEscapedAsLiteral) | on | available |
| [PgAsyncDispatchAssumedQuerySuccess](/rules/probable-bugs/PgAsyncDispatchAssumedQuerySuccess) | off | none |
| [PgAsyncResultsNotDrained](/rules/probable-bugs/PgAsyncResultsNotDrained) | on | none |
| [PgFetchedZeroRejected](/rules/probable-bugs/PgFetchedZeroRejected) | on | available |
| [UnicodeNormalizationFailureConsumed](/rules/probable-bugs/UnicodeNormalizationFailureConsumed) | on | none |
| [LocalizedNumberTrailingInputAccepted](/rules/probable-bugs/LocalizedNumberTrailingInputAccepted) | off | none |
| [NumberParseUsesCurrencyType](/rules/probable-bugs/NumberParseUsesCurrencyType) | on | none |
| [IntlCalendarSecondsPassedAsMilliseconds](/rules/probable-bugs/IntlCalendarSecondsPassedAsMilliseconds) | on | available |
| [IntlGregorianMonthOffByOne](/rules/probable-bugs/IntlGregorianMonthOffByOne) | off | none |
| [IntlTwelveHourFieldUsedForTwentyFourHourTime](/rules/probable-bugs/IntlTwelveHourFieldUsedForTwentyFourHourTime) | on | available |
| [IntlTimezoneOffsetMillisecondsUsedAsSeconds](/rules/probable-bugs/IntlTimezoneOffsetMillisecondsUsedAsSeconds) | on | available |
| [CollatorComparisonTruthinessReversed](/rules/probable-bugs/CollatorComparisonTruthinessReversed) | off | none |
| [CollatorSortDiscardsRequiredKeys](/rules/probable-bugs/CollatorSortDiscardsRequiredKeys) | off | available |
| [TransliterationFailureConsumed](/rules/probable-bugs/TransliterationFailureConsumed) | on | none |
| [CompressionDecoderFormatMismatch](/rules/probable-bugs/CompressionDecoderFormatMismatch) | on | available |
| [DecompressionFailureConsumed](/rules/probable-bugs/DecompressionFailureConsumed) | on | none |
| [IncrementalInflateEncodingMismatch](/rules/probable-bugs/IncrementalInflateEncodingMismatch) | on | none |
| [DeflateStreamNeverFinished](/rules/probable-bugs/DeflateStreamNeverFinished) | off | none |
| [CompressionLevelOutsideSupportedRange](/rules/probable-bugs/CompressionLevelOutsideSupportedRange) | on | none |
| [ZipFirstEntryTreatedAsMissing](/rules/probable-bugs/ZipFirstEntryTreatedAsMissing) | on | available |
| [ZipSourceRemovedBeforeArchiveClose](/rules/probable-bugs/ZipSourceRemovedBeforeArchiveClose) | on | none |
| [ZipEntryNameCollision](/rules/probable-bugs/ZipEntryNameCollision) | off | none |
| [ZipExtractionExceedsConfiguredBudget](/rules/probable-bugs/ZipExtractionExceedsConfiguredBudget) | off | none |
| [ZipEntryIndexPastEnd](/rules/probable-bugs/ZipEntryIndexPastEnd) | on | none |
| [ForkFailureTreatedAsParent](/rules/probable-bugs/ForkFailureTreatedAsParent) | on | none |
| [ForkChildFallsIntoParentWork](/rules/probable-bugs/ForkChildFallsIntoParentWork) | off | none |
| [WaitStatusUsedAsExitCode](/rules/probable-bugs/WaitStatusUsedAsExitCode) | off | none |
| [NonblockingWaitZeroTreatedAsReaped](/rules/probable-bugs/NonblockingWaitZeroTreatedAsReaped) | on | none |
| [ExitStatusReadWithoutNormalExitCheck](/rules/probable-bugs/ExitStatusReadWithoutNormalExitCheck) | on | none |
| [SignalHandlerHasNoDispatchMechanism](/rules/probable-bugs/SignalHandlerHasNoDispatchMechanism) | on | none |
| [AsyncSignalSetterReturnMisread](/rules/probable-bugs/AsyncSignalSetterReturnMisread) | off | available |
| [SelectWatchArraysNotRestored](/rules/probable-bugs/SelectWatchArraysNotRestored) | off | none |
| [SelectFailureTreatedAsReadiness](/rules/probable-bugs/SelectFailureTreatedAsReadiness) | on | none |
| [ConnectTimeoutAssumedToBoundReads](/rules/probable-bugs/ConnectTimeoutAssumedToBoundReads) | off | none |
