> **Status (2026-10-07):** all class **A** items and all class **B** items
> are resolved (spec updated, implemented, fixtures added; affected EA cases
> listed in `testdata/ea-divergences.json`; one B item — ClassConstantUsageCorrectness
> wrong-case imports — confirmed as correct upstream behaviour). The
> ConstantCanBeUsed and CascadeStringReplacement follow-ups below are fixed.
> Class C resolved as well (61 fixed, 8 declined — reasons in each spec's
> Divergences). Remaining work: the follow-ups listed at the end of this file.
> Follow-ups found while fixing class A:
>
> - ConstantCanBeUsed: `version_compare(PHP_VERSION, '7.1', '>')` is true on
>   7.1.0 but is rewritten to `PHP_VERSION_ID > 70100` (should be `>= 70100`;
>   likewise `<=` → `< 70100`).
> - CascadeStringReplacement: merging two literal search arrays with the same
>   string key loses an entry.

# Divergence candidates

Under the "Correctness over upstream fidelity" policy (`docs/internals/decisions.md`),
this file lists the items in the specs' *Divergences* sections where the spec
**kept** a suspected upstream bug or quirk (e.g. "keep for conformance",
"matches upstream", "the fixture depends on it"). Each of these could be
turned into a custos divergence.

Items are **not** listed when the spec already chose the correct behaviour:
the "Unstable variables" refinement and the other "custos refinement, not
upstream" entries, recommendations adopted because no fixture covers the
case, and PhpUnitDeprecations' corrected names. Items the spec describes as
correct, equivalent or "no action" are left out too.

A spot-check of `internal/inspection/rules/**` confirms that the items the spec left open
are implemented the upstream way. These are MkdirRaceCondition `|| !is_dir`,
SubStrUsedAsArrayAccess's inverted `??` gate, SubStrUsedAsStrPos's encoding
in the offset slot, VariableFunctionsUsage's call-time `&`, and
ClassConstantCanBeUsed D1/D2 with no class or `extends` guard.

Impact class:

- **A**: the fix changes program behaviour or produces invalid PHP.
- **B**: a false positive or a misleading report.
- **C**: a false negative, or cosmetic.

"Fixture?" is read from the spec's wording. **yes** means the spec says an EA
fixture expects or depends on it. **no** means the spec says no fixture covers
it. **unknown** means the spec says only "conformance" or "matches upstream".

| Rule | Problem | Class | Recommended correct behaviour | Fixture? |
|---|---|---|---|---|
| ArgumentUnpackingCanBeUsed | Unpacking a string-keyed array changes behaviour on PHP 8 (named args) and errors on 7.x | A | Skip, or give no fix, when the array may have string keys | unknown |
| ArrayIsListCanBeUsed | Loose `array_values($a) == $a` ignores key order, so it is not equivalent to `array_is_list` | A | Rewrite only the strict `===` form | unknown |
| ArrayPushMissUse | `array_push($a,$b)` → `$a[] = $b` differs when `$a` is not an array (ArrayAccess, string) | A | Report only when `$a` is known to be an array | unknown |
| CascadeStringReplacement | Spreading string-keyed arrays needs PHP 8.1, but upstream only gates on 7.4 | A | Gate the fix at ≥ 8.1 when the arrays may have string keys | unknown |
| ClassConstantCanBeUsed | `get_parent_class()` → `parent::class` in a class without a parent (fatal); `get_called_class()` outside a class | A | D2 only in classes with `extends`; D1 only inside class-likes | yes (D2) |
| ConstantCanBeUsed | `version_compare(PHP_VERSION,'8.1','==')` → `PHP_VERSION_ID === 80100` is not equivalent (patch level) | A | Do not rewrite `==`/`!=` forms, or give no fix | unknown |
| ConstantCanBeUsed | `get_class()` outside a class → `__CLASS__`, which is `''` there | A | Report only inside class-like bodies | no |
| ElvisOperatorCanBeUsed | `f() ? f() : x` → `f() ?: x` calls `f()` once instead of twice | A | Skip conditions that contain calls or other side effects | unknown |
| FixedTimeStartWith | Empty needle: `strncmp($s,'',0)===0` is always true, while `strpos` returned false on PHP < 8 | A | Skip empty needles (or gate at ≥ 8.0) | no |
| GetDebugTypeCanBeUsed | `get_debug_type()` returns different names than `gettype()` (`int`/`integer`, `float`/`double`, `null`/`NULL`) | A | Report without an auto-fix, or only where the result is not compared to gettype names | unknown |
| IfReturnReturnSimplification | Any binary operator accepted: `if ($a + $b) return true; return false;` → `return $a + $b;` (returns int) | A | Restrict to comparison, logical and `instanceof` conditions (or wrap in `(bool)`) | no |
| InstanceofCanBeUsed | `get_parent_class($o)=='X'` / `is_subclass_of` / `class_parents` → `instanceof`, which also matches `X` itself | A | Rewrite only exact equivalents (`is_a`, `get_class ===` with a final class); report the rest without a fix | unknown |
| IssetConstructsCanBeMerged | Moving the merged `isset` to the front reorders side-effecting fragments in the chain | A | Put the merged construct at the first hit's position | yes |
| MkdirRaceCondition | Temp-var fix emits `mkdir(...) \|\| !is_dir($concurrentDirectory)` (wrong negation) | A | Emit `\|\| is_dir($concurrentDirectory)`; record the fixture as a divergence | yes |
| NonSecureUniqidUsage | F2 changes semantics for `array_filter`/`array_reduce`/`array_walk` callbacks (different signatures) | A | Give no fix when `uniqid` is used as a callback string | unknown |
| NullCoalescingOperatorCanBeUsed | G2: `!empty($o) ? $o->p : 'x'` → `$o->p ?? 'x'` differs when `$o->p` is null | A | Skip the `empty` form unless the property is non-nullable | unknown |
| NullCoalescingOperatorCanBeUsed | G3: a probe that holds a non-object truthy value changes behaviour (warning before, fallback after) | A | Report only when the probe is known to be an object or null | unknown |
| NullCoalescingOperatorCanBeUsed | S3: `$v = f(); if (isset($w)) $v = $w;` → `$v = $w ?? f();`, so `f()` is no longer always evaluated | A | Skip when the previous value has side effects (calls, `new`, …) | unknown |
| ObGetCleanCanBeUsed | `ob_get_contents()` inside a closure defined in the previous statement is still rewritten | A | Require both calls to be direct statements of the same scope | no |
| OneTimeUseVariables | File-level code is not usage-counted, so it inlines variables that are reused or read in their own value (`$o = clone $o`) | A | Count usages at file level too; at least skip values that read `V` | yes |
| PhpUnitTests | The fix always uses `assertFileDoesNotExist`/`assertDirectoryDoesNotExist`, which do not exist before PHPUnit 9.1 | A | Below 9.1 suggest `assertFileNotExists`/`assertDirectoryNotExists` | yes |
| RealpathInStreamContext | R2 replaces `realpath('../x')` with `'../x'`, so resolution becomes include-path/CWD relative | A | Give no fix for relative literals (report only) | unknown |
| SenselessProxyMethod | A proxy that drops `return` (`parent::m($a);` for a value-returning parent) is reported; removing it changes the result | A | Do not report when the parent returns a value and the child does not | yes |
| SenselessTernaryOperator | Degenerate `$x === $x ? $x : $y` → `$y` changes behaviour | A | Require the operands of the condition to differ | no |
| SlowArrayOperationsInLoop | Generated `$xMax`/`$loopsMax` is not checked against existing variables (can clobber them) | A | Choose a fresh name not used in the scope | unknown |
| StaticClosureCanBeUsed | When a usage is unrecognised (returned, stored, …) all usage sites are discarded, so the closure is reported; making it static can break later binding | A | Treat unrecognised contexts as unsafe (except a direct `$v(...)` call) | no |
| StaticClosureCanBeUsed | Returned closures are reported, but callers may later bind them | A | Do not report closures that are returned | unknown |
| StaticClosureCanBeUsed | `self::instanceMethod()`/`static::instanceMethod()` inside the closure break once it is static | A | Treat calls to non-static methods via `self::`/`static::` as needing `$this` | unknown |
| StrEndsWithCanBeUsed | Empty needle: the original comparison is false (unless `$h` is empty) but `str_ends_with` returns true | A | Skip when the needle may be empty (literal `''`) | unknown |
| StrTrUsageAsStrReplace | `strtr` with a multi-char or empty `to` differs from `str_replace` | A | Report only when `to` is exactly one character | unknown |
| StringNormalization | Pattern A swaps `trim` with `ucfirst`/`lcfirst`/`ucwords`, which is not equivalent | A | Exclude the `ucfirst`/`lcfirst`/`ucwords` inner calls from pattern A | no |
| StringNormalization | Pattern B D5b drops `ucfirst`/`lcfirst`/`ucwords` nested in another of them (`lcfirst(ucwords($v))` → `lcfirst($v)`) | A | Do not collapse these three into each other | no |
| StrtotimeUsage | Fix drops the qualifier: `\strtotime('now')` → `time()` may call a namespaced `time()` | A | Keep a leading `\` (`\time()`) | unknown |
| SubStrShortHandUsage | `strlen`/`mb_strlen` mixed with `substr`/`mb_substr` (`mb_substr($s,0,strlen($s)-2)` → `-2`) | A | Require a matching byte/char length function | unknown |
| SubStrUsedAsArrayAccess | Inverted version gate: `?? ''` added below 7.0 (parse error), bare offset at ≥ 7.0 (warning, not `''`) | A | Emit `({access} ?? '')` at ≥ 7.0; no fix below 7.0; record the divergence | yes |
| SubStrUsedAsArrayAccess | `strlen($s) - n` can be negative; from PHP 7.1 negative offsets count from the end, unlike `substr` | A | Skip when the computed offset may be negative | unknown |
| SubStrUsedAsArrayAccess | `$s[$i]` with a non-integer string `$i` behaves differently from `substr` | A | Report only when the offset is known to be an int | unknown |
| SubStrUsedAsStrPos | 4-arg `mb_substr`: the encoding is emitted as `mb_strpos`'s offset argument | A | Emit `mb_strpos(a, O, 0, enc)`; record the divergence | yes |
| SubStrUsedAsStrPos | The length argument is not checked against `O` (`substr($s,0,3) == $longer`) | A | Require the length to be `strlen(O)` or to match the literal's length | unknown |
| SubStrUsedAsStrPos | `strtoupper(substr(...)) === $p` → `stripos(...) === 0`, which is only equivalent when `$p` is upper-case | A | Require `$p` to be an upper-case literal (lower-case for `strtolower`) | unknown |
| TernaryOperatorSimplify | Inverting `<`,`>`,`<=`,`>=` is not equivalent for NAN operands | A | Give no fix with inversion unless the operands are known not to be float | unknown |
| TypeUnsafeComparison | `'1e3'`, `' 1'`, `'1.'` treated as non-numeric get the strict fix (`1000 == '1e3'` was true) | A | Use PHP's full numeric-string grammar | unknown |
| UnnecessaryCasting | `(int)($a / $b)` with two ints reported; the division may return float, so removing the cast changes the result | A | Type `/` as `int\|float`; do not report | yes |
| VariableFunctionsUsage | Part A drops string keys, which are named arguments on PHP 8+ | A | On ≥ 8.0 emit `name: value` or skip arrays with string keys | yes |
| VariableFunctionsUsage | Call-time `$fn($a, &$b)` produced (fatal since 5.4) | A | Skip when any argument has a call-time `&` | yes |
| VariableFunctionsUsage | `first` not a variable or string while `second` has `::`: the receiver is dropped (`[$list[$i],'Base::m']` → `Base::m(...)`) | A | Skip this shape | unknown |
| AlterInForeach | A by-ref loop ending an `elseif` branch followed by `else` is reported even when `unset($v)` follows the `if` | B | Check the statement after the whole `if` chain | unknown |
| AlterInForeach | Part C reports `$a[$k] = 0` even when the value is already by-reference | B | Skip when the value is by-reference | unknown |
| AmbiguousMethodsCallsInArrayMapping | Side-effecting calls (`next($it)`) reported; extracting them changes behaviour | B | Skip calls with side effects (non-pure functions) | unknown |
| ArgumentUnpackingCanBeUsed | Name matched unresolved: a namespaced user `Foo\call_user_func_array` is reported | B | Match only global/unqualified-resolving calls | no |
| CallableMethodValidity | Protected methods reported even when `is_callable()` is called from within the hierarchy | B | Allow non-public methods when called from the class or a subclass | unknown |
| CallableParameterUseCaseInTypeContext | An unresolvable class in `R` is treated as incompatible (`new Unknown()` reported) | B | Treat unknown classes as unknown (no report) | unknown |
| ClassConstantUsageCorrectness | An import written in a different case than the declaration forces a report (textual replace fails) | B | Compare FQNs case-insensitively | unknown |
| ClassConstantUsageCorrectness | A class imported both plainly and under an alias: the plain, correct spelling is reported | B | Keep both the plain and alias entries in `L` | unknown |
| ClassMockingCorrectness | Class names inside parameter default values are inspected (part B) | B | Inspect only mock-creation arguments | unknown |
| CompactArguments | Arrow functions treated as a separate scope; names from the parent are reported as undefined | B | Arrow functions inherit the parent scope | no |
| CryptographicallySecureAlgorithms | Namespaced `Other\MCRYPT_DES` matched on its last segment | B | Match only global/unqualified constants | unknown |
| DisallowWritingIntoStaticProperties | `SELF::$p` is not recognised as `self` (case-sensitive), so it is reported in closures | B | Compare `self` case-insensitively | unknown |
| DynamicCallsToScopeIntrospection | R3 counts assignments regardless of order or reachability (even after the call) | B | Consider only assignments that reach the call | unknown |
| EfferentObjectCoupling | `\Foo` and `\foo` counted as two classes (inflated metric) | B | Deduplicate case-insensitively | unknown |
| ForgottenDebugOutput | Method wrappers are never recognised (`\Class.method` compared with `\Class::method`), so code inside them is reported | B | Compare in `\Class::method` form | unknown |
| GetClassUsage | Namespaced user `Some\get_class` matched by name | B | Match only global-resolving calls | unknown |
| HostnameSubstitution | Intermediate sanitising (`$base = str_replace(...)`) ignored | B | Stop tracking after a re-assignment through a sanitiser | unknown |
| ImplodeArgumentsOrder | `implode('a','b')` (both literals) reported and swapped | B | Skip when the first argument is also a string literal | no |
| InstanceofCanBeUsed | Namespaced user `App\is_a()` matched | B | Match only global-resolving calls | unknown |
| LoopWhichDoesNotLoop | Classes implementing `Iterator`/`IteratorAggregate` not exempt (only four exact names) | B | Exempt by type hierarchy | unknown |
| MagicMethodsValidity | Case-sensitive dispatch: `__ToString` reported as a non-magic `__` method | B | Match magic names case-insensitively | no |
| MagicMethodsValidity | A subclass instance returned from `__set_state` is reported | B | Accept subtypes of the expected class | unknown |
| MultiAssignmentUsage | Part B does not compare indexes (non-distinct or non-consecutive accesses reported) | B | Require distinct literal indexes | unknown |
| NotOptimalIfConditions | Method/static calls treated as reorderable, though they may have side effects | B | Treat method/static calls as non-reorderable; record the fixture | yes |
| NotOptimalIfConditions | D4 reports the specific type for `&&`, where the broader check is the redundant one | B | For `&&` report the broader check | unknown |
| NotOptimalIfConditions | `isset($x[f()]) && $x` reported although `$x` is the guarded array (S3 ignores `cur` itself) | B | Include `cur` in the S3 check | yes |
| NotOptimalRegularExpressions | Letters, digits, `\` and leading whitespace accepted as delimiters (PHP rejects them) | B | Reject invalid delimiters (no analysis) | unknown |
| NullPointerException | Assertion names / `is_null` compared case-sensitively (`AssertNotNull` does not stop the walk) | B | Compare case-insensitively | unknown |
| NullPointerException | Not flow-sensitive: a dereference inside a branch that implies non-null is still reported | B | Respect null checks in enclosing conditions | unknown |
| OffsetOperations | Membership ignores class hierarchy (a subclass index of the declared type is reported) | B | Accept subtypes | unknown |
| PhpUnitTests | `assertRegExp`/`assertNotRegExp` suggested at every version (deprecated since 9.1) | B | At ≥ 9.1 suggest `assertMatchesRegularExpression`/`assertDoesNotMatchRegularExpression` | unknown |
| PhpUnitTests | No test-case check: any class docblock and any `assert*` call on any receiver is examined | B | Restrict to TestCase subclasses (or classes ending in `Test`) | yes |
| PowerOperatorCanBeUsed | Namespaced user `App\pow()` reported | B | Match only global-resolving calls | unknown |
| PregQuoteUsage | Callee matched by name only | B | Match only global-resolving calls | unknown |
| ReferencingObjects | `callable`/`object` by-reference parameters reported (callables may be strings or arrays) | B | Exclude `callable` (and `object` when reassigned) | unknown |
| SecurityAdvisories | A mixed-case `name` (`Bakery/oven`) makes the project's own packages count as third-party | B | Lower-case the owner prefix as well | unknown |
| SenselessMethodDuplication | Static vs non-static mismatch between child and parent not checked | B | Skip when staticness differs | unknown |
| SimpleXmlLoadFileUsage | `Foo\simplexml_load_file()` (user function) matched, and the fix drops the qualifier | B | Match only unqualified/`\`-qualified calls | unknown |
| SlowArrayOperationsInLoop | Any binary operator qualifies (`$ok && count($a)` reported and rewritten) | B | Restrict to comparison operators | unknown |
| SlowArrayOperationsInLoop | G4a/G4b filters only look at the immediate block (`switch`/`try`/nested braces still reported) | B | Apply the filters through nested blocks | unknown |
| StrContainsCanBeUsed | Namespace ignored (`App\strpos` matches; fix gives `App\str_contains`) | B | Match only global-resolving calls | unknown |
| StrEndsWithCanBeUsed | Name checks ignore namespaces (also for `strlen`) | B | Match only global-resolving calls | unknown |
| StrStartsWithCanBeUsed | Name check ignores namespaces; the generated call keeps a foreign qualifier | B | Match only global-resolving calls | unknown |
| StringsFirstCharactersCompare | Leading-zero length `010` read as decimal (PHP reads octal 8) | B | Parse integer literals per PHP (octal/hex/binary) | unknown |
| StrtotimeUsage | Namespaced user function named `strtotime` reported | B | Match only global-resolving calls | unknown |
| SuspiciousAssignments | D16 counts a compound assignment in the preceding `if` as the conditional write (`$v .= 'x'` then `$v = ''`) | B | Accept only plain `=` | no |
| SuspiciousAssignments | `$v = 1; $v = $v;` reported as an overwrite | B | Skip when the RHS reads the target | unknown |
| SuspiciousBinaryOperation | D6 with an empty right-operand type set makes almost any comparison qualify | B | Treat an empty type set as unknown (no report) | unknown |
| SuspiciousReturn | `return` inside closures/anonymous classes in the `try` counts (`try { $f = function () { return 1; }; } finally { return 2; }`) | B | Stop the search at function boundaries | no |
| ThrowRawException | Only the class's own `$message` counts; a parent user class presetting it does not suppress the report | B | Walk user parent classes for `$message` | unknown |
| TypeUnsafeArraySearch | User `App\in_array` with two arguments reported | B | Match only global-resolving calls | unknown |
| TypeUnsafeComparison | Kind M reported for `?Invoice` where the class lacks `__toString()` | B | Skip null-plus-class unions without `__toString` | unknown |
| UnSafeIsSetOverArray | `object` pseudo-type treated like a scalar in D10 (an array_key_exists suggestion on objects) | B | Treat `object` as an object type | unknown |
| UnnecessaryAssertion | `assertInternalType` reported regardless of the requested type, even when it contradicts the return type | B | Report only when the requested type matches the declared type | yes |
| UnqualifiedReference | Part B checks the outer function, not the callback, so unknown callback names are reported | B | Resolve the callback name; report only known global functions | unknown |
| UntrustedInclusion | Stream-wrapper (`phar://`), `./`/`../` and UNC paths reported | B | Skip stream-wrapper and UNC paths | unknown |
| UnusedConstructorDependencies | Assignment targets inside closures in the constructor are reported | B | Ignore closure bodies when counting constructor references | no |
| UselessUnset | By-ref parameter rebound with `global`/`static` then `unset` is still reported | B | Skip when the name is rebound by `global`/`static` | no |
| UsingInclusionReturnValue | `@include 'x.php';` (statement with the silence operator) reported | B | Look through `@` when deciding statement context | unknown |
| AccessModifierPresented | `final const X` (implicit visibility) not reported | C | Report `final` constants without visibility | unknown |
| AliasFunctionsUsage | Case-sensitive name match (`SIZEOF()` not reported) | C | Match case-insensitively | no |
| AlterInForeach | Unset check uses the first argument: `unset($a['k'], $v)` reports nothing | C | Test each unset argument individually | unknown |
| AmbiguousMethodsCallsInArrayMapping | Parenthesised assignment statement `($m[f()] = f());` skipped | C | Look through parentheses | unknown |
| ArraySearchUsedAsInArray | Case-sensitive function name | C | Match case-insensitively | no |
| ArrayUniqueCanBeUsed | Case-sensitive function names | C | Match case-insensitively | no |
| ArrayUniqueCanBeUsed | The fix drops namespace qualifiers (`\count` → `count`) | C | Preserve the qualifier | unknown |
| CascadeStringReplacement | A 4th `$count` argument suppresses all reports, even part C | C | Suppress only the parts that would drop `$count` | unknown |
| CascadingDirnameCalls | D2 skips a dirname call that is the second argument of an outer dirname | C | Skip only chain members | unknown |
| CaseInsensitiveStringFunctionsMissUse | Escapes not decoded (`"\x2F"` not reported) | C | Decode escapes before the letter test | unknown |
| CompactCanBeUsed | Escapes in keys not decoded (`"t\x61g" => $tag` not matched) | C | Decode key literals | unknown |
| CryptographicallySecureRandomness | The false-check test is name-based and position-insensitive (a check in an unrelated branch counts) | C | Require the check to guard the result | unknown |
| DateIntervalSpecification | A literal far from `new`, reached through value discovery, is highlighted, once per call site | C | Deduplicate; report at the call | unknown |
| DeprecatedConstructorStyle | Class and method names compared case-sensitively | C | Compare case-insensitively | unknown |
| DeprecatedIniOptions | Table frozen at PHP 7.3 | C | Extend with later deprecations/removals | unknown |
| DirectoryConstantCanBeUsed | Magic constant/function names compared case-sensitively | C | Compare case-insensitively | unknown |
| DisconnectedForeachInstruction | D7 marks variables as modified on pure reads (`echo $cfg['x']`) | C | Count only writes | unknown |
| DisconnectedForeachInstruction | Catch variables and inner-loop header variables treated as reads, not writes | C | Treat them as writes | unknown |
| DuplicateArrayKeys | Raw literal comparison: `"\x41"` vs `'A'`, `'1'` vs `1` not detected | C | Compare PHP-normalised key values | unknown |
| DynamicCallsToScopeIntrospection | Callback names matched case-sensitively | C | Match case-insensitively | unknown |
| EncryptionInitializationVectorRandomness | In D5 any call named like a secure function counts, whatever its receiver (`$rng->random_bytes`) | C | Accept only global function calls | unknown |
| EncryptionInitializationVectorRandomness | With 6–8 `openssl_encrypt` arguments the IV is not checked | C | Check the IV for ≥ 5 arguments | unknown |
| FixedTimeStartWith | Qualifier dropped (`\strpos` → `strncmp`) | C | Preserve a leading `\` | unknown |
| ForeachInvariants | A limit variable reused by several loops is never accepted (all assignments in the function are collected) | C | Consider only the assignment reaching the loop | unknown |
| ForgottenDebugOutput | A namespaced `function Acme\dd()` counts as a debug wrapper | C | Compare FQNs | unknown |
| GetClassUsage | The null-check search is order-based and branch-insensitive | C | Require a dominating null check | unknown |
| HostnameSubstitution | Left/middle concatenation operands report a different node or nothing | C | Report the actual occurrence | unknown |
| IncorrectRandomRange | Hex/octal/binary/float bounds skipped | C | Parse PHP numeric literals | unknown |
| InfinityLoop | Receivers compared literally (`SELF::flush()`, `Static::boot()` missed) | C | Compare keywords case-insensitively | no |
| InstanceofCanBeUsed | `class_implements()` with interface names (the realistic use) never reported | C | Accept interfaces for `class_implements` | unknown |
| InvertedIfElseConstructs | The D4b result keeps `false` as an operand (`false !== x()`) | C | Emit the simplified positive form | unknown |
| IsEmptyFunctionUsage | The property-access guard descends into call arguments (`empty(make($o->p))` skipped) | C | Stop the descent at call boundaries | unknown |
| IssetArgumentExistence | First-mention scan enters nested closures (an inner use suppresses outer reports) | C | Stop at closure boundaries | unknown |
| LoopWhichDoesNotLoop | `continue` inside a `switch` taken as continuing the loop (PHP treats it as `break`) | C | Count `switch` as a loop level | no |
| MissUsingParentKeyword | D3 compares names case-sensitively | C | Compare case-insensitively | unknown |
| MkdirRaceCondition | D6 does not compare the `is_dir()` argument with `mkdir()`'s | C | Require the same directory expression | unknown |
| NotOptimalIfConditions | Cost model is case-sensitive (`IS_ARRAY()` costs 5) | C | Look up names case-insensitively | unknown |
| NotOptimalRegularExpressions | A literal reached from several calls is reported once per call | C | Deduplicate by range | unknown |
| NotOptimalRegularExpressions | 1-argument `preg_quote` reported only when the literal looks delimited | C | Report every 1-argument call | unknown |
| PackedHashtableOptimization | Keys outside the 32-bit range stop the analysis | C | Use 64-bit integer keys | unknown |
| PackedHashtableOptimization | Hex/octal/binary number keys stop the analysis | C | Evaluate PHP integer literals | unknown |
| PhpUnitTests | Superfluous slots padded with a literal `null` (`assertEmpty($x, 'm', null)`) | C | Append the remaining original arguments | unknown |
| PhpUnitTests | Function names matched case-sensitively (`Count($x)` missed) | C | Match case-insensitively | unknown |
| PotentialMalware | D5a/D2 case-insensitive, D6/D7 case-sensitive | C | Match function names case-insensitively throughout | unknown |
| PrintfScanfArguments | D4 skips every format using `%n$` + a conversion letter (argument-count errors missed) | C | Restrict the interpolation test to double-quoted strings; parse `%n$` | unknown |
| SecurityAdvisories | `mikey179/vfsStream` can never match (list entry not lower-cased) | C | Compare case-insensitively | no |
| ShortListSyntaxCanBeUsed | `list()` used as a sub-expression is skipped | C | Convert it too | unknown |
| SimpleXmlLoadFileUsage | Case-sensitive name match | C | Match case-insensitively | no |
| SlowArrayOperationsInLoop | `size` (not a PHP function) listed while `sizeof` is missing | C | Replace `size` with `sizeof` | unknown |
| StaticLambdaBinding | An arrow fn using `$this` nested in a static closure not reported (fails at runtime) | C | Attribute arrow-fn `$this` to the enclosing closure | unknown |
| StaticLambdaBinding | Only the first offending `$this`/`parent::` per function is reported | C | Report each occurrence (one fix) | unknown |
| StrTrUsageAsStrReplace | Case-sensitive name match | C | Match case-insensitively | unknown |
| StrlenInEmptyStringCheckContext | Only reported inside a function or class (E1, looks accidental) | C | Also report at file level | unknown |
| SubStrShortHandUsage | 4-argument `substr` emits `null` plus the extra argument | C | Skip calls with too many arguments | unknown |
| SuspiciousBinaryOperation | D10a considers only `&&`/`\|\|` (mixing with `and`/`or`/`xor` missed) | C | Include `and`/`or`/`xor` | unknown |
| SuspiciousFunctionCalls | Case-sensitive name match | C | Match case-insensitively | unknown |
| ThrowRawException | Parenthesised operand `throw (new \Exception('x'))` not reported | C | Look through parentheses | unknown |
| TraitsPropertiesConflicts | Only default values compared; visibility/`static`/`readonly`/type mismatches (fatal) missed | C | Compare all of PHP's compatibility criteria | unknown |
| TraitsPropertiesConflicts | Check A is silent for incompatible duplicates (only compatible ones get a weak warning) | C | Report incompatible duplicates as errors, as check B does | unknown |
| TypeUnsafeArraySearch | `'-1'`, `'1.5'` treated as non-numeric (E1 skip) | C | Use the full numeric-string grammar | unknown |
| TypeUnsafeComparison | `\Closure` listed as comparable but never matches | C | Normalise the entry so it matches | unknown |
| UnSafeIsSetOverArray | D9 inspects only the last `[...]` (`$m['a'.$b]['c']` missed) | C | Inspect every index level | unknown |
| UnSafeIsSetOverArray | `isset($this->promoted)` on constructor-promoted properties not reported | C | Resolve promoted properties as declarations | unknown |
| UnnecessaryUseAlias | Function-name comparison is case-sensitive | C | Compare function names case-insensitively | unknown |
| UnserializeExploits | A direct `unserialize($_POST)` is not reported (only `$_POST['x']`) | C | Report superglobals taken as-is | unknown |
| UnserializeExploits | `['allowed_classes' => true]` accepted although it allows every class | C | Treat `true` as unsafe | no |
| UnusedGotoLabel | A `goto` in a nested closure marks an outer label as used (impossible jump) | C | Stop the search at function boundaries | no |
| UsingInclusionOnceReturnValue | Requires lowercase `_once` (`REQUIRE_ONCE $f` missed) | C | Match keywords case-insensitively | no |
| VariableFunctionsUsage | `call_user_func('parent::run')` rewritten to `parent::run()` (`parent::` checked only for the method part) | C | Apply the `parent::` check to single-string callables | unknown |

## Summary

| Class | Count | Pinned by an EA fixture (yes) |
|---|---|---|
| A: changes behaviour / invalid PHP | 46 | 11 |
| B: false positive / misleading | 59 | 4 |
| C: false negative / cosmetic | 69 | 0 |
| **Total** | **174** | **15** |

Fixing an item marked **yes** needs a new entry in `testdata/ea-divergences.json`.
The rest (fixture "no" or "unknown") can be fixed without known conformance
cost, but run the EA conformance suite to confirm. The class A items marked
**yes** are worth doing first, because they ship broken fixes today:
MkdirRaceCondition, SubStrUsedAsArrayAccess, SubStrUsedAsStrPos (encoding),
VariableFunctionsUsage (keys, `&`), PhpUnitTests (pre-9.1 names),
ClassConstantCanBeUsed D2, OneTimeUseVariables, UnnecessaryCasting,
IssetConstructsCanBeMerged and SenselessProxyMethod.

## Follow-ups found during class B

- MultiAssignmentUsage: `$row = $row[0]; $b = $row[1];` (base overwritten in between) and `load()[0]; load()[1]` (destructuring changes call count) may be false positives.
- **Done (2026-10-07).** StrlenInEmptyStringCheckContext (class A, found during class C): the fix inserts the argument unparenthesised, so `strlen($p ?: $q) > 0` becomes `'' !== $p ?: $q` (wrong precedence).
- **Done (2026-10-07).** PrintfScanfArguments (class B, found during class C): when the format can be a literal or something else (`$message ?: 'fmt %s'`, nowdoc alternatives), only the literal branch is checked → false positives (infection NoSourceFound.php:102, webmozart Assert.php:377).
- **Done (2026-10-07).** Function names still matched case-sensitively (class C): CaseInsensitiveStringFunctionsMissUse, CascadingDirnameCalls, AmbiguousMethodsCallsInArrayMapping (called-name comparison).

## Follow-ups found during the name-matching audit

- **Done (2026-10-07).** Fixes emitting an unqualified builtin name that a same-named namespaced function would capture: RealpathInStreamContext (`dirname(`), MkdirRaceCondition (`is_dir(`), ArgumentUnpackingCanBeUsed (bare callee). Emit `\name(` when a bare name would not reach the global function. Shared helpers `util.QualifiedBuiltin`/`QualifiedBuiltinFor` (+ `util.StringCallableFunction`/`StringCallableClass` for string callables); the audit fixed 28 rules in all.
- **Done (2026-10-07).** Namespaced constants accepted as builtins: ConstantCanBeUsed (`Foo\PHP_VERSION`), DateTimeConstantsUsage (`X\DATE_ISO8601`). Resolved via `util.GlobalConstName`; inserted constants spelled with `util.QualifiedGlobalConst`.
- **Done (2026-10-07).** `util.Equivalent` compares function/method/class names case-sensitively; switch callers to `util.EquivalentFoldNames` where names are compared. 28 rules switched (incl. NotOptimalIfConditions).
- **Done (2026-10-07).** UnqualifiedReference checks the imported name, not the alias (`use function strlen as sl;`).
- **Done (2026-10-07).** TypesCastingCanBeUsed compares the `settype` type string case-sensitively.
- **Done (2026-10-07).** NotOptimalIfConditions looks up its cost table by written name (affects suggested order only).
