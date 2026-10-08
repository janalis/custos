---
id: VariableFunctionsUsage
group: Performance
kind: semantic
needs: [types]
php: { min: "", max: "" }
---

# VariableFunctionsUsage

## Summary

`call_user_func()` / `forward_static_call()` with a callable that is known at
the call site can be written as a direct call (`$fn(...)`, `Cls::m(...)`,
`$obj->m(...)`), which is faster and analysable. Likewise
`call_user_func_array()` / `forward_static_call_array()` with a literal
argument array can pass the arguments inline.

## Detection

Applies to plain function calls that resolve to the global function (names
compared case-insensitively, as PHP does; `\` and a global `use function`
import are fine, but a same-named function declared in the current namespace,
one imported from another namespace, or a qualified non-global name such as
`Ns\call_user_func` does not count).

### Part A — inline the argument array

- **D1** Name is `call_user_func_array` (target `call_user_func`) or
  `forward_static_call_array` (target `forward_static_call`).
- **D2** Exactly 2 arguments, and the 2nd is an array literal (`array(...)` or
  `[...]`).
- **D3** The array has at least one element. Element values are collected in
  order; for `key => value` elements only the value is taken (the key is
  discarded).
- **D4** No element is by-reference (`&$v`, also `'k' => &$v`) and no element
  is a spread (`...$v`).
- **D4a** At PHP level **8.0 or above**, every element key (when present)
  must be an integer literal. A string key (or any non-literal key, which may
  be a string) is a named argument from PHP 8 on, so the array is not
  inlined. Below 8.0 keys are ignored by `call_user_func_array` and are
  simply dropped.

### Part B — direct call

- **D5** Name is `call_user_func` or `forward_static_call`, with at least 1
  argument. Let `A0` be the 1st argument.
- **D5a** No argument after `A0` is written with a call-time by-reference
  `&` (`&$b`, `& $b`): the direct call would carry it, which is a fatal
  error since PHP 5.4.
- **D6** Determine the callable parts `(first, second)`:
  - **D6a** `A0` is an array literal: collect its element values (as in D3).
    If none → no report. If the first element is a simple variable (incl.
    `$this`), its inferred type must be fully known and must not contain
    `string`; otherwise (unknown/partially unknown type, or a string member)
    → no report. `first` = 1st element, `second` = 2nd element (if any;
    further elements ignored).
  - **D6b** Otherwise: if `A0` is a simple variable and the PHP level is
    **below 5.4** → no report. `first` = `A0`, no `second`.
- **D7** `first` must be a string literal without interpolation or a simple
  variable; anything else (array access, property fetch, call, interpolated
  string, …) → **no report**, whatever `second` holds. Then compute the
  target text:
  - If `second` exists:
    - default method text = `{` + verbatim `second` text + `}`;
    - if `second` is a string literal without interpolation, let `c` be its
      unescaped value (single-quoted: `\\` → `\`, `\'` → `'`; double-quoted:
      usual escape sequences):
      - `c` starts with `parent::` (case-insensitive) → **no report**;
      - `c` contains a `:` anywhere → the target is the plain function text
        `c` (it overrides `first`'s text; `first` must still pass the check
        above);
      - otherwise the method text is `c`;
    - if `second` is a simple variable → method text is its verbatim text
      (`$m`);
    - an interpolated string or any other expression keeps the `{…}` form.
  - Class/function text from `first`:
    - string literal without interpolation → its unescaped value; when
      there is no `second` and that value starts with `parent::`
      (case-insensitive) → **no report** (same rule as the method part);
    - simple variable → its verbatim text.
- **D8** Arguments text: the call's arguments from the 2nd onward, each as
  verbatim text, prefixed with `...` when it is an argument unpacking (and
  with `name: ` for a named argument); joined by `, `.
- Replacement:
  - override case or no `second`: `{target}({args})`;
  - with `second`: `{class}{sep}{method}({args})` where `{sep}` is `::` if
    `first` is **not** a simple variable or the outer call is
    `forward_static_call`, and `->` otherwise.

(The parser must accept call-time pass-by-reference `&$x` inside argument
lists, as upstream fixtures contain it.)

## Exceptions (no report)

- **E1** Part A: argument array empty, contains `&` or `...` elements, is not a
  literal (variable, call), or the call has ≠ 2 arguments.
- **E1a** Part A at PHP ≥ 8.0: an element has a non-integer-literal key
  (`['to' => $to]`, `[$k => $v]`).
- **E1b** Part B: an argument carries a call-time `&`.
- **E2** Part B: `A0` is a variable and the level is below 5.4.
- **E3** Part B: array callable whose object element is a variable with an
  unknown or string-containing type.
- **E4** Part B: a callable string starting with `parent::`
  (case-insensitive), whether it is the method part of an array callable
  (`[$m, 'parent::send']`) or the whole callable (`'parent::send'`).
- **E5** Part B: class/function part that is not a non-interpolated string or
  a simple variable (`$list[$i]`, `$o->p`, `"$cls::m"`, `make()`), even when
  the method part contains `::` (`[$list[$i], 'Base::m']`), or an empty
  array.

## Report

- Range: the whole outer call (from the start of its name, including any
  qualifier, to its closing `)`; whitespace between name and `(` is inside).
- Severity: info (weak warning).
- Messages:
  - Part A: `Pass the arguments inline: '{replacement}'.`
  - Part B: `Call it directly: '{replacement}'.`

## Fix

- **F1** Part A: replace the call with
  `{target}({A0}, {values})` — `{target}` is the bare target name
  (`call_user_func` / `forward_static_call`, without the original
  qualifier), `{A0}` the verbatim 1st argument, `{values}` the element value
  texts joined by `, `. Example: `\call_user_func_array($h, [0 => $x, 2])` →
  `call_user_func($h, $x, 2)`; below PHP 8.0 also
  `call_user_func_array($h, ['a' => $x, 2])` → `call_user_func($h, $x, 2)`.
- **F2** Part B: replace the call with the D8 replacement. Examples:
  - `call_user_func($fn, $a, $b)` → `$fn($a, $b)`
  - `call_user_func('Pkg\\Tool::run', $x)` → `Pkg\Tool::run($x)`
  - `call_user_func([$svc, 'handle'], 1)` → `$svc->handle(1)`
  - `forward_static_call([$svc, 'handle'], 1)` → `$svc::handle(1)`
  - `call_user_func([$svc, "on{$evt}"])` → `$svc->{"on{$evt}"}()`
  - `call_user_func(['Base', 'Other::make'])` → `Other::make()`
- No parentheses are added around the replacement.
- **Name spelling.** A string callable always holds an absolute name, while
  the written replacement resolves against the namespace and imports at
  the call. Names taken from string literals are therefore written with a
  leading `\` when, written as-is at the reported position, they would
  resolve elsewhere:
  - a function name (no `::`): unqualified names get `\` when an
    unqualified call would not reach the global function (a `use function`
    import under that name, or a same-named function declared in the
    namespace); qualified names (`Pkg\run`) get `\` when they would resolve
    relative to the namespace or through a class import;
  - a class name (array callable `['Repo', 'm']`, or the part before `::`
    in `'Repo::m'`, including the D7 override string): `\` when the class
    name would resolve to another class (namespace-relative or imported
    alias); `self`/`static`/`parent` are kept;
  - an existing leading `\` is kept.
  F1's target name (`call_user_func` / `forward_static_call`) is written
  `\call_user_func` when an unqualified call would not reach the global
  function. Example in `namespace Shop`: `call_user_func('Pkg\Tool::run', $x)`
  → `\Pkg\Tool::run($x)`; `call_user_func(['Repo', 'find'], 1)` →
  `\Repo::find(1)`.

## Options

None.

## PHP versions

- Part B with a plain variable callable (D6b) requires level ≥ 5.4.
- Part A skips arrays with non-integer keys from level 8.0 (D4a).
  Everything else is ungated. Upstream fixtures run at 5.3 and 5.4.

## Examples

Level 8.0+:

```php
<?php
class Mailer {
    public static function send($to, $body) { return true; }
}
$m = new Mailer();

<weak_warning descr="Call it directly: '$hook($to, $log)'.">call_user_func($hook, $to, $log)</weak_warning>;
<weak_warning descr="Call it directly: 'Mailer::send($to, $body)'.">call_user_func(['Mailer', 'send'], $to, $body)</weak_warning>;
<weak_warning descr="Call it directly: '$m->send($to, $body)'.">call_user_func(array($m, 'send'), $to, $body)</weak_warning>;
<weak_warning descr="Call it directly: '$m::send(...$rest)'.">forward_static_call([$m, 'send'], ...$rest)</weak_warning>;
<weak_warning descr="Call it directly: '$m->$verb()'.">call_user_func([$m, $verb])</weak_warning>;
<weak_warning descr="Call it directly: 'App\Mailer::send($to)'.">\call_user_func('App\\Mailer::send', $to)</weak_warning>;
<weak_warning descr="Pass the arguments inline: 'call_user_func($hook, $to, $body)'.">call_user_func_array($hook, [0 => $to, 1 => $body])</weak_warning>;
<weak_warning descr="Pass the arguments inline: 'forward_static_call('Mailer::send', 1, 2)'.">forward_static_call_array('Mailer::send', array(1, 2))</weak_warning>;

call_user_func_array($hook, []);
call_user_func_array($hook, [&$to]);
call_user_func_array($hook, $args);
call_user_func_array($hook, ['to' => $to, $body]);
call_user_func($hook, $to, &$log);
call_user_func([$pool[0], 'Base::send'], $to);
call_user_func([$m, 'parent::send'], $to);
call_user_func([$pool[0], 'send'], $to);
call_user_func("{$cls}::send", $to);
call_user_func(build(), $to);
$name = 'mailer';
call_user_func([$name, 'send'], $to);
```

```php
<?php
class Mailer {
    public static function send($to, $body) { return true; }
}
$m = new Mailer();

$hook($to, $log);
Mailer::send($to, $body);
$m->send($to, $body);
$m::send(...$rest);
$m->$verb();
App\Mailer::send($to);
call_user_func($hook, $to, $body);
forward_static_call('Mailer::send', 1, 2);

call_user_func_array($hook, []);
call_user_func_array($hook, [&$to]);
call_user_func_array($hook, $args);
call_user_func_array($hook, ['to' => $to, $body]);
call_user_func($hook, $to, &$log);
call_user_func([$pool[0], 'Base::send'], $to);
call_user_func([$m, 'parent::send'], $to);
call_user_func([$pool[0], 'send'], $to);
call_user_func("{$cls}::send", $to);
call_user_func(build(), $to);
$name = 'mailer';
call_user_func([$name, 'send'], $to);
```

Level 7.4: `call_user_func_array($hook, ['to' => $to, $body])` →
`call_user_func($hook, $to, $body)` (keys ignored before PHP 8).

Level 5.3: `call_user_func($hook, 1)` is not reported; `call_user_func('trim', $s)`
→ `trim($s)` still is.

## Divergences

- **Keys in Part A (custos diverges):** upstream always drops the keys of
  the argument array. From PHP 8 on, string keys are named arguments, so the
  positional rewrite calls the function differently. custos keeps dropping
  keys below 8.0 (where they are ignored) and skips arrays with non-integer
  keys from 8.0 on (D4a). Upstream fixtures run below 8.0 and are
  unaffected.
- **Call-time pass-by-reference (custos diverges):** upstream copies a
  call-time `&` into the direct call (`$fn($a, &$b)`), which is a fatal
  error since PHP 5.4. custos does not report such calls (D5a).
- **`parent::` in a single string (custos diverges):** upstream only skips
  `parent::` when it is the method part of an array callable, so
  `call_user_func('parent::run')` is still rewritten to `parent::run()`.
  custos applies the same exclusion to a single callable string, and
  matches `parent` case-insensitively in both places (E4).
- **Receiver dropped with a `::` method string (custos diverges):** upstream
  applies the `:` override before checking `first`, so
  `[$list[$i], 'Base::m']` becomes `Base::m(...)` and the receiver (and its
  evaluation) disappears. custos requires `first` to be a string or simple
  variable in every case (D7, E5).
- **Leading `\` in a callable string** (`'\trim'`) is copied verbatim
  (`\trim(...)`), fine.
- **Function-name matching — custos diverges from upstream.** Upstream matches
  `call_user_func*`/`forward_static_call*` by the name as written
  (case-sensitive, any namespace qualifier, no resolution), so a differently
  cased call such as `Call_User_Func('trim', $x)` is missed while a namespaced
  or imported user function of the same name is reported (and rewritten) as if
  it were the builtin. custos matches case-insensitively and only calls that
  reach the global function.
- **Name spelling (custos diverges).** Upstream inserts the names from
  string callables (and the F1 target) as written. Inside a namespace that
  turns absolute names into relative ones (`'Pkg\Tool::run'` becomes a call
  to `Shop\Pkg\Tool::run`), and a namespaced or imported function of the
  same name captures an unqualified one. custos adds a leading `\` in those
  cases (Fix, "Name spelling").
