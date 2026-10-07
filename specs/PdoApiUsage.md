---
id: PdoApiUsage
group: Control flow
kind: semantic
needs: [names, types, hierarchy, index, stubs]
php: { min: "", max: "" }
---

# PdoApiUsage

## Summary
Preparing a statement with `PDO::prepare()` and immediately running it with an
argument-less `execute()` gains nothing from the prepare step: no parameters
are bound. A single `PDO::query()` call does the same job.

## Detection
- **D1** A method call (`->`, also `?->`) whose method name is exactly
  `execute` (case-insensitive, as PHP compares method names) with **zero**
  arguments.
- **D2** The call is the whole expression of an expression statement
  (`$st->execute();`). Calls used as values (`if ($st->execute())`,
  `$ok = $st->execute();`, `return $st->execute();`) are not inspected.
- **D3** Find the previous sibling statement in the same statement list,
  skipping whitespace and **all** comments in between (line comments, block
  comments and any number of doc comments). If none, stop.
- **D4** That statement is an expression statement whose top expression is an
  assignment (plain `=`; upstream also accepts compound assignment operators)
  whose right-hand side is **directly** a method call (no parentheses, no
  wrapping expression) named `prepare` (case-insensitive).
- **D5** The `prepare` call resolves to a method that is `\PDO::prepare`, or
  to a method declared in a class (not a trait) whose ancestry — the class
  itself, its parent classes and implemented interfaces — includes `\PDO`.
  This requires resolving the type of the receiver (e.g. a `\PDO`-typed
  parameter or property, `new PDO(...)`, a subclass of PDO). Unresolvable
  receivers are not reported.
- **D6** The assignment's left-hand side and the `execute` call's receiver
  are structurally equivalent expressions (same variable `$st`, same property
  chain `$this->stmt`, same array element `$q['a']`, …; whitespace and
  comments ignored).

The type of the `execute` receiver is not checked beyond D6.
- **Name case.** Wherever this rule compares two expressions for
  equivalence, the names PHP resolves case-insensitively — function and
  method names, class names in calls, `new`, `instanceof` and `::`
  accesses, and keywords — compare case-insensitively (`Cache::$map['k']`
  matches `cache::$map['k']`, `$o->Name()` matches `$o->name()`); variable,
  property and constant names stay case-sensitive.

## Exceptions (no report)
- **E1** `execute(...)` with any argument, even an empty array.
- **E2** `execute()` not used as a standalone statement.
- **E3** Any non-comment statement between the assignment and the
  `execute()` statement.
- **E4** `prepare` result wrapped (parenthesised, ternary, `?:`, function
  argument) or `prepare` called on a non-PDO object.
- **E5** Different target (`$a = $pdo->prepare(...); $b->execute();`).
- **E6** A bare `$pdo->query(...)` call is never reported.

## Report
- Range: the `execute()` method call expression, from the start of its
  receiver to the closing `)` (without the `;`).
- Severity: info (weak warning).
- Message: `No parameters are bound; call query() instead of prepare() + execute().`

## Fix
- **F1** Delete the whitespace run immediately preceding the `execute`
  statement (if the statement is preceded by whitespace), then delete the
  `execute` statement itself (including `;`). Comments between the two
  statements are kept.
- **F2** Rename the `prepare` method name in the assignment to `query`,
  keeping the receiver, operator and all arguments exactly as written.

Example: `$st = $db->prepare('SELECT 1');⏎    $st->execute();⏎` →
`$st = $db->query('SELECT 1');⏎`

## Options
None.

## PHP versions
None. The upstream fixture runs at the IDE test default level (below 7.1).

## Examples

```php
<?php
class Repo
{
    /** @var \PDO */
    private $db;
    private $stmt;

    public function __construct(\PDO $db) { $this->db = $db; }

    public function warmup(\PDO $conn)
    {
        $ping = $conn->prepare('SELECT 42');
        // ready to go
        /** note */
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$ping->execute()</weak_warning>;

        $this->stmt = $this->db->prepare('SELECT now()');
        <weak_warning descr="No parameters are bound; call query() instead of prepare() + execute().">$this->stmt->execute()</weak_warning>;

        $bound = $conn->prepare('SELECT * FROM t WHERE id = ?');
        $bound->execute([7]);

        $later = $conn->prepare('SELECT 1');
        log_it('x');
        $later->execute();

        $used = $conn->prepare('SELECT 2');
        if ($used->execute()) {
            return $used;
        }

        $conn->query('SELECT 3');
        return null;
    }
}
```

```php
<?php
class Repo
{
    /** @var \PDO */
    private $db;
    private $stmt;

    public function __construct(\PDO $db) { $this->db = $db; }

    public function warmup(\PDO $conn)
    {
        $ping = $conn->query('SELECT 42');
        // ready to go
        /** note */

        $this->stmt = $this->db->query('SELECT now()');

        $bound = $conn->prepare('SELECT * FROM t WHERE id = ?');
        $bound->execute([7]);

        $later = $conn->prepare('SELECT 1');
        log_it('x');
        $later->execute();

        $used = $conn->prepare('SELECT 2');
        if ($used->execute()) {
            return $used;
        }

        $conn->query('SELECT 3');
        return null;
    }
}
```

## Divergences
- The fix keeps all `prepare()` arguments; a second `prepare()` argument
  (driver options array) becomes `query()`'s fetch-mode argument, changing
  meaning. Recommendation: offer the fix only when `prepare()` has exactly one
  argument (still report). No upstream fixture has a second argument.
- Upstream accepts compound assignments in D4 (`$st .= $pdo->prepare(…)`),
  which is nonsensical; recommendation: accept only `=`. Not covered.
- If the receiver's type cannot be resolved to PDO (D5), nothing is reported;
  a typed `\PDO` parameter, property with `@var \PDO` docblock or `new \PDO`
  must be enough for an implementation to pass upstream fixtures (the fixture
  uses a `\PDO`-typed parameter).
- **Method-name case (custos diverges):** upstream compares `execute` and
  `prepare` case-sensitively, so `$st = $pdo->Prepare(…); $st->EXECUTE();`
  was missed although PHP calls the same methods. custos compares the method
  names case-insensitively (D1, D4); the fix still writes `query`.
- **Name case (custos diverges).** Upstream compares the expressions
  textually, so operands that differ only in the case of a function, method
  or class name (`Stats::$n` vs `stats::$n`), which PHP treats as the same,
  are not recognised as equivalent. custos folds the case of those names
  (Detection, "Name case").
