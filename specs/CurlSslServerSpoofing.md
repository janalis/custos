---
id: CurlSslServerSpoofing
group: Security
kind: semantic
needs: [names, index]
php: { min: "", max: "" }
---

# CurlSslServerSpoofing

## Summary
Turning off cURL's TLS peer or host-name verification
(`CURLOPT_SSL_VERIFYPEER`, `CURLOPT_SSL_VERIFYHOST`) lets anyone in the
middle impersonate the server. Report places where these options are set to a
disabling value.

## Detection
Start from every global constant reference whose name (last segment,
case-sensitive) is exactly `CURLOPT_SSL_VERIFYHOST` (kind **H**) or
`CURLOPT_SSL_VERIFYPEER` (kind **P**). Let `K` be that reference. Determine
the *setting site* from where `K` sits; exactly one of:

- **D1 (curl_setopt call)** `K` is a direct argument of a plain function call
  resolving to the global function `curl_setopt` (name compared
  case-insensitively; a same-named namespaced function does not count), and
  that call has exactly **3** arguments. The value `V` is
  the third argument. The reported node is the whole call.
  (Upstream does not check that `K` is the second argument — see
  Divergences.)
- **D2 (array element)** `K` is the key of a `key => value` array element
  (any array literal, `[...]` or `array(...)`, in any context — e.g. a
  `curl_setopt_array()` options array). `V` is the element's value. The
  reported node is the whole `key => value` element.
- **D3 (array write)** `K` is the index of an array access `X[K]`; walk up
  through any enclosing array accesses (`X[K][...]`); the node reached must
  be a plain assignment `… = V` (operator `=` only). The reported node is the
  whole assignment expression (without the `;`).

Then decide from `V` using *value discovery* (as defined in the
`CallableMethodValidity` spec) — it expands ternaries (both branches, the
condition is ignored), `??` (both sides), variables (parameter defaults and
plain assignments within the enclosing function; nothing at top level),
properties, class constants and global constants (resolved to their
`define()` value; `true`/`false`/`null` stay as themselves). Each discovered
value is classified:

- **D4 (kind H)** — "enable" when the value is a string literal whose raw
  content is exactly `2`, or any other expression whose source text is
  exactly the single character `2`; "disable" when it is a string literal
  with any other content, **any** constant reference (`true`, `false`,
  `null`, …), or another one-character expression (`0`, `1`, …).
- **D5 (kind P)** — "enable" when the value is a string literal whose raw
  content is exactly `1`, the constant `true` (case-insensitive), or any
  other one-character expression whose text is `1`; "disable" when it is a
  string literal with any other content, any other constant reference
  (`false`, `null`, …), or another one-character expression (`0`, `2`, …).
- Values that fit none of these (function calls, multi-character numbers
  such as `10` or `2.0`, unresolved expressions) are ignored.
- **D6** Report when at least one value is "disable" and none is "enable".
  Nothing discovered → no report. Unknown discovery result (a variable on
  the path is also incremented/decremented or compound-assigned in its
  scope; custos refinement, see Divergences) → no report.

## Exceptions (no report)
- **E1** `curl_setopt` with other than 3 arguments; calls to other
  functions, including a namespaced lookalike `curl_setopt` that the call
  resolves to.
- **E2** `K` used as an array element *value* (`[CURLOPT_SSL_VERIFYPEER]`),
  in a comparison, as an argument of other functions, or in an array access
  not under a plain assignment (`$o[CURLOPT_SSL_VERIFYPEER] ??= 0`,
  `isset($o[CURLOPT_SSL_VERIFYHOST])`).
- **E3** Mixed discoveries such as `$strict ? 2 : 0` (H) or
  `$dev ? 0 : true` (P) contain an "enable" → not reported.
- **E4** Values that cannot be classified (`getenv('VERIFY')`, `10`).
- **E5** Other `CURLOPT_*` constants.

## Report
- Range: D1 the whole `curl_setopt(...)` call (name to closing `)`); D2 the
  whole array element from key start to value end; D3 the whole assignment
  from target start to value end.
- Severity: error.
- Message:
  - H: `Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.`
  - P: `Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.`

## Fix
None.

## Options
None.

## PHP versions
No gating.

## Examples

```php
<?php
function fetch($handle, $insecure)
{
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, 1)</error>;
    <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, FALSE)</error>;
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, $insecure ? 0 : '1')</error>;
    curl_setopt($handle, CURLOPT_SSL_VERIFYHOST, $insecure ? 0 : 2);
    curl_setopt($handle, CURLOPT_SSL_VERIFYPEER, '1');

    $peer = null;
    curl_setopt_array($handle, [
        CURLOPT_TIMEOUT => 5,
        <error descr="Peer certificate verification is disabled; keep CURLOPT_SSL_VERIFYPEER enabled.">CURLOPT_SSL_VERIFYPEER => $peer</error>,
        CURLOPT_SSL_VERIFYHOST => "2",
    ]);

    $conf = [];
    <error descr="Host name verification is disabled; keep CURLOPT_SSL_VERIFYHOST at 2.">$conf['tls'][CURLOPT_SSL_VERIFYHOST] = true</error>;
    $conf[CURLOPT_SSL_VERIFYPEER] = true;
    return $conf;
}
```

## Divergences
- **`curl_setopt` matching (custos diverges from upstream).** Upstream
  accepts any call whose written last segment is exactly `curl_setopt`, so a
  namespace's own `curl_setopt()` is checked while `\Curl_SetOpt()` is not.
  custos resolves the call to the global function, in any case (D1).
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analysed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable; recorded in `docs/decisions.md` ("Spec-level false positives").
- D1 accepts `K` in any argument position of `curl_setopt`
  (`curl_setopt($h, $opt, CURLOPT_SSL_VERIFYPEER)` would evaluate the
  constant itself as the value; that resolves to a multi-digit number and is
  ignored, so in practice nothing is reported). Recommendation: require `K`
  to be the second argument — equivalent on all realistic code.
- For H, `true` counts as disabling (cURL treats `true` as `1`, which is not
  the safe value `2`). Keep.
