# Native inspections

Native inspections are designed for custos. Each has a
[rule reference page](/rules/) with detection conditions, examples, PHP version
requirements and available fixes. Native rules use their stable custos ID for
configuration and suppression.

## Choosing rules

Some native inspections depend on the intended meaning of values, keys, dates
or text and are disabled by default. Enable them when their
contracts fit your project. For example, preserving numeric keys matters for
record maps, while reindexing may be appropriate for lists.

```json
{
  "rules": {
    "ArrayMergeNumericKeyLoss": { "enabled": true },
    "FilteredListJsonShape": { "enabled": true }
  }
}
```

Use `custos explain FilteredListJsonShape` to inspect its default and fix
contract. To inspect one rule in isolation:

```sh
custos analyse --rule FilteredListJsonShape src/
```

## Additional PHP and extension contracts

Native inspections check language behavior and extension APIs, including:

- **Language and values:** reference aliases, generators and fibers, array
  shapes, reflection, numeric bounds, regex captures and JSON values.
- **Security and HTTP:** serialization, encryption, sodium state and nonces,
  sessions and cookies, cURL callbacks, response framing and TLS handshakes.
- **Data and resources:** database bindings, transactions and result lifetimes,
  SPL iterators, filesystem failure results, streams, CSV records, process
  pipes and shared memory.
- **Text, documents and images:** encodings, internationalization, BCMath and
  GMP, binary formats, ZIP and compression, DOM and XML, GD and Imagick.

Examples include resuming an unstarted fiber, mismatched database binding
counts, consuming a SQLite result after finalization, moving a DOM node between
documents without importing it, or accepting a ZIP error code as successful
opening. Each rule page describes the evidence required for a finding and
whether a fix is available.

Intent-sensitive findings are off by default. Examples include intentional
shared references, GET bodies, accumulated command output, permission policy,
and signature canonicalization. Local-flow findings require the operations
and downstream use described by their rule page; unknown data and escaped
handles cannot establish those facts.

Several optional rules accept explicit project policy. Configure the option
values explicitly; the signature-function list is empty by default and needs
your project’s actual function names:

```json
{
  "rules": {
    "AuthenticationCookieAllowsScriptAccess": {
      "enabled": true,
      "options": { "authenticationCookies": ["auth_token"] }
    },
    "WorldWritablePermission": {
      "enabled": true,
      "options": { "sensitivePaths": ["**/*.pem", "**/.env", "**/config.php"] }
    },
    "FormEncodingUsedForRfc3986Signature": {
      "enabled": true,
      "options": { "signatureFunctions": ["signRfc3986"] }
    }
  }
}
```

These lists establish specific policy: a cookie name identifies a configured
authentication cookie, a path glob identifies a sensitive file, and an exact
function name identifies a signature requiring RFC3986 query encoding.
Ordinary variable or function names do not establish these contracts.

`ImagickSingleBlobDropsRequiredFrames` requires an explicit frame-preservation
policy as well as proof that an image contains multiple frames. Enable both
the rule and its option when exporting every frame is part of your contract:

```json
{
  "rules": {
    "ImagickSingleBlobDropsRequiredFrames": {
      "enabled": true,
      "options": { "preserveFrames": true }
    }
  }
}
```

The option defaults to `false`. A filename or variable name does not establish
frame count or preservation policy.

## Flow and project calls

Flow inspections track reachable assignments, branch guards, object and
resource identities, and context-specific validation. They can distinguish
an unchecked read result from one rejected by a strict failure check.
Project analysis also builds summaries for resolved project functions and
methods, allowing supported wrapper calls to carry return values and unsafe
uses across files. The editor updates those summaries for open documents.

Ambiguous declarations, escaped resources and unsupported operations discard
definite facts. Recursive calls and bounded analysis exhaustion likewise
cannot establish a definite finding. This keeps an unknown result from being
treated as proof of a bug.

## Explicit application contracts

Security inspections can use application documentation to establish intent.
`RedirectContinuesProtectedExecution` and `UntrustedForwardedClientAddress`
recognize a resolved callable annotated with `@custos-protected` as protected
work. The forwarded-address rule requires an address comparison that controls
that protected work; logging an address does not establish an access decision.
`FastDigestUsedForPasswordStorage`
recognizes a password parameter and credential persistence through these exact
annotations:

```php
<?php
/** @custos-credential-store */
function saveCredential($hash) {}

/** @param password-string $password */
function registerPassword($password) {
    saveCredential(md5($password));
}
```

Names alone do not establish these contracts. See each rule page for its
precise scope.

## Fixes

A fix requires the rule's specific prerequisites and preserves source comments
and argument evaluation order. Some repair a proven bug, such as adding an
omitted callback return or preserving a large JSON integer as a string. Rules
requiring a choice of error-handling or application policy report a finding without an automatic fix.

Available fixes replace NaN equality tests with `is_nan()` when numeric types
are proven, require ZIP opening to return exactly `true`, and require OpenSSL
signature verification to return exactly `1`. Additional narrowly gated fixes
remove pattern escaping from literal replacement text, normalize parsed query
keys, and correct a directly constructed sodium nonce length. Missing error
handling, encryption key derivation and authentication policy require a project
decision and have no automatic correction.

Use the normal [fix command](./cli) to review and apply available edits.
