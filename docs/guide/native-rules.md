# Native inspections

The catalogue includes 100 inspections designed for custos. Each has a
[rule reference page](/rules/) with detection conditions, examples, PHP version
requirements and available fixes. Native rules use their stable custos ID for
configuration and suppression.

## Choosing rules

66 native inspections are enabled by default. The remaining 34 depend on the
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

Fixes are available for 19 native inspections. A fix requires the rule's
specific prerequisites and preserves source comments and argument evaluation
order. Some repair a proven bug, such as adding an omitted callback return or
preserving a large JSON integer as a string. Rules requiring a choice of
error-handling or application policy report a finding without an automatic fix.

Use the normal [fix command](./cli) to review and apply available edits.
