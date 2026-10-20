# Native rules

Native rules are designed for custos, independently of Php Inspections (EA
Extended). They cover language behavior and extension APIs: references,
generators and fibers, JSON and regular expressions, encryption and sessions,
databases and streams, encodings, DOM, ZIP, images and more. The
[rule reference](/rules/) marks them as *native*, in its index and with a
badge on each rule page; each page describes what the rule needs to see
before it reports, and whether a quick-fix exists.

Native rules have no PhpStorm name: configure and suppress them by rule ID
(`@custos-ignore <ID>`).

## Rules off by default

Some native rules depend on what your code means, such as whether an array is
a list or a record map, and are off by default. Try one in isolation first:
`--rule` runs it even while it is off.

```sh
custos explain ArrayMergeNumericKeyLoss
custos analyse --rule ArrayMergeNumericKeyLoss src/
```

Then enable the ones whose contract fits your project:

```json
{ "rules": { "ArrayMergeNumericKeyLoss": { "enabled": true } } }
```

## Policy options

Several native rules only report against a policy you state explicitly, such
as the names of your authentication cookies or of your sensitive files. Names
in the code alone never establish that policy. Each rule page lists its
options and defaults; for example:

```json
{
  "rules": {
    "AuthenticationCookieAllowsScriptAccess": {
      "enabled": true,
      "options": { "authenticationCookies": ["auth_token", "remember_me"] }
    }
  }
}
```

## custos annotations

A few security rules read intent from docblock annotations:

- `@custos-protected` on a project function marks it as protected work, for
  [RedirectContinuesProtectedExecution](/rules/probable-bugs/RedirectContinuesProtectedExecution)
  and [UntrustedForwardedClientAddress](/rules/probable-bugs/UntrustedForwardedClientAddress).
- [FastDigestUsedForPasswordStorage](/rules/security/FastDigestUsedForPasswordStorage)
  needs **both** of these: `@custos-credential-store` on the project function
  that stores credentials, and `@param password-string $name` on the
  parameter that holds the password. It reports a fast digest of that
  password passed to that function.

```php
/** @custos-credential-store */
function saveCredential($hash) {}

/** @param password-string $password */
function register($password) {
    saveCredential(md5($password)); // reported
}
```
