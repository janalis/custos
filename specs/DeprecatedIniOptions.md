---
id: DeprecatedIniOptions
group: Compatibility
kind: syntax
needs: []
php: { min: "", max: "" }
---

# DeprecatedIniOptions

## Summary

Reading, changing or restoring an ini directive that the targeted PHP version
has deprecated or removed is either a no-op or a future breakage. Point out
such directives, naming the version and the replacement when there is one.

## Detection

Visit every plain function call.

- **D1** The call resolves to one of the global functions `ini_set`,
  `ini_get`, `ini_alter`, `ini_restore`: names compare case-insensitively,
  and a call resolving to a same-named namespaced function (declared in the
  current namespace, imported, or written qualified) does not count.
- **D2** At least one argument, and the first argument is **directly** a
  string literal (single- or double-quoted; no variable/constant tracing).
- **D3** Let `name` be the literal's raw content (between the quotes),
  lower-cased (ASCII). Look it up in the table below (exact match after
  lower-casing).
- **D4** Let `L` be the configured language level.
  - If the entry has a *removed* version and `L >= removed` → report
    **removed** (pattern R).
  - Else if the entry has a *deprecated* version and `L >= deprecated` →
    report **deprecated** (pattern P).
  - Otherwise no report.

### Directive table (facts)

| Directive | Deprecated | Removed | Replacement |
|---|---|---|---|
| `define_syslog_variables` | 5.3 | 5.4 | – |
| `magic_quotes_gpc` | 5.3 | 5.4 | – |
| `magic_quotes_runtime` | 5.3 | 5.4 | – |
| `magic_quotes_sybase` | 5.3 | 5.4 | – |
| `highlight.bg` | – | 5.4 | – |
| `xsl.security_prefs` | 5.4 | 7.0 | `XsltProcessor->setSecurityPrefs()` |
| `safe_mode` | 5.3 | 5.4 | – |
| `safe_mode_gid` | 5.3 | 5.4 | – |
| `safe_mode_include_dir` | 5.3 | 5.4 | – |
| `safe_mode_exec_dir` | 5.3 | 5.4 | – |
| `safe_mode_allowed_env_vars` | 5.3 | 5.4 | – |
| `safe_mode_protected_env_vars` | 5.3 | 5.4 | – |
| `sql.safe_mode` | – | 7.2 | – |
| `asp_tags` | – | 7.0 | – |
| `always_populate_raw_post_data` | 5.6 | 7.0 | – |
| `y2k_compliance` | – | 5.4 | – |
| `zend.ze1_compatibility_mode` | – | 5.3 | – |
| `allow_call_time_pass_reference` | 5.3 | 5.4 | – |
| `register_globals` | 5.3 | 5.4 | – |
| `register_long_arrays` | 5.3 | 5.4 | – |
| `session.hash_function` | – | 7.1 | – |
| `session.hash_bits_per_character` | – | 7.1 | – |
| `session.entropy_file` | – | 7.1 | – |
| `session.entropy_length` | – | 7.1 | – |
| `session.bug_compat_42` | – | 5.4 | – |
| `session.bug_compat_warn` | – | 5.4 | – |
| `iconv.input_encoding` | 5.6 | – | `default_charset` |
| `iconv.output_encoding` | 5.6 | – | `default_charset` |
| `iconv.internal_encoding` | 5.6 | – | `default_charset` |
| `mbstring.script_encoding` | – | 5.4 | `zend.script_encoding` |
| `mbstring.func_overload` | 7.2 | 8.0 | – |
| `mbstring.internal_encoding` | 5.6 | – | `default_charset` |
| `mbstring.http_input` | 5.6 | – | `default_charset` |
| `mbstring.http_output` | 5.6 | – | `default_charset` |
| `track_errors` | 7.2 | 8.0 | – |
| `pdo_odbc.db2_instance_name` | 7.3 | 8.0 | – |
| `opcache.load_comments` | – | 7.0 | – |
| `opcache.fast_shutdown` | – | 7.2 | – |
| `opcache.inherited_hack` | – | 7.3 | – |
| `allow_url_include` | 7.4 | – | – |
| `assert.quiet_eval` | – | 8.0 | – |
| `log_errors_max_len` | – | 8.1 | – |
| `mysqlnd.fetch_data_copy` | – | 8.1 | – |
| `auto_detect_line_endings` | 8.1 | – | – |
| `date.default_latitude` | 8.1 | – | – |
| `date.default_longitude` | 8.1 | – | – |
| `date.sunrise_zenith` | 8.1 | – | – |
| `date.sunset_zenith` | 8.1 | – | – |
| `filter.default` | 8.1 | – | – |
| `filter.default_flags` | 8.1 | – | – |
| `assert.active` | 8.3 | – | `zend.assertions` |
| `assert.bail` | 8.3 | – | – |
| `assert.callback` | 8.3 | – | – |
| `assert.exception` | 8.3 | – | – |
| `assert.warning` | 8.3 | – | – |
| `opcache.consistency_checks` | – | 8.3 | – |
| `session.sid_length` | 8.4 | – | – |
| `session.sid_bits_per_character` | 8.4 | – | – |

The removal versions of `mbstring.func_overload`, `track_errors` and
`pdo_odbc.db2_instance_name`, and every row from `allow_url_include` down,
are custos additions taken from the PHP changelogs (see Divergences).
Directives deprecated only for some values (PHP 8.4's
`session.use_only_cookies = 0`, `session.use_trans_sid = 1`, …) are not
listed: reading or setting them is not deprecated in itself.

## Exceptions (no report)

- **E1** First argument not a string literal (variable, constant,
  concatenation).
- **E2** Directive not in the table.
- **E3** Language level below the entry's deprecated version (or below the
  removed version when the entry has no deprecated version).
- **E4** Other functions (`ini_get_all`, `get_cfg_var`), method calls named
  like the targets.

## Report

- Range: the first argument string literal, including its quotes.
- Severity: warning for both patterns (P is additionally rendered as
  deprecated/strikethrough).
- Messages (`{name}` is the lower-cased directive, `{ver}` the version
  formatted `X.Y.0`, `{alt}` the replacement):
  - R without replacement: `Ini directive '{name}' no longer exists since PHP {ver}.`
  - R with replacement: `Ini directive '{name}' no longer exists since PHP {ver}; use {alt}.`
  - P without replacement: `Ini directive '{name}' is deprecated since PHP {ver}.`
  - P with replacement: `Ini directive '{name}' is deprecated since PHP {ver}; use {alt}.`

## Fix

None.

## Options

None.

## PHP versions

Fully level-dependent (D4). Upstream's only fixture runs at the IDE test
default level, which (as that fixture proves: `always_populate_raw_post_data`
reported as *deprecated*, not removed) is at least 5.6 and below 7.0 —
custos conformance should run it at 5.6.

## Examples

Target PHP 5.6:

```php
<?php
$enc = ini_get(<warning descr="Ini directive 'mbstring.http_output' is deprecated since PHP 5.6.0; use default_charset.">'mbstring.http_output'</warning>);
ini_set(<warning descr="Ini directive 'safe_mode' no longer exists since PHP 5.4.0.">"SAFE_MODE"</warning>, '0');
ini_alter(<warning descr="Ini directive 'mbstring.script_encoding' no longer exists since PHP 5.4.0; use zend.script_encoding.">'mbstring.script_encoding'</warning>, 'UTF-8');
\ini_restore(<warning descr="Ini directive 'always_populate_raw_post_data' is deprecated since PHP 5.6.0.">"always_populate_raw_post_data"</warning>);
ini_set('asp_tags', '1');
ini_set('track_errors', '1');
ini_set('memory_limit', '64M');
$key = 'safe_mode';
ini_get($key);
```

Target PHP 7.2: the same `asp_tags` line reports *no longer exists since
PHP 7.0.0*, `track_errors` reports *deprecated since PHP 7.2.0*, and
`always_populate_raw_post_data` reports *no longer exists since PHP 7.0.0*.

## Divergences

- **Function matching (custos diverges from upstream).** Upstream matches
  the written last segment case-sensitively without resolution, so
  `INI_SET('safe_mode', ...)` is missed and a namespace's own `ini_set()` is
  checked. custos resolves the call to the global function, in any case.
- **Table extended past PHP 7.3 (custos diverges):** upstream's table stops
  at 7.3, so directives deprecated or removed in 7.4–8.4 are never reported
  and 7.2/7.3 deprecations removed in 8.0 are still called "deprecated" at
  8.x. custos adds those directives and removal versions (directive table).
  Upstream's fixture runs at 5.6, where none of the additions apply.
