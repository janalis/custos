---
id: EncryptionInitializationVectorRandomness
group: Security
kind: semantic
needs: [names, index, hierarchy]
php: { min: "", max: "" }
---

# EncryptionInitializationVectorRandomness

## Summary
An encryption initialization vector must be unpredictable. When the IV passed
to `openssl_encrypt()` / `mcrypt_encrypt()` can be traced to something other
than a cryptographic random generator (a literal, `mt_rand()`, …), report it.

## Detection
- **D1** A plain function call (not a method call) resolving to the global
  function `openssl_encrypt` (variant **O**) or `mcrypt_encrypt` (variant
  **M**); names compare case-insensitively and a same-named namespaced
  function (declared in the current namespace, imported or written
  qualified) does not count.
- **D2** The call has at least **5** arguments (PHP 7.1+ `openssl_encrypt`
  takes up to 8: tag, AAD and tag length follow the IV), and the fifth
  argument `A` (index 4, the IV) has non-empty source text.
- **D3** Run *value discovery* (as defined in the `CallableMethodValidity`
  spec) on `A`. If it yields nothing, stop.
- **D4** For each discovered value `v`: it is *secure* when `v` is a plain
  function call (not a method or static call) resolving to the global
  `random_bytes`, `openssl_random_pseudo_bytes` or `mcrypt_create_iv` (any
  case; a same-named namespaced function is not secure). In D5 the method
  and static call names are compared case-insensitively too. Every non-secure value contributes its source text to
  the list `R` (one entry per discovered node; identical texts from
  different nodes appear twice).
- **D5 (wrapper exception)** If `R` is empty, stop. Otherwise, if value
  discovery produced exactly **one** value, and that value is a call of any
  kind (function, method or static call) that resolves to a declaration
  with a body, and that body contains — anywhere — a call (function or
  method, any receiver) whose name is one of the three secure names, stop
  (treated as a wrapper around a secure generator).
- **D6** Otherwise report `A`.

## Exceptions (no report)
- **E1** Fewer than 5 arguments (the IV omitted).
- **E2** All discovered values are calls to one of the secure functions.
- **E3** IV produced by a single call to a resolvable user function/method
  whose body calls a secure generator (D5).
- **E4** Value discovery finds nothing (e.g. a variable at top level, an
  unresolvable constant).

## Report
- Range: the fifth argument `A` exactly as written (a variable, property
  fetch, class constant…), even when the offending values were found
  elsewhere.
- Severity: error.
- Message: `Generate the IV with {G}(); it may come from: {list}.` where
  `{G}` is `openssl_random_pseudo_bytes` for O and `mcrypt_create_iv` for M,
  and `{list}` is `R` sorted in ascending code-unit (ASCII) order joined with
  `, ` (so quoted literals such as `'abc'` sort before `mt_rand()`).

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
class Box
{
    const SALT = 'static-iv';
    private $nonce = 'fixed';
    private $fresh;

    public function lock($msg, $pass, $iv = '0000')
    {
        $iv = uniqid();
        $out = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: '0000', uniqid().">$iv</error>);
        $old = mcrypt_encrypt('rijndael-128', $pass, $msg, 'cbc', <error descr="Generate the IV with mcrypt_create_iv(); it may come from: 'static-iv'.">self::SALT</error>);

        $this->nonce = rand();
        $alt = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, <error descr="Generate the IV with openssl_random_pseudo_bytes(); it may come from: 'fixed', rand().">$this->nonce</error>);

        $this->fresh = random_bytes(16);
        $ok1 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, $this->fresh);
        $ok2 = openssl_encrypt($msg, 'aes-256-cbc', $pass, 0, $this->makeIv());
        $ok3 = openssl_encrypt($msg, 'aes-256-cbc', $pass);

        return [$out, $old, $alt, $ok1, $ok2, $ok3];
    }

    private function makeIv()
    {
        return openssl_random_pseudo_bytes(16);
    }
}
```

## Divergences
- **Function-name matching (custos diverges from upstream).** Upstream
  matches the encryption functions and the secure generators by their
  written last segment, case-sensitively and without resolution. custos
  resolves them: `\OpenSSL_Encrypt(..., \Random_Bytes(16))` is analysed and
  secure, a namespace's own `random_bytes()` is not a secure source, and a
  namespace's own `openssl_encrypt()` is not checked.
- In D4 a *method* call named like a secure function
  (`$rng->random_bytes(16)`) is not secure, but in D5 any call named like a
  secure function inside the wrapper body counts, regardless of receiver.
  Kept on purpose: inside a wrapper, a method named like a secure generator
  is almost always a CSPRNG facade, and requiring global calls there would
  mostly add false positives.
- **IV checked with 6–8 arguments (custos diverges):** upstream only looks
  at calls with exactly 5 arguments, so an AEAD call such as
  `openssl_encrypt($m, 'aes-256-gcm', $k, 0, 'fixed-nonce', $tag)` — where a
  repeated nonce is catastrophic — goes unreported. custos checks the fifth
  argument whenever there are at least 5 (D2).
- **Only reaching assignments (custos diverges).** Upstream discovers every
  assignment of an IV variable in the function, so
  `$iv = random_bytes(16); openssl_encrypt(…, $iv); $iv = base64_encode($iv);`
  was reported for `base64_encode($iv)`, which runs after the call, and a
  parameter default overwritten before the call was listed too. For a local
  variable argument inside a function custos takes only the plain `=`
  assignments that reach the call, plus the parameter default when the
  entry value reaches it (`util.PossibleValuesReaching`); other arguments
  keep D3.
- **Wrappers in other files (custos diverges).** D5 also accepts a
  wrapper declared in another file of the project: the index records
  whether a function or method body calls a secure generator by name, so
  `openssl_encrypt(…, make_iv(16))` with `make_iv()` returning
  `random_bytes()` elsewhere is not reported.
