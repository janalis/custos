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
`define()` value; `true`/`false`/`null` stay as themselves) — with two
custos refinements (see Divergences):

- **Local variables are flow-sensitive.** A variable contributes only the
  plain assignments that can reach its use (and its parameter default only
  when the value it had on entry can still reach it). An assignment
  statement executed on every path to the use hides everything written
  before it; an assignment in a branch or a loop is one more alternative.
- **Overwritten properties.** The discovery result is unknown when the last
  word on the property is an unclassifiable value — one in which nothing
  can be classified as "enable" or "disable": a parameter without default,
  a call result, an unstable variable. Two places are read, each in source
  order: the current function up to the place the property is read, and
  the constructor of the class declaring it (there through `$this->name`
  or a static reference to the name). An unclassifiable assignment counts
  only when it is a statement that always runs: in the current function, a
  statement of a block enclosing the read; in the constructor, a statement
  of its body. A classifiable assignment written after it makes the
  property decidable again, an assignment written after the read is
  ignored, and an assignment in a branch, or one where the unknown is only
  one alternative next to a classifiable value
  (`$opts['verify'] ?? false`, `$dev ? 0 : $cfg`), does not make the
  result unknown.

Each discovered value is classified:

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
  scope, or a property is overwritten with an unclassifiable value; custos
  refinements, see Divergences) → no report.

## Exceptions (no report)

- **E1** `curl_setopt` with other than 3 arguments; calls to other
  functions, including a namespaced lookalike `curl_setopt` that the call
  resolves to.
- **E2** `K` used as an array element *value* (`[CURLOPT_SSL_VERIFYPEER]`),
  in a comparison, as an argument of other functions, or in an array access
  not under a plain assignment (`$o[CURLOPT_SSL_VERIFYPEER] ??= 0`,
  `isset($o[CURLOPT_SSL_VERIFYHOST])`), or as a key in a destructuring
  pattern (`[CURLOPT_SSL_VERIFYPEER => $peer] = $opts;`, `list(...)`,
  `foreach ($all as [CURLOPT_SSL_VERIFYHOST => $h])`, at any nesting
  depth): the pattern reads the option, it does not set it.
- **E3** Mixed discoveries such as `$strict ? 2 : 0` (H) or
  `$dev ? 0 : true` (P) contain an "enable" → not reported.
- **E4** Values that cannot be classified (`getenv('VERIFY')`, `10`).
- **E5** Other `CURLOPT_*` constants.
- **E7** (custos, see Divergences) D2 in an array literal whose every
  element value is a string literal with non-numeric, non-blank content
  (`[CURLOPT_PROXY => 'proxy', CURLOPT_SSL_VERIFYPEER => 'verify_peer']`):
  such an array maps options to names, it does not set them.
- **E6** A disabling value that cannot be the one set: a local variable's
  earlier value replaced on every path before the use
  (`$verify = false; $verify = resolve($config);`), or a property default
  replaced by an unclassifiable value (`private $verifyPeer = false;` with
  a constructor doing `$this->verifyPeer = $verifyPeer;`). A property
  disabled, used and then restored (`$this->peer = false;`, the
  `curl_setopt` call, `$this->peer = $previous;`) is still reported.

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

- **E7 — custos diverges from upstream.** Upstream reports an array
  that maps cURL options to names as if it set them to strings.
- **Destructuring patterns (custos diverges).** A `key => $var` element of
  an array used as a destructuring target is not a setting; custos skips
  it (E2) instead of discovering the target variable's earlier values
  (`$peer = false; [CURLOPT_SSL_VERIFYPEER => $peer] = $opts;` was reported).
- **`curl_setopt` matching (custos diverges from upstream).** Upstream
  accepts any call whose written last segment is exactly `curl_setopt`, so a
  namespace's own `curl_setopt()` is checked while `\Curl_SetOpt()` is not.
  custos resolves the call to the global function, in any case (D1).
- **Unstable variables — custos refinement, not upstream.** Value discovery ignores `++`/`--` and compound assignments upstream, so a variable later incremented or extended is analyzed with its initial value only. custos makes the result unknown (no report), as in the shared value discovery of `CallableMethodValidity`. No upstream fixture relies on such a variable.
- **Dead initial values (custos diverges).** Upstream collects every
  assignment of a variable or property regardless of order, so a disabling
  initial value is reported even when it is always replaced by something
  unknown before the option is set. custos only takes the assignments
  reaching a local variable's use and stays silent for a property
  overwritten with an unclassifiable value (Detection, E6). A disabling
  value that remains one possible alternative is still reported.
- D1 accepts `K` in any argument position of `curl_setopt`
  (`curl_setopt($h, $opt, CURLOPT_SSL_VERIFYPEER)` would evaluate the
  constant itself as the value; that resolves to a multi-digit number and is
  ignored, so in practice nothing is reported). custos does the same: `K` is
  not required to be the second argument.
- For H, `true` counts as disabling (cURL treats `true` as `1`, which is not
  the safe value `2`). Keep.
