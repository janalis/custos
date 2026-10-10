# Native rule expansion audit

The expansion adds 100 independently specified native inspections to the existing
catalogue, bringing it to 378 rules, including 200 native inspections. Specification
and implementation were performed by separate agents. No EA implementation or
fixtures were used to design these native rules.

Every new specification and its own marked example was reviewed. Additional
implementation reviews examined source provenance, aliases, execution state,
mutation contexts, response headers and fix edits. The matrix below records the
contract reviewed for each ID; detection is limited to the precise supported
subset in its specification, and unsupported proof produces no finding.

## Findings and corrections

- Result-array inspections must distinguish reads from writes, unset,
  reference passing and same-statement mutation. Regex results additionally
  depend on the requested offset; unsupported query-name normalization cannot
  establish an absent key.
- Array inspections must evaluate surviving entries after duplicate-key
  overwrites. A singleton object array does not require string comparison in
  `array_unique`. Numeric-key slice proofs additionally exclude duplicate or
  unknown effective keys because literal positions can differ from the
  effective selected slice.
- Reference and exception inspections require surviving bindings and reachable
  paths. Empty foreach input, body rebinding, terminating catches, finally
  restoration, handled exceptions and dead or delegated yields invalidate proof.
  Exception objects with custom constructors or destructors can perform
  intentional work and are excluded from the discarded-exception finding.
- Repeated decoding of parsed query values requires a direct standalone call,
  plain assignment, return, or single-expression echo. Compound expressions
  and wrappers can replace the parsed output before the decoding call and are
  excluded.
- Request origin must survive earlier explicit superglobal writes, reference
  escapes, GLOBALS targets and unknown calls, including zero-argument, method
  and static calls, plus earlier object construction. Only the supported
  resolved pure builtins are call exceptions.
  Source tracking excludes nested scopes and stops when its bounded analysis
  cannot establish the contract.
- cURL findings require final effective configuration. Empty credentials,
  non-Basic authentication, later method changes and unknown bulk setters
  cannot establish the proposed credential or request-mode hazards.
- Empty multi-transfer setup is valid completion; removed handles invalidate
  nonempty-transfer proof. Database execution with replacement parameters and
  non-result-producing queries invalidate stale-binding or outstanding-result
  assumptions. SQL assignments are distinct from null comparisons.
- A Latin-1 conversion can happen to produce valid UTF-8. Combining marks after
  control characters do not establish the supported grapheme-splitting case.
- XML errors cannot accumulate in a loop that never runs or clears/disables
  collection in its header. Nested parse or document-constructor argument calls
  can also alter collection state and discard proof. Explicit XXE protection is
  version-dependent;
  network blocking alone does not protect local external resources.
- Response MIME and CORS findings require the final supported header state.
  Later corrective or unknown response operations discard proof; image MIME
  detection is restricted to the final top-level encoder statement and excludes
  explicit output-buffer callbacks that can transform its bytes. Its prior-syntax
  scan is limited to 4096 nodes; exhausted proof cannot establish a finding.
- Ignored callback returns establish a pure transform only when supported
  string builtin arguments are definitely strings. Untyped object conversion
  can perform intentional work and is excluded.
- Fixes preserve evaluation count and retained source bytes. Line, block and
  `#` comments that would be discarded suppress the fix. No error-handling,
  encryption derivation or application policy is invented by a fix.

The initial manual harness reproduced 38 failing cases across 26 rule IDs,
including a missed HEAD-mode reset and unsafe fix-count behavior as well as
false positives. All 43 final harness cases pass. The persisted audit tests
contain 85 table cases across 32 inspection packages, with 14 additional
source-helper cases and further shared AST-query helper tests. The categories above describe the corrected behaviors;
they are not counts of individual defects.

## Unsafe fix boundaries

| Rule | Reproduced problem | Correction |
| --- | --- | --- |
| ParseStrKeyNormalizationMismatch | A normalized-key edit changed a writable array access; leading query whitespace also produced the wrong replacement key. | Offer edits only for stable reads and the supported normalization subset. |
| ZipOpenTruthyResult | Rewriting a negated condition discarded an embedded PHP hash comment. | Retain the finding and omit edits that would discard comments. |
| PregQuoteUsedOnReplacement | Removing the quoting wrapper discarded an embedded PHP hash comment. | Retain the finding and omit edits that would discard comments. |

PHP syntax checks complement the explicit expected-edit fixtures and mutation/context
regressions. Parsable output alone does not establish behavioral correctness.

## Per-rule review matrix

| Rule | Default | Fix | Contract reviewed |
| --- | --- | --- | --- |
| [ForeachReferenceSurvivesLoop](/rules/probable-bugs/ForeachReferenceSurvivesLoop) | on | none | Live reference after iteration |
| [LoopClosureCapturesReference](/rules/probable-bugs/LoopClosureCapturesReference) | off | none | Deferred same-scope counter reads |
| [ArrayCopyRetainsReferences](/rules/probable-bugs/ArrayCopyRetainsReferences) | off | none | Shared element reference identities |
| [ArrayFillSharesObject](/rules/probable-bugs/ArrayFillSharesObject) | off | none | Shared object entry mutations |
| [ExceptionConstructedNotThrown](/rules/probable-bugs/ExceptionConstructedNotThrown) | on | none | Discarded Throwable construction |
| [UnreachableCatchClause](/rules/probable-bugs/UnreachableCatchClause) | on | none | Resolved catch type coverage |
| [FinallyThrowMasksException](/rules/probable-bugs/FinallyThrowMasksException) | off | none | Uncaught exception propagation |
| [CatchVariableOverwritesLocal](/rules/probable-bugs/CatchVariableOverwritesLocal) | off | none | Catch/finally binding restoration |
| [GeneratorReturnBeforeCompletion](/rules/probable-bugs/GeneratorReturnBeforeCompletion) | on | none | Initial direct yielding state |
| [FiberResumeBeforeStart](/rules/probable-bugs/FiberResumeBeforeStart) | on | none | Owned fiber lifecycle |
| [ArrayUnionDropsRightValues](/rules/probable-bugs/ArrayUnionDropsRightValues) | off | none | List-key collisions and value iteration |
| [RecursiveReplaceRetainsListTail](/rules/probable-bugs/RecursiveReplaceRetainsListTail) | off | none | Retained tail and immediate access |
| [ArraySliceDiscardsRequiredKeys](/rules/probable-bugs/ArraySliceDiscardsRequiredKeys) | off | none | Reindexed keys and actual reads |
| [ArraySpliceDiscardsReplacementKeys](/rules/probable-bugs/ArraySpliceDiscardsReplacementKeys) | off | none | Discarded map keys and actual reads |
| [ArrayColumnDuplicateIndexLoss](/rules/probable-bugs/ArrayColumnDuplicateIndexLoss) | on | none | Effective rows and duplicate indices |
| [ArrayWalkCallbackReturnIgnored](/rules/probable-bugs/ArrayWalkCallbackReturnIgnored) | on | none | Ignored pure callback returns |
| [ArrayMultisortLengthMismatch](/rules/probable-bugs/ArrayMultisortLengthMismatch) | on | none | Known parallel array lengths |
| [ArrayRandKeyUsedAsValue](/rules/probable-bugs/ArrayRandKeyUsedAsValue) | off | none | Required key-versus-value contracts |
| [ArrayDiffOnNestedArrays](/rules/probable-bugs/ArrayDiffOnNestedArrays) | on | none | Surviving nested-array values |
| [ArrayUniqueOnNonStringableObjects](/rules/probable-bugs/ArrayUniqueOnNonStringableObjects) | on | none | Surviving object comparison requirements |
| [JsonNumericCheckChangesIdentifiers](/rules/probable-bugs/JsonNumericCheckChangesIdentifiers) | off | none | Surviving formatted identifier strings |
| [JsonPartialOutputOverridesThrow](/rules/probable-bugs/JsonPartialOutputOverridesThrow) | on | none | Conflicting JSON error policies |
| [JsonForceObjectChangesNestedLists](/rules/probable-bugs/JsonForceObjectChangesNestedLists) | off | none | Surviving nested list shapes |
| [JsonEncodeFlagPassedToDecode](/rules/probable-bugs/JsonEncodeFlagPassedToDecode) | on | none | Resolved decoding flag names |
| [UnserializeFalseAmbiguity](/rules/probable-bugs/UnserializeFalseAmbiguity) | on | none | Valid serialized false ambiguity |
| [Base64ValidationWithoutStrictMode](/rules/probable-bugs/Base64ValidationWithoutStrictMode) | on | none | Strict decoding as validation |
| [HexDecodeOddLengthLiteral](/rules/probable-bugs/HexDecodeOddLengthLiteral) | on | none | Known hexadecimal byte-pair lengths |
| [IntegerCastOutOfRangeLiteral](/rules/probable-bugs/IntegerCastOutOfRangeLiteral) | on | none | Portable integer-range overflow |
| [NanCheckedWithEquality](/rules/probable-bugs/NanCheckedWithEquality) | on | gated | Numeric NaN predicate fixes |
| [FractionalArrayKeyTruncation](/rules/probable-bugs/FractionalArrayKeyTruncation) | on | none | Fractional numeric key conversion |
| [OpenSslKeyLengthMismatch](/rules/security/OpenSslKeyLengthMismatch) | on | none | Cipher-specific key byte lengths |
| [PasswordUsedDirectlyAsEncryptionKey](/rules/security/PasswordUsedDirectlyAsEncryptionKey) | off | none | Untouched configured password sources |
| [AeadAuthenticationTagDiscarded](/rules/security/AeadAuthenticationTagDiscarded) | on | none | Required authenticated output tag |
| [AeadDecryptionTagLengthUnchecked](/rules/security/AeadDecryptionTagLengthUnchecked) | off | none | Untouched input and validated tag length |
| [DecryptionFailureUsedAsPlaintext](/rules/security/DecryptionFailureUsedAsPlaintext) | on | none | Guarded plaintext consumption |
| [OpenSslVerifyTruthyResult](/rules/security/OpenSslVerifyTruthyResult) | on | gated | Explicit signature verification success |
| [BcryptPasswordTruncation](/rules/security/BcryptPasswordTruncation) | on | none | Known bcrypt input byte length |
| [HardcodedCredentialAtKnownSink](/rules/security/HardcodedCredentialAtKnownSink) | off | none | Resolved credential sinks and excluded paths |
| [OpenSslRawCiphertextOptionMismatch](/rules/security/OpenSslRawCiphertextOptionMismatch) | on | none | Matching ciphertext encoding options |
| [UntrustedLdapFilter](/rules/security/UntrustedLdapFilter) | on | none | Untouched LDAP filter input and escaping |
| [RequestControlledSessionId](/rules/security/RequestControlledSessionId) | on | none | Untouched request session identifiers |
| [SessionDestroyLeavesLocalAuthentication](/rules/probable-bugs/SessionDestroyLeavesLocalAuthentication) | on | none | Stored versus local session state |
| [SessionRegenerationFailureUnchecked](/rules/security/SessionRegenerationFailureUnchecked) | off | none | Configured authentication establishment |
| [SessionNameChangedWhileActive](/rules/probable-bugs/SessionNameChangedWhileActive) | on | none | Actual active-session lifetime |
| [AuthenticationCookieAllowsScriptAccess](/rules/security/AuthenticationCookieAllowsScriptAccess) | off | none | Explicit authentication cookie policy |
| [HostCookiePrefixContractViolation](/rules/security/HostCookiePrefixContractViolation) | on | none | Host-prefixed cookie attributes |
| [CookieDeletionScopeMismatch](/rules/probable-bugs/CookieDeletionScopeMismatch) | on | none | Matching cookie deletion scope |
| [CredentialedCorsUsesWildcardOrigin](/rules/security/CredentialedCorsUsesWildcardOrigin) | on | none | Final credentialed CORS header state |
| [DynamicCorsOriginWithoutVary](/rules/probable-bugs/DynamicCorsOriginWithoutVary) | off | none | Untouched origin and shared-cache variation |
| [BasicAuthenticationOverPlainHttp](/rules/security/BasicAuthenticationOverPlainHttp) | on | none | Nonempty Basic credentials and final HTTPS state |
| [CurlHeadersIncludedInDecodedBody](/rules/probable-bugs/CurlHeadersIncludedInDecodedBody) | on | none | Separate headers and JSON body |
| [CurlMultipartBodyWithJsonContentType](/rules/probable-bugs/CurlMultipartBodyWithJsonContentType) | on | none | Multipart body versus declared JSON |
| [CurlCustomHeadWithoutNoBody](/rules/probable-bugs/CurlCustomHeadWithoutNoBody) | on | none | Final HEAD/no-body transfer mode |
| [CurlCustomGetRetainsPostBody](/rules/probable-bugs/CurlCustomGetRetainsPostBody) | off | none | Final GET/body transfer mode |
| [CurlMethodOptionOrderConflict](/rules/probable-bugs/CurlMethodOptionOrderConflict) | on | none | Conflicting ordered request modes |
| [CurlWriteCallbackMissingByteCount](/rules/probable-bugs/CurlWriteCallbackMissingByteCount) | on | none | Successful callback byte-count returns |
| [CurlExplicitInfiniteTimeout](/rules/probable-bugs/CurlExplicitInfiniteTimeout) | off | none | Final effective timeout configuration |
| [CurlMultiSuccessAssumedPerTransfer](/rules/probable-bugs/CurlMultiSuccessAssumedPerTransfer) | on | none | Nonempty multi-transfer completion results |
| [CurlHeaderListOverwritten](/rules/probable-bugs/CurlHeaderListOverwritten) | off | none | Replacement header-list loss |
| [CurlRedirectForwardsCredentials](/rules/security/CurlRedirectForwardsCredentials) | off | none | Nonempty credentials and redirect forwarding |
| [MysqliBoundArrayReassigned](/rules/probable-bugs/MysqliBoundArrayReassigned) | on | none | Existing bindings versus replacement parameters |
| [MysqliBindingTypeCountMismatch](/rules/probable-bugs/MysqliBindingTypeCountMismatch) | on | none | Binding argument/type count |
| [MysqliInvalidBindingType](/rules/probable-bugs/MysqliInvalidBindingType) | on | none | Supported binding type letters |
| [MysqliPlaceholderBindingMismatch](/rules/probable-bugs/MysqliPlaceholderBindingMismatch) | on | none | SQL placeholder lexer and binding count |
| [MysqliUnbufferedResultBlocksNextQuery](/rules/probable-bugs/MysqliUnbufferedResultBlocksNextQuery) | on | none | Outstanding result-producing query state |
| [PdoArrayBoundToSinglePlaceholder](/rules/probable-bugs/PdoArrayBoundToSinglePlaceholder) | on | none | Scalar placeholder binding contracts |
| [SqlNullComparedWithEquality](/rules/probable-bugs/SqlNullComparedWithEquality) | on | none | SQL predicate context versus assignments |
| [MysqlDdlImplicitlyCommitsTransaction](/rules/probable-bugs/MysqlDdlImplicitlyCommitsTransaction) | on | none | MySQL-specific transaction boundaries |
| [NestedPdoTransactionWithoutSavepoint](/rules/probable-bugs/NestedPdoTransactionWithoutSavepoint) | on | none | Owned transaction state |
| [PdoFetchBothLeaksDuplicateColumns](/rules/probable-bugs/PdoFetchBothLeaksDuplicateColumns) | off | none | Fetch representation during serialization |
| [ReadLoopProcessesEofFailure](/rules/probable-bugs/ReadLoopProcessesEofFailure) | on | none | Strict string consumer after failed read |
| [NonBlockingEmptyReadTreatedAsEof](/rules/probable-bugs/NonBlockingEmptyReadTreatedAsEof) | on | none | Nonblocking empty read versus EOF |
| [FileStatCacheAfterExternalMutation](/rules/probable-bugs/FileStatCacheAfterExternalMutation) | off | none | Known external mutation and stat cache |
| [RenameFailureUnchecked](/rules/probable-bugs/RenameFailureUnchecked) | off | none | Renaming result before success return |
| [PermissionModeWrittenInDecimal](/rules/probable-bugs/PermissionModeWrittenInDecimal) | off | none | Explicit permission notation policy |
| [WorldWritablePermission](/rules/security/WorldWritablePermission) | off | none | Configured sensitive path permissions |
| [SharedLockUsedForWriting](/rules/probable-bugs/SharedLockUsedForWriting) | on | none | Exclusive write-lock state |
| [ProcessPipesDrainedSequentially](/rules/probable-bugs/ProcessPipesDrainedSequentially) | off | none | Owned blocking output-pipe drainage |
| [ProcessClosedBeforeOwnedPipes](/rules/probable-bugs/ProcessClosedBeforeOwnedPipes) | on | none | Owned pipes before process waiting |
| [ExecOutputArrayAccumulates](/rules/probable-bugs/ExecOutputArrayAccumulates) | off | none | Independent command output collection |
| [PregMatchAllLayoutMismatch](/rules/probable-bugs/PregMatchAllLayoutMismatch) | on | none | Read-only indices and zero regex offset |
| [OptionalRegexCaptureReadWithoutNullFlag](/rules/probable-bugs/OptionalRegexCaptureReadWithoutNullFlag) | on | none | Read-only optional capture and zero offset |
| [MultibytePositionUsedAsByteOffset](/rules/probable-bugs/MultibytePositionUsedAsByteOffset) | on | none | Character indices versus byte offsets |
| [EncodingConversionArgumentsReversed](/rules/probable-bugs/EncodingConversionArgumentsReversed) | off | none | Actually invalid converted UTF-8 bytes |
| [CodePointSliceSplitsGrapheme](/rules/probable-bugs/CodePointSliceSplitsGrapheme) | off | none | Supported grapheme boundaries and controls |
| [PregQuoteUsedOnReplacement](/rules/probable-bugs/PregQuoteUsedOnReplacement) | on | gated | Literal replacement and comment-preserving fix |
| [CtypeIntegerInterpretedAsCharacterCode](/rules/probable-bugs/CtypeIntegerInterpretedAsCharacterCode) | off | none | Explicit character-code interpretation policy |
| [ParseStrKeyNormalizationMismatch](/rules/probable-bugs/ParseStrKeyNormalizationMismatch) | on | gated | Read-only normalized query keys |
| [ParsedQueryValueDecodedTwice](/rules/probable-bugs/ParsedQueryValueDecodedTwice) | on | none | Proven second URL-decoding change |
| [FormEncodingUsedForRfc3986Signature](/rules/probable-bugs/FormEncodingUsedForRfc3986Signature) | off | none | Configured signature encoding requirements |
| [ZipOpenTruthyResult](/rules/probable-bugs/ZipOpenTruthyResult) | on | gated | Explicit ZIP success and preserved comments |
| [ZipCloseFailureUnchecked](/rules/probable-bugs/ZipCloseFailureUnchecked) | off | none | Archive finalization before success return |
| [ZipEntryStreamUsedAfterArchiveClose](/rules/probable-bugs/ZipEntryStreamUsedAfterArchiveClose) | on | none | Archive-owned entry-stream lifetime |
| [UntrustedXmlExternalEntityExpansion](/rules/security/UntrustedXmlExternalEntityExpansion) | on | none | Untouched XML input and supported XXE flags |
| [LibxmlErrorBufferNeverCleared](/rules/probable-bugs/LibxmlErrorBufferNeverCleared) | off | none | Executable parsing loops and error collection state |
| [XPathIgnoresDefaultNamespace](/rules/probable-bugs/XPathIgnoresDefaultNamespace) | on | none | Unchanged namespaced document and simple XPath |
| [ImageEncoderContentTypeMismatch](/rules/probable-bugs/ImageEncoderContentTypeMismatch) | on | none | Final top-level response encoder and MIME state |
| [ImageDecodeFailureUnchecked](/rules/probable-bugs/ImageDecodeFailureUnchecked) | on | none | Guarded image decoding results |
| [IntlFormattingFailureUnchecked](/rules/probable-bugs/IntlFormattingFailureUnchecked) | on | none | Guarded formatting results |
| [SodiumNonceLengthMismatch](/rules/probable-bugs/SodiumNonceLengthMismatch) | on | gated | Known secretbox nonce length and exact fix |

## Evidence and verification

PHP runtime probes independently confirmed duplicate-key survival, leading
query-name whitespace handling, offset-dependent regex results, a Latin-1
conversion whose bytes remain valid UTF-8, and singleton object deduplication.
The relevant contracts were checked against the
[PHP array manual](https://www.php.net/manual/en/language.types.array.php),
[query parser manual](https://www.php.net/manual/en/function.parse-str.php),
[regex matching manual](https://www.php.net/manual/en/function.preg-match.php)
and [mysqli execution manual](https://www.php.net/manual/en/mysqli-stmt.execute.php).

The original 203 positive, negative and fix PHP fragments passed syntax checks.
Specification examples were then synchronized to the independently authored own
fixtures, including exact diagnostic ranges and expected fixed bytes. Markdown
lint passed after the specification refinements. Corpus fix validation passed:
427 individual fixes across 206 files and combined fixes on 166 files produced
zero broken outputs. The documentation production build also passed. The final
`make verify` run passed, including both 100% statement-coverage gates and the
clean-room scan. The production documentation build and corpus fix check passed
again after the final query-decoding, XML and output-buffer refinements.

| Final check | Result |
| --- | --- |
| Architecture and catalogue completeness | Passed |
| Lint, tests and own fixtures | Passed |
| Focused coverage for updated inspection guards | Passed: 100% |
| Whole inspection and production/shared-helper statement coverage | Passed: 100% |
| Documentation production build | Passed |
| Generated rule-doc freshness | Passed |
| Clean-room scan | Passed |
| Corpus fix validation | Passed: 427 individual fixes, 206 files, 166 combined-fix files, zero broken outputs |
