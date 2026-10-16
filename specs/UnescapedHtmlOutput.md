---
id: UnescapedHtmlOutput
group: Security
kind: semantic
needs: [names, types, flow]
php: { min: "", max: "" }
---

# UnescapedHtmlOutput

## Summary

Detect request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping.

## Detection

Bounded, pure constant conditions select only their executable flow successor.
The same applies to ternary arms. User constants, calls and unsupported scalar
coercions remain unknown; labels inside skipped branches can still be entered
by a legal jump. Lexical reachability queries retain their existing contract.

Literal variable-variable and global-table spellings of request superglobals
carry the same request provenance as direct reads. A function binding a global
can read request data written by another function of that file before its first
call; propagation follows longer chains within the shared analysis budget.

- D1. Report request-derived text emitted inside a proven HTML text-node context without context-appropriate escaping. Both `echo` and `print` are output statements.
- D1a. The context is proven from the bytes one statement emits before the request-derived part: concatenated literals, the literal fragments of a double-quoted string or heredoc (escape sequences decoded), and the earlier operands of a comma-separated `echo`. Those bytes must close an ordinary tag and leave no tag, attribute, comment, `script` or `style` element open.
- D1b. Dynamic parts emitted before the request-derived one are passed over when they cannot hold markup: a value typed integer, float or boolean anywhere in the statement, and a value that derives from tracked inputs and is escaped for HTML on every path when it is emitted into a text node. Such a part adds nothing to the proven bytes. Any other dynamic part (an unknown value, a call result, escaped text inside a tag, a value escaped on some paths only) could open an attribute or a raw-text element: nothing after it is examined.
- D1c. The first request-derived part found this way decides the statement; one finding is reported however many parts are unescaped. Later parts of the same operand must be free of side effects, because the whole operand is evaluated before any of it is emitted; later `echo` operands are not restricted.
- D2. Resolve builtin calls and named arguments before checking the contract. Project-local callees are considered only when their bodies or immutable summaries establish the required facts. The default is enabled, with positive proof required.
- D3. For flow-dependent conditions, require a reachable source-to-use path with no intervening invalidation. Unknown calls, ambiguous aliases, recursion or analysis-budget exhaustion cannot establish definite facts.

A catch contributes values only when its try body can throw. Bounded inert literal expressions, definite plain reads and proven scalar local writes do not enter a catch. Calls, unknown reads/accesses, output callbacks, discarded objects, aliases, dynamic bindings and tick declarations preserve conservative exception paths. Finally blocks still run along every applicable normal or abrupt exit.

## Exceptions (no report)

- E1. Trusted literals, htmlspecialchars/htmlentities with correct charset/flags and unknown output context are excluded. Markup emitted by an earlier statement or by inline HTML outside the statement does not establish a context.
- E1a. No report when the latest dominating literal `Content-Type` header in the same scope selects an inert media type: `text/plain`, `text/csv`, `application/json`, any `+json` type or `application/octet-stream`. Every other type, including `image/svg+xml` and `application/xhtml+xml`, can execute injected markup and does not suppress the report.
- E2. Unresolved targets, incomplete required facts, suppressed findings and PHP versions lacking the referenced API or language feature produce no report. User declarations shadowing builtin names are not builtin contracts.

- E3. Known boolean or numeric scalar output and resolved builtin `md5`, `sha1`, `hash` or `hash_hmac` hexadecimal output cannot inject the target text syntax. Raw binary or unknown output modes retain source taint. These narrow text proofs do not establish URL or filesystem-path authorization.
- E4. A value that only selects between results is judged by those results: a `match` or ternary is request-derived only when one of its arm results is; the subject and the arm conditions are not emitted. This also applies to a variable whose single reaching assignment is such an expression.
- E5. The emitted value is numeric or boolean, and therefore not reported, when it is the result of binary `+`, `-`, `*`, `/`, `%` or `**` (an array union is emitted as the constant text `Array`), of unary `-` or `+`, of a resolved builtin returning only an integer or boolean (`strlen`, `mb_strlen`, `count`, `sizeof`, `crc32`, `ord`, the `strcmp` family, the `strpos` family, `levenshtein`, `in_array`, `array_key_exists`, `strtotime`, every `is_*` and `ctype_*`), or of `sprintf` with a literal format whose conversions are all numeric (`b d e E f F g G o u x X`, plus `%%`). A format using `%s`, `%c` or a custom padding character keeps the taint. The proof covers only the emitted value itself: the same arithmetic passed on to another call, such as `implode`, stays request-derived.
- E6. No report when a request superglobal feeding the value (or `$GLOBALS`) is rewritten earlier in the same scope by a write that may replace the request text: a plain assignment of a value without request text to it or to one of its elements, `++`/`--`, `unset`, a `foreach` that iterates it by reference or assigns it the keys or elements of a value without request text, or a positional argument bound to a by-reference parameter of a resolved callee that may change it. A write to an element with a literal key counts only when the value reads that same element, the whole superglobal, an element with a non-literal key, or the superglobal through a callee; a write to the whole superglobal or with a non-literal or nested key always counts. A reference bound to the superglobal or to one of its elements (`$entry = &$_GET['v']`, `$_GET['v'] = &$local`) replaces nothing by itself; the same kinds of write to the bound variable then count as writes to what it is bound to: the whole superglobal, the element a literal key names, or an element under an unknown key (a computed or nested one). These do not count, whatever their target: an assignment that stores a value still holding request text (`$_GET = $_POST`, `$_GET = array_merge($_GET, $_POST)`, `$_GET['v'] = trim($_GET['v'])`, `$entry = $_POST['v']` through a reference), a compound assignment (`$_GET += $_POST`), a resolved builtin that fills its by-reference argument while another argument holds request text (`parse_str($_SERVER['QUERY_STRING'], $_GET)`), a write to another literal element, writes in nested closures, writes after the output statement and outside any loop around it, reads such as `count($_GET)`, and builtins that take the array by reference only to read it or move its pointer (`extract`, `current`, `key`, `end`, `reset`, `next`, `prev`, `each`). The result of a resolved function that takes a callback (`array_map('htmlspecialchars', $_GET)`) is not known to hold request text, so assigning it counts as a rewrite. The writes are read in the order they are written: one that stores request text into the whole superglobal fills every element again, one that stores it into an element with a literal key fills that element, and one that stores it under a computed or nested key, or through an entry of `$GLOBALS`, may fill any element, so `unset($_GET['v']); $_GET += $_POST;` leaves `$_GET['v']` reported, until a later replacement empties it again. Each slot of a destructuring target is a write of its own, and a `foreach` target assigned the keys or elements of request data is filled by it. A call or an object construction made after a replacement ends what the replacement established when its callee may store request text in a superglobal; a call inside the assigned value runs before the assignment and does not. A call written `f(...)` only makes a closure and stores nothing; a call given an array written with an element by reference (`f([&$_GET])`) may store anything, since the callee then holds the caller's variable. A resolved PHP function is taken not to store any (what it is given by reference is read as above) unless it runs code of the project: it is given a callback other than the name of a PHP function that only computes on text and numbers (a closed list of functions that raise no warning whatever text they are given: `strlen`, `trim`, `intval`, `strtoupper`, `strcmp`, `htmlspecialchars`, `md5` and the like, not `hexdec`, `abs` or `chr`); it calls a handler or callbacks registered earlier (`trigger_error`, `unserialize`, `gc_collect_cycles`, the session functions that start, write, reset, abort, destroy or collect a session, regenerate or create its id, or encode or decode it, `ob_flush`, `ob_clean`, `ob_end_flush`, `ob_end_clean`, `ob_get_flush`, `ob_get_clean`, `curl_exec`, `curl_multi_exec`, `xml_parse`, `xml_parse_into_struct`, `preg_replace_callback_array`); it loads a class by name (`class_exists` and its siblings, `spl_autoload`, `spl_autoload_call`, `class_alias`, `stream_wrapper_register`, `stream_filter_register`, and `method_exists`, `property_exists`, `is_subclass_of`, `is_a` asked to read text as a class name, `get_parent_class`, `class_implements`, `class_parents`, `class_uses`, `get_class_methods`, `get_class_vars`, unless their first argument is written `$this`, `new …`, `clone …`, `self::class` or `static::class`; `is_callable`, `constant` and `defined`, unless their argument is literal text without `::` or, for `is_callable`, a function written in place); it constructs a class it is given (`simplexml_load_string`, `simplexml_load_file`, `mysqli_fetch_object`, `pg_fetch_object` with a class argument); or it is a `filter_var`, `filter_input`, `filter_var_array` or `filter_input_array` call that names `FILTER_CALLBACK`, or whose filter is not named by one of PHP's own `FILTER_…` constants (a variable, a number or a constant of the project may stand for `FILTER_CALLBACK`); for the two functions that take a table of filters, their second argument, each filter of a table written in place is looked at, alone or under the key `'filter'`, and a table that is not written in place with constants and literals counts as naming any. Arguments given by name or unpacked may be any of these. A method or a constructor PHP itself declares is taken not to store any either, unless it is given such a callback; belongs to `Closure`, `Generator`, `Fiber` or the reflection classes; is `PDOStatement::fetchObject` or `mysqli_result::fetch_object` given a class, `PDOStatement::fetchAll` given a second argument, `ArrayObject::setIteratorClass`, `SplFileInfo::getFileInfo`, `getPathInfo`, `setFileClass` or `setInfoClass` given a class, `DOMDocument::registerNodeClass` given a class to register, `SoapServer::handle`, a method of an iterator that wraps another one (`IteratorIterator` and the PHP classes derived from it, `MultipleIterator`, `RecursiveIteratorIterator`, `RecursiveTreeIterator`), or the construction of an `ArrayObject` given the class of its iterators; or is a method, neither final nor a constructor, of a PHP class that a class of the project extends, which may call a method the project replaces (as inserting into an `SplHeap` calls its `compare`). Constructing a class that finds no constructor and no destructor runs nothing (before PHP 8 a method named like its class is a constructor, and is not looked into). A function, method or constructor of the analysed project does not either when the declaration that runs is known (a function declared once; for a method, as in the string-parameter rule of the failure inspections: final, private, named by its class, or replaced by no project class, each class involved declared once, every replacement counted too), takes no parameter by reference, and its body, the default values of its parameters and the closures written in it included, names no request superglobal, no `$GLOBALS` and no variable by a computed name (`${'_GET'}` is the superglobal), evaluates and includes no code, and calls or constructs nothing but what PHP declares, on the terms above, or a class of the project without constructor and destructor: a callee that itself calls a function or method of the project is not looked into further. Every other callee may: an unresolved function or method, a receiver of unknown class, a computed name, `new static`, an anonymous class, a function that calls another function or a method of the project. What is said here of PHP's own functions and methods is a limit of the analysis, not a property of PHP. Not seen: code that runs without a call being written (an object converted to text, counted, iterated, serialized or cloned by a PHP function, a magic method, a property hook, an autoloader started by a class name written in code); a class behind a stream wrapper path; an `assert()` callback; PHP functions registered on an `XSLTProcessor`; rows fetched into a class chosen by an earlier `setFetchMode()`; and a reference to a request superglobal kept elsewhere (a global variable, a static property or a property another function bound to it), which a callee can write through without naming the superglobal. A PHP function called directly in the scope of the output is not read as a refill, whatever it runs. When the output statement sits in a loop, what fills a superglobal later in that loop does so before the next turn and counts, while a replacement made after the output does not. After a replacement of the whole superglobal, the value is reported only when every element of it that still feeds the value was filled again.
- E7. The narrowing of E4 and E5 through a variable's reaching assignment is not applied, and the report is kept, when the scope can rebind that variable behind the assignment: a reference alias, a by-reference closure capture or argument, a by-reference `foreach` over it, a `catch` binding or `static` declaration of the same name, `extract()`, a variable variable, a `global` declaration of it or a `$GLOBALS` access that is not a different literal entry.
- E8. Request data that a `switch` compared with constant labels is the matched label inside an arm reached only through string labels that do not read as numbers (not empty and not starting, after whitespace, with a digit, a sign or a dot), and after the switch when every other way out replaced it or left the scope: `switch ($sort) { case 'name': case 'date': break; default: $sort = 'name'; }` validates `$sort`. A numeric, empty, boolean or null label, a label expression, a fall-through from another arm, a `default` arm and any other subject prove nothing. The subject is a plain variable or an element of a request superglobal read under one constant key (`switch ($_GET['action'])`, not `$_FILES`); the element is the matched label only while nothing that may run code is evaluated: a call, an object creation, an assignment, an `unset`, or the use of a local variable, property, offset of a local or class constant ends it, as does a reference to the superglobal. A matched label is text the program wrote: the parts of the output that follow it are still examined. So is a value derived from one by joining it with constant text or passing it through a builtin that only edits text (`$class = $sort . '-col'`, `strtoupper($sort)`): the output parts that follow such a value are examined as well. A `switch (true)` reads each case expression like the condition of an `if`/`elseif` chain.

## Report

- Range: the complete violating call or expression identified by D1; for a declaration-state violation, the offending property access or assignment. Highlight the final unsafe use for source-to-use findings.
- Severity: error.
- Message: `Escape untrusted text for its HTML output context.`

## Fix

None. Wrapping the value in `htmlspecialchars` is not behaviour-preserving in general: the page charset cannot be proven from one output statement, so a forced `UTF-8` with `ENT_SUBSTITUTE` may rewrite bytes of a page in another encoding, and an array-valued request parameter, emitted today as `Array`, would raise a `TypeError` on PHP 8. The escaping call, its flags and its charset are left to the author.

## Options

None. Enable or disable this rule through normal rule configuration.

## PHP versions

Apply only on PHP versions supporting the referenced APIs and syntax (within PHP 5.3–8.5). Builtin behavior and argument contracts follow the configured PHP version. Extension-specific calls require resolved extension API symbols; do not treat same-named application functions as extension calls.

## Examples

```php
<?php
<error descr="Escape untrusted text for its HTML output context.">echo '<p>' . $_GET['label'] . '</p>';</error>
<error descr="Escape untrusted text for its HTML output context.">echo "<li>{$_POST['title']}</li>";</error>
<error descr="Escape untrusted text for its HTML output context.">echo '<td>', $_GET['cell'], '</td>';</error>
<error descr="Escape untrusted text for its HTML output context.">print '<p>' . $_COOKIE['theme']</error>;

function row()
{
    $name = $_GET['name'];
    $page = (int) $_GET['page'];
    <error descr="Escape untrusted text for its HTML output context.">echo '<i>' . htmlspecialchars($name) . '</i><b>' . $name . '</b>';</error>
    <error descr="Escape untrusted text for its HTML output context.">echo "<a href=\"?page=$page\">$name</a>";</error>
}
```

Not reported:

```php
<?php
$order = match ($_GET['sort']) { 'desc' => 'DESC', default => 'ASC' };
echo '<p>' . $order . '</p>';
echo '<p>' . strlen($_GET['q']) . ' characters</p>';
echo '<p>' . ($_GET['page'] + 1) . '</p>';

function listing($row)
{
    // The unknown title may have opened an attribute or a script element.
    echo '<i>' . $row['title'] . '</i><b>' . $_GET['name'] . '</b>';
}

function export()
{
    header('Content-Type: text/csv');
    echo '<p>' . $_GET['label'];
}

function cleaned()
{
    $_GET['label'] = 'fixed text';
    echo '<p>' . $_GET['label'];
}
```

### Assertion assumptions

Only plain variable tests, literals, boolean operators, strict comparisons,
instanceof, isset/empty and native type predicates supply assertion guards.
A direct native flock result can prove that operation succeeded.
Conditions with property/array accesses, evaluated code, conversions or other
effects supply none; an earlier test may precede a rewrite.

A direct native `assert()` call is a guard only when the resolved project
setting `flow.assertions` is true. Its default is false, since PHP can disable
assertions. On PHP 7 and later, writes inside assertion arguments may be skipped:
a sanitizer or cast written there supplies no definite safety proof in the
default mode. Dynamic callable invocations evaluate arguments normally and do
not supply an assertion guard. With the setting enabled, failed assertions are
assumed to terminate. A later description expression that can rewrite the tested
value prevents using the condition as a guard. Before PHP 8, evaluated string
assertions invalidate local proofs; their PHP code is not interpreted.

## Divergences

Native custos rule. No upstream inspection or fixture comparison applies. These conditions are independently specified and examples are original.
