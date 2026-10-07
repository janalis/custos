---
id: AliasFunctionsUsage
group: Language level migration
kind: semantic
needs: [names, stubs]
php: { min: "", max: "" }
---

# AliasFunctionsUsage

## Summary
Several built-in PHP functions exist only as alternative names for another
built-in (`sizeof` for `count`, `join` for `implode`, …), and a few of those
alternative names were deprecated or removed. Calling the canonical function
keeps code consistent and avoids relying on legacy names.

## Detection
Applies to plain function calls only (not method calls, static calls,
`new`, `use function` imports, or callable strings).

- **D1** The call's name (last segment, without any namespace qualifier),
  compared **case-insensitively** as PHP does for function names, is a key of
  one of the two tables below. `SIZEOF()` and `SizeOf()` are alias calls.
- **D2** The call resolves to a function whose fully-qualified name is in the
  global namespace (`\name`). Concretely:
  - a fully-qualified call `\join(...)` → global → candidate;
  - an unqualified call in the global namespace → global → candidate;
  - an unqualified call inside a namespace resolves to the global built-in
    only when there is no function of that name in the current namespace and
    no `use function` import bringing in a non-global function with that
    name. If the call resolves to a namespaced function
    (`Vendor\join`), it is **not** reported;
  - if the call cannot be resolved at all (function unknown to the stubs and
    not declared in the project), it is not reported. All names in the
    tables are treated as known built-ins.
- **D3** The node is not part of a `use function` statement.

### Table A — replaceable aliases (report + fix)
| alias | canonical |
|---|---|
| close | closedir |
| is_double | is_float |
| is_integer | is_int |
| is_long | is_int |
| is_real | is_float |
| sizeof | count |
| doubleval | floatval |
| fputs | fwrite |
| join | implode |
| key_exists | array_key_exists |
| chop | rtrim |
| ini_alter | ini_set |
| is_writeable | is_writable |
| pos | current |
| show_source | highlight_file |
| strchr | strstr |
| set_file_buffer | stream_set_write_buffer |
| session_commit | session_write_close |
| socket_getopt | socket_get_option |
| socket_setopt | socket_set_option |
| openssl_get_privatekey | openssl_pkey_get_private |
| posix_errno | posix_get_last_error |
| ldap_close | ldap_unbind |
| pcntl_errno | pcntl_get_last_error |
| ftp_quit | ftp_close |
| socket_set_blocking | stream_set_blocking |
| stream_register_wrapper | stream_wrapper_register |
| socket_set_timeout | stream_set_timeout |
| socket_get_status | stream_get_meta_data |
| diskfreespace | disk_free_space |
| odbc_do | odbc_exec |
| odbc_field_precision | odbc_field_len |
| recode | recode_string |
| mysqli_escape_string | mysqli_real_escape_string |
| mysqli_execute | mysqli_stmt_execute |

(`rand`/`srand` are deliberately absent: another rule handles them.)

### Table B — legacy aliases (report only, no fix)
| alias | status (for the message) |
|---|---|
| mysqli_bind_param | deprecated 5.3, removed 5.4 |
| mysqli_bind_result | deprecated 5.3, removed 5.4 |
| mysqli_client_encoding | deprecated 5.3, removed 5.4 |
| mysqli_fetch | deprecated 5.3, removed 5.4 |
| mysqli_param_count | deprecated 5.3, removed 5.4 |
| mysqli_get_metadata | deprecated 5.3, removed 5.4 |
| mysqli_send_long_data | deprecated 5.3, removed 5.4 |
| ocifreecursor | deprecated 5.4, discouraged |
| magic_quotes_runtime | deprecated 5.3, removed 7.0 |

The argument list is irrelevant (any number of arguments, including none).

## Exceptions (no report)
- **E1** The call resolves to a function declared in a namespace (local
  function with the same short name, or one imported via `use function`).
- **E2** (removed: names are matched case-insensitively, see D1.)
- **E3** Method calls (`$x->join()`), static calls (`A::sizeof()`), function
  declarations, and string callables (`'join'`).
- **E4** Names not in either table.

## Report
- Range: the function name identifier only — the bare name token, excluding
  any leading `\` or namespace qualifier and excluding the argument list.
  For `\join($a, $b)` only `join` is highlighted.
- Severity: warning (rendered as "deprecated" style; conformance only checks
  the severity class warning).
- Message:
  - Table A: `Use '{canonical}(...)' instead of the alias '{alias}(...)'.`
  - Table B: `'{alias}(...)' is a legacy alias ({status}); stop relying on it.`
  - `{alias}` is the name as written in the call (`'SizeOf(...)'`);
    `{canonical}` is the lower-case table value.

## Fix
- **F1** Table A only: replace the name identifier token with the canonical
  name. Any namespace qualifier (`\`) and the arguments are kept untouched:
  `\fputs($h, $s)` → `\fwrite($h, $s)`; `sizeof($list)` → `count($list)`.
  When the alias is written unqualified and an unqualified call to the
  canonical name at that position would not reach the global function (a
  `use function` import under that name, or a same-named function declared
  in the current namespace), the canonical name is written `\canonical`
  (`namespace App; function count($x) {…} sizeof($l)` → `\count($l)`); the
  message names that spelling too.
- **F2** Table B: no fix offered.

## Options
None.

## PHP versions
No gating. (Removed functions in Table B are still reported at any level.)

## Examples

```php
<?php
$total = <warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
$csv   = \<warning descr="Use 'implode(...)' instead of the alias 'join(...)'.">join</warning>(';', $cells);
$ok    = <warning descr="Use 'is_int(...)' instead of the alias 'is_long(...)'.">is_long</warning>($port) && <warning descr="Use 'is_writable(...)' instead of the alias 'is_writeable(...)'.">is_writeable</warning>($dir);
<warning descr="'magic_quotes_runtime(...)' is a legacy alias (deprecated 5.3, removed 7.0); stop relying on it.">magic_quotes_runtime</warning>(0);
$n = <warning descr="Use 'count(...)' instead of the alias 'SizeOf(...)'.">SizeOf</warning>($rows);
$s = $builder->join(',');  // E3
```

```php
<?php
$total = count($rows);
$csv   = \implode(';', $cells);
$ok    = is_int($port) && is_writable($dir);
magic_quotes_runtime(0);
$n = count($rows);
$s = $builder->join(',');  // E3
```

Namespace handling:

```php
<?php
namespace Shop\Util {
    function chop($s) { return $s; }

    $a = chop($label);   // E1: resolves to Shop\Util\chop
    $b = \<warning descr="Use 'rtrim(...)' instead of the alias 'chop(...)'.">chop</warning>($label);
}

namespace Shop\Web {
    use function Shop\Util\chop;

    $c = chop($label);   // E1: imported namespaced function
    $d = \<warning descr="Use 'rtrim(...)' instead of the alias 'chop(...)'.">chop</warning>($label);
    $e = <warning descr="Use 'current(...)' instead of the alias 'pos(...)'.">pos</warning>($stack); // falls back to global
}
```

## Divergences
- custos diverges: upstream matches the alias name case-sensitively, so
  `SIZEOF()` or `Join()` are missed. PHP function names are case-insensitive,
  so these are the same alias calls; custos matches case-insensitively (D1).
  The fix still writes the lower-case canonical name.
- Some Table A/B functions no longer exist in recent PHP (e.g. `recode`,
  `magic_quotes_runtime`); upstream still treats them as resolvable through
  its stubs. Implement with a built-in name list so resolution succeeds
  regardless of the project PHP level.
- **Builtin spelling (custos diverges).** Upstream inserts the bare
  canonical name, which a same-named function declared in or imported into
  the namespace captures (the fixed call then runs user code). custos
  writes `\canonical` in that case (F1).
