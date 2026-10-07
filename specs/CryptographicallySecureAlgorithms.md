---
id: CryptographicallySecureAlgorithms
group: Security
kind: syntax
needs: []
php: { min: "", max: "" }
---

# CryptographicallySecureAlgorithms

## Summary
Some cipher/hash selector constants of mcrypt, OpenSSL and `crypt()` pick
algorithms that are broken or weak (DES, 3DES, RC2, RC4, MD5) or that are
commonly mistaken for AES (Rijndael with 192/256-bit blocks). Point at each
use so a stronger algorithm can be chosen.

## Detection
- **D1** A constant reference (not a class constant `X::C`, not a string)
  that resolves to a **global** constant whose name, compared
  case-sensitively, is one of the names in the table below. It resolves to
  the global constant when written fully qualified (`\MCRYPT_DES`), or
  unqualified where PHP falls back to the global constant: outside any
  namespace, or inside one that neither imports the name (`use const`) nor
  declares a constant of that name. Qualified references
  (`Crypto\MCRYPT_DES`, `namespace\MCRYPT_DES`) and imported or
  namespace-declared constants of the same name are not reported.
- **D2** Not a test context: the file path does not end with `Test.php`,
  `Spec.php` or `.phpt` and does not contain `/Fixtures/`, and the enclosing
  class (if any) has an FQN that neither ends with `Test` nor contains
  `\Tests\` or `\Test\`.

| Constant | Problem family | Suggested alternative |
|---|---|---|
| `MCRYPT_RIJNDAEL_192` | Rijndael with a 192-bit block is not AES | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_RIJNDAEL_256` | Rijndael with a 256-bit block is not AES | `MCRYPT_RIJNDAEL_128` with a 256-bit key |
| `MCRYPT_3DES` | 3DES | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_TRIPLEDES` | 3DES | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_DES_COMPAT` | DES | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_DES` | DES | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_RC2` | RC2 | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_RC4` | RC4 | `MCRYPT_RIJNDAEL_128` |
| `MCRYPT_ARCFOUR` | RC4 | `MCRYPT_RIJNDAEL_128` |
| `OPENSSL_CIPHER_3DES` | 3DES | an `AES-128-*` cipher |
| `OPENSSL_CIPHER_DES` | DES | an `AES-128-*` cipher |
| `OPENSSL_CIPHER_RC2_40` | RC2 | an `AES-128-*` cipher |
| `OPENSSL_CIPHER_RC2_64` | RC2 | an `AES-128-*` cipher |
| `CRYPT_MD5` | MD5 | `CRYPT_BLOWFISH` |
| `CRYPT_STD_DES` | DES | `CRYPT_BLOWFISH` |

Context does not matter: any read of such a constant (argument, echo,
comparison, array element, default value…) is reported. The constant does
not need to be resolvable.

## Exceptions (no report)
- **E1** Test contexts (D2).
- **E2** Other constants, including `OPENSSL_ALGO_MD5`/`OPENSSL_ALGO_SHA1`
  and friends, `MCRYPT_RIJNDAEL_128`, `CRYPT_BLOWFISH`.
- **E3** Different spelling case (`mcrypt_des`), class constants
  (`Cipher::MCRYPT_DES`), algorithm names given as strings (`'des-ede3'`),
  calls such as `md5()`.
- **E4** References that do not resolve to the global constant (D1).

## Report
- Range: the constant reference node as written, including a leading `\`
  or namespace qualifier if present (`\MCRYPT_DES` → the whole
  `\MCRYPT_DES`).
- Severity: error.
- Message: `Weak algorithm selected via {CONST} ({family}); prefer {alternative}.`
  (values from the table).

## Fix
None.

## Options
None.

## PHP versions
No gating (mcrypt constants are reported even on PHP versions where the
extension no longer exists).

## Examples

```php
<?php
namespace Vault;

final class Sealer
{
    public function seal($key, $data)
    {
        $legacy = mcrypt_encrypt(<error descr="Weak algorithm selected via MCRYPT_TRIPLEDES (3DES); prefer MCRYPT_RIJNDAEL_128.">MCRYPT_TRIPLEDES</error>, $key, $data, 'cbc');
        $wide   = [<error descr="Weak algorithm selected via MCRYPT_RIJNDAEL_256 (Rijndael with a 256-bit block is not AES); prefer MCRYPT_RIJNDAEL_128 with a 256-bit key.">\MCRYPT_RIJNDAEL_256</error>];
        if (<error descr="Weak algorithm selected via CRYPT_MD5 (MD5); prefer CRYPT_BLOWFISH.">CRYPT_MD5</error> === 1) {
            return openssl_encrypt($data, 'aes-128-gcm', $key);
        }
        return [$legacy, $wide, MCRYPT_RIJNDAEL_128, CRYPT_BLOWFISH, OPENSSL_ALGO_SHA1];
    }
}

class CipherMatrixTest
{
    public function provider()
    {
        return [MCRYPT_DES, OPENSSL_CIPHER_RC2_40];
    }
}
```

## Divergences
- **Name resolution (custos diverges):** upstream matches only the last
  segment of the constant name, so `Other\MCRYPT_DES` (or an unqualified
  reference to a namespace's own `MCRYPT_DES`) is reported although it is a
  different, user-defined constant that selects nothing weak. custos
  reports only references resolving to the global constant (D1).
