# Native inspections

The catalogue includes 300 inspections designed for custos. Each has a
[rule reference page](/rules/) with detection conditions, examples, PHP version
requirements and available fixes. Native rules use their stable custos ID for
configuration and suppression.

## Choosing rules

201 native inspections are enabled by default. The remaining 99 depend on the
intended meaning of values, keys, dates or text, so enable them when their
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

An earlier 100-rule expansion covers reference aliases, generators and fibers,
array shapes, serialization, encryption, sessions and cookies, cURL,
database bindings and transactions, process pipes, encodings, ZIP, XML,
images and sodium. They report concrete misuse such as resuming an unstarted
fiber, mismatched binding counts, an image MIME type that conflicts with its
encoder, or accepting a ZIP error code as successful opening.

A further 100 inspections add language and reflection contracts, SPL iterator
behavior, stream modes and CSV records, HTTP response framing, BCMath and GMP,
binary formats, DOM and XML ownership, SQLite and PostgreSQL results,
internationalization, compression, archive budgets and process status.
Examples include constructing an abstract class, moving a DOM node between
documents without importing it, consuming a SQLite result after finalization,
and declaring a response length that differs from the proven emitted bytes.

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

Three security inspections accept application documentation to establish intent.
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

Fixes are available for 41 native inspections. A fix requires the rule's
specific prerequisites and preserves source comments and argument evaluation
order. Some repair a proven bug, such as adding an omitted callback return or
preserving a large JSON integer as a string. Rules requiring a choice of
error-handling or application policy report a finding without an automatic fix.

The new fixes replace NaN equality tests with `is_nan()` when numeric types
are proven, require ZIP opening to return exactly `true`, and require OpenSSL
signature verification to return exactly `1`. Additional narrowly gated fixes
remove pattern escaping from literal replacement text, normalize parsed query
keys, and correct a directly constructed sodium nonce length. Missing error
handling, encryption
key derivation and authentication policy require a project decision and have
no automatic correction.

Use the normal [fix command](./cli) to review and apply available edits.
