---
id: ForgottenDebugOutput
group: Probable bugs
kind: semantic
needs: [names, types, hierarchy, index]
php: { min: "", max: "" }
---

# ForgottenDebugOutput

## Summary

Dumping, tracing and profiler helpers (`var_dump()`, `print_r()`,
framework dumpers, Xdebug functions, …) are typically added while debugging
and forgotten afterwards. Flag their calls unless the context shows the
output is intentional.

## Configuration model

The rule works from a list of *debug entries* (strings). Each entry, after
trimming surrounding whitespace, is either

- a **function entry** (no `::`): a bare function name such as `var_dump`, or
- a **method entry** `ClassFQN::method` (split at the first `::`; the class
  part is the FQN with leading `\`, e.g. `\Acme\Debug::dump`).

Effective list:

- when option `migratedIntoUserSpace` is `false` (default), the effective list
  is the built-in defaults (below) **plus** every entry of option
  `configuration`;
- when `true`, the effective list is exactly `configuration`.

Built-in defaults:

- Method entries: `\Codeception\Util\Debug::pause`,
  `\Codeception\Util\Debug::debug`, `\Doctrine\Common\Util\Debug::dump`,
  `\Doctrine\Common\Util\Debug::export`,
  `\Symfony\Component\Debug\Debug::enable`,
  `\Symfony\Component\Debug\ErrorHandler::register`,
  `\Symfony\Component\Debug\ExceptionHandler::register`,
  `\Symfony\Component\Debug\DebugClassLoader::enable`,
  `\Zend\Debug\Debug::dump`, `\Zend\Di\Display\Console::export`,
  `\TYPO3\CMS\Core\Utility\DebugUtility::debug`,
  `\Illuminate\Support\Debug\Dumper::dump`.
- Function entries: `dd`, `dump`, `trap`, `debug_print_backtrace`,
  `debug_zval_dump`, `error_log`, `phpinfo`, `print_r`, `var_export`,
  `var_dump`, `dpm`, `dsm`, `dvm`, `kpr`, `dpq` (upstream also lists `wp_die`, see Divergences), and the Xdebug
  functions `xdebug_break`, `xdebug_call_class`, `xdebug_call_file`,
  `xdebug_call_function`, `xdebug_call_line`, `xdebug_code_coverage_started`,
  `xdebug_debug_zval`, `xdebug_debug_zval_stdout`, `xdebug_dump_superglobals`,
  `xdebug_enable`, `xdebug_get_code_coverage`, `xdebug_get_collected_errors`,
  `xdebug_get_declared_vars`, `xdebug_get_function_stack`,
  `xdebug_get_headers`, `xdebug_get_monitored_functions`,
  `xdebug_get_profiler_filename`, `xdebug_get_stack_depth`,
  `xdebug_get_tracefile_name`, `xdebug_is_enabled`, `xdebug_memory_usage`,
  `xdebug_peak_memory_usage`, `xdebug_print_function_stack`,
  `xdebug_start_code_coverage`, `xdebug_start_error_collection`,
  `xdebug_start_function_monitor`, `xdebug_start_trace`,
  `xdebug_stop_code_coverage`, `xdebug_stop_error_collection`,
  `xdebug_stop_function_monitor`, `xdebug_stop_trace`, `xdebug_time_index`,
  `xdebug_var_dump`.

Conformance harness: upstream cases supply extra entries through raw calls of
the form `ForgottenDebugOutput.registerCustomDebugMethod("<entry>")`; each
such call appends `<entry>` (with Java string escapes decoded, i.e. `\\` →
`\`) to `configuration`. An empty entry is harmless (it matches nothing).

## Detection

### Function calls

- **D1** A (non-method) function call whose name — last segment as written,
  compared case-insensitively as PHP compares function names (custos
  diverges), no namespace resolution — equals a function entry. So
  `\var_dump()` and `Some\Ns\var_dump()` qualify; a function entry written
  with a leading `\` never matches anything.
- **D2** The call is not the name inside a `use function …;` import.
- **D3** Argument-count allowance: for these functions a call with exactly the
  given number of arguments is considered intentional and skipped (names
  compared case-insensitively):
  `phpinfo` — 1, `print_r` — 2, `var_export` — 2. (No allowance for
  `var_dump`, `debug_zval_dump`, `debug_print_backtrace` or any other name:
  they are reported with any number of arguments.)
- **D4** Not in an *output-buffered* context (E2) and not inside a *debug
  wrapper* (E3) → report.

### Method calls

- **D5** A method call (`->`, `?->` or `::`) whose name as written
  equals the method part of some method entry, compared case-insensitively
  as PHP compares method names (custos diverges).
- **D6** For the method entries with that method name (in any order), the
  call resolves to a method whose **declaring** class FQN equals the entry's
  class part (compared case-insensitively, as PHP compares class names;
  leading `\` required in the entry).
  Calls through a subclass that inherit the method qualify; a subclass
  override does not.
- **D7** Not inside a debug wrapper (E3) → report once (stop after the first
  matching entry). The output-buffer exception does not apply to method
  calls.

## Exceptions (no report)

- **E1** Allowed argument counts (D3): `print_r($x, true)`,
  `var_export($x, true)`, `phpinfo(INFO_MODULES)`.
- **E2** Output buffering: the call is the expression of an expression
  statement (optionally behind a single `@` error-suppression operator, i.e.
  `@print_r($x);`), and the statement immediately preceding it in the same
  statement list is an expression statement whose expression is a plain
  function call named (as written, case-insensitive) `ob_start`, with any
  arguments: `ob_start(); var_dump($x);`. Only the very next statement is
  covered; calls nested in larger expressions (`echo print_r($x);`,
  `$s = print_r($x);`) are never covered.
- **E3** Debug wrapper: the nearest enclosing function-like is
  - a *named function* whose short name equals (case-insensitively) a function entry of the
    effective list — e.g. `function dd($v) { var_dump($v); }` is not reported
    when `dd` is in the list; or
  - a *method* whose short name equals (case-insensitively) a function
    entry (custos, see Divergences) — e.g. `public function dump()` calling
    `var_dump()`; or
  - a *method* whose identity `\<declaring class FQN>::<method name>` (compared
    case-insensitively) equals a method entry — e.g. calls inside
    `\Acme\Tracer::dump()` itself are not reported when
    `\Acme\Tracer::dump` is in the list. Methods of anonymous classes are
    never wrappers.
  Closures and arrow functions are never wrappers.
- **E4** `use function var_dump;` import lines.

## Report

- Range: the whole call expression — for functions from the name (including
  any leading `\`/qualifier) to the closing `)`; for methods from the start of
  the receiver/class name to the closing `)` (e.g. `Tracer::dump($x)`,
  `$this->logger->dump($x)`). No trailing `;`.
- Severity: error.
- Message: `Debug output call; remove it if it was left over from debugging.`

## Fix

None.

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| configuration | list of strings | empty (see below) | Additional debug entries (`function_name` or `\Class::method`). |
| migratedIntoUserSpace | bool | false | When false, the built-in defaults are merged into `configuration`; when true, only `configuration` is used. |

Upstream persists the merged list into `configuration` once and flips
`migratedIntoUserSpace` to true; custos only needs the "effective list"
semantics above.

## PHP versions

None.

## Examples

Effective list = defaults + `audit_dump` + `\Acme\Tracer::dump` +
`\Acme\Probe::dump` (`migratedIntoUserSpace = false`).

```php
<?php
namespace Acme;

use function var_dump;

class Tracer { public static function dump($v) {} }
class Probe  { public function dump($v) {} }
class Quiet  { public function dump($v) {} }

function audit_dump($v) {
    print_r($v);
}

function work($order, Probe $probe, Quiet $quiet) {
    <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($order, $probe)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">\print_r($order)</error>;
    $text = print_r($order, true);
    echo var_export($order, true);
    <error descr="Debug output call; remove it if it was left over from debugging.">var_export($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">error_log('order seen', 3, '/var/log/app.log')</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">debug_print_backtrace()</error>;
    phpinfo(INFO_GENERAL);
    <error descr="Debug output call; remove it if it was left over from debugging.">audit_dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">Tracer::dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">$probe->dump($order)</error>;
    $quiet->dump($order);

    ob_start();
    var_dump($order);
    ob_start();
    @print_r($order);
    ob_start();
    $copy = $order;
    <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($copy)</error>;
}
```

## Divergences

- **Name case (custos diverges).** Upstream compares function names, method
  names, the declaring class and `ob_start` case-sensitively, so
  `VAR_DUMP($x)`, `Debug::Dump($x)` or a listed method whose class is
  written in another case escape the check (and `OB_START();` does not
  cover the next call) although PHP treats them as the same symbols. custos
  compares all of them case-insensitively (D1, D3, D5, D6, E2, E3).
- Method wrappers (custos diverges from upstream). Upstream builds a
  method's identity with a `.` separator and compares it with entries
  written with `::`, so no method is ever recognised as a debug wrapper and
  the dumping code inside a listed method (`\Acme\Tracer::dump()`) is
  reported. That is a false positive: the method *is* the configured debug
  helper. custos compares the `\Class::method` form (E3).
- The debug-wrapper check uses only the function's short name, so a
  namespaced `function Acme\dd()` also counts as a wrapper. Kept on purpose:
  D1 matches calls by short name too (calls to `Acme\dd()` are reported as
  debug output), so treating that function as a debug helper is consistent;
  comparing FQNs here would only add reports inside helpers whose callers
  are already flagged.
- Closures: upstream compares the closure's internal name with the list;
  treat closures as never being wrappers.
- **custos diverges — `wp_die` and dumping methods.** Upstream's default
  list contains `wp_die`, WordPress's error-page/terminate helper used for
  permission and nonce guards (563 error findings on WordPress core, none of
  them debugging leftovers); custos drops it from the defaults (users can
  still add it through `configuration`). A method named like a function
  entry (`dump()`, `dd()` of a `Dumpable` trait or a query builder) is the
  dumping helper itself, so calls inside it are not reported (E3), like
  named functions. No upstream fixture uses `wp_die` or such methods.
