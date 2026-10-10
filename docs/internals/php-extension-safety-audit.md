# False-positive and quick-fix audit of the extension expansion

The audit covers the 100 rules listed in the
[expansion matrix](./php-extension-expansion-audit.md#rule-matrix), including all
16 rules offering quick-fixes. It reviews rule implementations against their
independent specifications, follows the shared proof queries, and probes branch
selection, mutation, aliases, named arguments, namespace shadowing, target PHP
versions, comments and assignment precedence. No EA source or fixture was read.

The audit found false positives and unsafe edits. The findings below have been
corrected and have regression fixtures. This is a bounded local analysis review;
it does not establish that arbitrary application code has no false positives.

## Findings and corrections

| Area | Reproduction and impact | Correction |
| --- | --- | --- |
| Signal fix placement | An unbraced outer `if` controlled an inner signal setter condition. Inserting a setter before the inner condition made that condition unconditional. | Withhold the fix outside top-level or braced statement lists. |
| Compression encoding | `gzuncompress(gzencode("hello", encoding: ZLIB_ENCODING_DEFLATE))` already decodes correctly. Argument counting missed the named encoding and changed it to a failing decoder. | Explicit encoding stays unknown for decoder mismatch, incremental inflate and matching-encoder failure proofs, including named arguments. |
| Decoder availability | A mismatched gzip decoder on PHP 5.3 could be replaced with unavailable `gzdecode`. | Retain the finding but withhold that replacement before PHP 5.4. |
| Offset mutation | Assigning seconds to an offset output variable after `getOffset` still triggered a division-by-1000 fix. | Discard proof after intervening uses, compound assignments, opaque calls, aliases, nested scopes or budget exhaustion. |
| SQLite constant shadowing | An inserted bare `SQLITE3_ASSOC` could resolve to an application constant with a different mode. | Resolve inserted constants through the same namespace-aware spelling used for replacements. |
| Parenthesized producers | Parentheses around a SQLite fetch or PostgreSQL escape bypassed checks for other consumers. A fix could remove needed numeric keys or change a separately emitted literal. | Unwrap producer parents before checking assignment consumers. |
| Captured consumers | A nested closure could consume numeric SQLite keys or PostgreSQL escaped literals, while the producer fix changed its captured value. | Include nested captures and references created before assignment in consumer safety checks; bound those traversals. |
| ZIP assignment precedence | Repairing `if (!$i = $z->locateName("x"))` without assignment parentheses stored a boolean instead of the entry index. | Parenthesize the assignment before the strict comparison. |
| Fix comments | Replacing a stream output statement deleted suffix comments. A negated ZIP test could lose a prefix comment, and a parenthesized GMP base replacement could erase an inner comment. | Remove only the output keyword and separating whitespace; withhold the ZIP edit when it loses a prefix comment; change only the GMP literal token. |
| Collection histories | Queue array writes, `offsetSet`, conditional insertions and escaping method/static arguments were ignored, allowing an incorrect empty-collection error. | Require a bounded, entirely modeled history; uncertain mutations remove empty proof. |
| Nonblocking wait comparisons | `!== 0` and `!= 0` were flagged even though they exclude the pending result. | Restrict detection to the specified `!= -1`, `!== -1` and `>= 0` subset. |
| ZIP failed branches | Removing a source or adding a fallback entry in an unsuccessful addition's `else` branch was treated as following a successful addition. | Require containment in the successful branch. |
| Signal getter mode | A local containing null was treated as a setter, and the proposed fix changed getter semantics. | Resolve null through local values and parentheses before inspecting setter use. |
| Selection scope | A watch array in another function, or an array subsequently replaced by null, established an incorrect preservation warning. | Use the selection's own variable scope and latest dominating assignment, with a proof budget. |
| SQLite runtime versions | Current PHP already resets before binding during execution. Execute alone also does not prove an active result cursor on older targets. | Exclude PHP 7.2 and later; older targets require a successful guarded row fetch from the prior execute result. |
| Promoted attributes | A property-only attribute on a constructor-promoted property was reported as an invalid parameter attribute. | Accept either property or parameter targets at the combined declaration. |
| Unreachable operations | A 76-file dead-code probe found 27 rules reporting operations after unconditional return. | Add reachability guards and own fixtures to those 27 rules. |
| Intent-sensitive defaults | Seek failure truthiness, signal previous-state checks, collation truthiness, raw wait-status comparisons and dispatch acknowledgements can be deliberate. | Make these five rules opt-in and document their policies. The expansion now has 70 enabled rules and 30 opt-in rules; the complete native catalogue has 201 enabled and 99 opt-in rules. |

For example, the ZIP repair now preserves both the condition and the assigned
index:

```php
<?php
$z = new ZipArchive();
if (($i = $z->locateName("x")) === false) {
    echo "missing";
}
```

An unbraced nested signal condition retains its diagnostic when explicitly
enabled, but has no automatic fix. Mixed-use PostgreSQL escaping retains its
identifier-use finding without changing its other consumers. Mixed-use SQLite
rows are excluded from the JSON shape rule.

## Runtime evidence

Local PHP 8.5 probes confirmed that the named deflate encoding round-trip returns
its original input, and that reflection can instantiate a property-only attribute
on a promoted property. They also confirmed successful SQLite rebinding without
an explicit reset on the modern runtime.

The SQLite version exclusion follows the
[documented rebinding history](https://www.php.net/manual/en/sqlite3stmt.bindparam.php).
The gzip format exclusion follows the
[encoder's explicit encoding contract](https://www.php.net/manual/en/function.gzencode.php).
PHP 7.2.14 and 7.3.0 introduced the reset behavior; custos uses minor-version
targets, so excluding all 7.2 targets conservatively avoids a patch-version
assumption. Old runtime binaries were not executed during this audit.

## Validation and limits

`make verify` passes, including architecture, pinned linters, full tests, own
fixtures, both 100% statement-coverage gates and the clean-room scan. The binary
and VitePress documentation site build successfully. Rule pages have been
regenerated.

The audit adds 82 regression inputs across 35 rule fixture directories.
Own regression fixtures check exact diagnostic ranges, expected fixed bytes,
parsing and fix stability. Runtime probes supplement those checks for the
contracts above. Allocation benchmarks exercise the new bounded collection,
offset and selection proofs, including parsing and engine work. Representative
local runs measured about 53 microseconds / 48 KB for a collection proof,
186 microseconds / 95 KB for an offset proof, and 38 microseconds / 48 KB for a
selection proof. These are whole-operation measurements, not latency budgets.

The final dead-code probe analyzes 76 files with zero findings. All 37 expected
fixed fixtures pass PHP 8.5 syntax linting. The expanded 531-file fixture corpus
accepts 840 individual fixes and combined fixes on 337 files without increasing
syntax errors. These corpus counts include applicable existing inspections as
well as the new rules. Parsing does not by itself establish semantic
safety; the producer-consumer and control-flow regressions address the specific
unsafe edits found here.

Unknown receivers, opaque validation, cross-file effects and unsupported control
flow remain conservative exclusions. Opt-in policies express application
preferences rather than proving intent from names or comments. A rule's absence
of findings does not certify runtime correctness or an archive's safety.
