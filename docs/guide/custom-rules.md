# Custom rules

A project can declare its own rules in `custos.json`: a PHP code pattern to
find, a message, and optionally a quick-fix. Custom rules run with the
built-in ones in `custos analyse`, `custos fix`, the language server and the
MCP server.

```json
{
  "customRules": [
    {
      "id": "NoDumpInSrc",
      "match": ["\\dump(...)", "\\dd(...)"],
      "paths": ["src/**"],
      "exclude": ["src/Debug/**"],
      "message": "Remove the debug call",
      "severity": "error",
      "fix": ""
    },
    {
      "id": "UseClock",
      "match": "new \\DateTimeImmutable()",
      "message": "Get the time from the clock",
      "description": "Time comes from the injected clock so tests can freeze it.",
      "fix": "$this->clock->now()",
      "fixTitle": "Use the clock"
    }
  ]
}
```

Backslashes are doubled because the patterns are JSON strings:
`"\\dump(...)"` is the pattern `\dump(...)`.

## Rule fields

| Field | Meaning |
|---|---|
| `id` | Rule ID: an uppercase letter, then letters and digits. It must not name a built-in rule. |
| `match` | One pattern or a list of them; the rule reports code matching any. |
| `message` | The finding's message. A metavariable in it (`$X`) quotes the code it matched. |
| `severity` | `"error"`, `"warning"` (default) or `"info"`. |
| `enabled` | `false` turns the rule off by default; `--rule` still runs it. Default `true`. |
| `description` | Longer text shown by `custos explain` and in SARIF output. Default: the message. |
| `php` | Minimum target PHP version, such as `"8.1"`: below it the rule does not run. |
| `paths` | Globs of project-relative paths the rule checks. Default: every analyzed file. |
| `exclude` | Globs of project-relative paths the rule skips. |
| `inside` | Code the match must lie within (see [Context](#context)). |
| `notInside` | Code the match must not lie within. |
| `where` | Constraints on metavariables (see [Constraints](#constraints)). |
| `fix` | Replacement code (see [Quick-fixes](#quick-fixes)). |
| `fixTitle` | Title of the quick-fix in editors. |

In globs, `*` matches within one path segment and `**` matches any number of
segments; `src/` stands for `src/**`, and a leading `./` is ignored.

A custom rule is configured and suppressed like any other rule, by its ID:
`rules` entries in `custos.json` override its severity or turn it off, and
`@custos-ignore NoDumpInSrc` silences it. `custos rules` lists custom rules in
the `Custom` group, and `custos explain NoDumpInSrc` describes one.

### Rule files

Rules can also live in separate JSON files, each holding a list of rule
objects. `customRuleFiles` names them, relative to the project root:

```json
{ "customRuleFiles": ["custos-rules/debug.json", "custos-rules/time.json"] }
```

custos stops with a configuration error (exit code `2`) when a rule is
invalid: a pattern that is not PHP, a duplicate or built-in ID, a metavariable
the message, fix or constraints use but a pattern does not bind, an invalid
glob, type or regular expression.

## Patterns

A pattern is one PHP expression or statement, written as in the code it
finds. Names in patterns are fully qualified (`\dump`, `\App\Clock`), and
custos compares them with the names the code resolves to: an imported alias,
a `use function` import or a class written relative to its namespace still
match. An unqualified function call in a namespace matches the global
function unless the namespace declares one of that name, as PHP decides.
`self`, `static` and `parent` match the class they stand for.

Other parts compare by meaning: string literals by value (`'a'` matches
`"a"`), method names case-insensitively, and extra parentheses in the code
are ignored. `new Foo` and `new Foo()` are the same, as are `[]` and
`array()`, `exit` and `die`, braces and the alternative syntax, and modifiers
in any order. A closure written without `use` matches one with any.

### Metavariables

A variable in uppercase, such as `$X` or `$ARGS`, is a metavariable: it
matches any expression and remembers it. A metavariable used twice must match
the same code both times.

| Pattern | Matches |
|---|---|
| `$X == null` | any loose comparison with `null` |
| `$X === $X` | an expression compared with itself |
| `$this->repository->find($ID)` | that call with one argument |
| `\in_array($NEEDLE, $HAYSTACK)` | `in_array()` with exactly two arguments |
| `\dump(...)` | `dump()` with any arguments |
| `\sprintf($FORMAT, ...$ARGS)` | `sprintf()` with at least one argument; `$ARGS` holds the others |
| `$_->save()` | `save()` on anything (`$_` matches without remembering) |

Lowercase variables match that variable only: `$this->x` matches `$this->x`
and not `$that->x`.

### Statements

In a block, a metavariable alone as a statement (`$S;`) matches any one
statement, and `$_;` matches any number of statements, none included:

| Pattern | Matches |
|---|---|
| `if ($C) { $S; }` | an `if` whose block holds exactly one statement |
| `foreach ($A as $V) { $_; $V->save(); $_; }` | a loop that saves its value somewhere in its body |
| `try { $_; } catch (\Exception $E) { }` | a `try` whose `catch (\Exception)` is empty |

A pattern of several statements matches them in a row, in any block or at
the top of a file:

```json
{
  "id": "InlineReturnedCount",
  "match": "$N = \\count($A); return $N;",
  "message": "Return the count directly",
  "fix": "return \\count($A);"
}
```

`$_;` between them allows other statements in between:
`$F = \fopen($P, $M); $_; \fclose($F);` finds a file opened and closed in
the same block.

Each `$_;` may try many ways to split a block, so the work of matching one
statement list is bounded: past it, custos stops and reports the file's
[analysis as incomplete](./cli#analysis-status) instead of a finding.

## Context

`inside` lists patterns of which one must match code around the match, and
`notInside` patterns none may match. They share the match's metavariables:

```json
{
  "id": "SaveInLoop",
  "match": "$V->save()",
  "inside": ["foreach ($A as $V) { $_; }"],
  "message": "Saving $V once per iteration; save the batch after the loop"
}
```

Here `$V` must be the loop's value, so `$other->save()` in the same loop is
not reported. With `"notInside": ["$DB->transaction(function () { $_; })"]`,
a rule skips code within a transaction callback.

## Constraints

`where` restricts what a metavariable may match:

```json
{
  "id": "FormatDateLocale",
  "match": "$D->format($F)",
  "where": {
    "$D": { "type": "\\DateTimeInterface" },
    "$F": { "regex": "^'[^']*[DlMF]" }
  },
  "message": "Day and month names in $F are always English; use IntlDateFormatter"
}
```

| Constraint | The metavariable matches only when |
|---|---|
| `type` | the type custos infers for it is always this PHPDoc type (fully qualified classes) |
| `regex` | its source code matches this regular expression |
| `not` | it matches none of these patterns |

A `type` constraint makes custos index the project, as the type-checking
rules do.

## Quick-fixes

`fix` is the code that replaces the match; the metavariables in it stand for
the code they matched:

```json
{
  "id": "StrictNullCompare",
  "match": "$X == null",
  "message": "Compare $X with ===",
  "fix": "$X === null"
}
```

custos adds parentheses where the replaced code needs them, so `$X * 2` with
`$X` matching `$a + $b` becomes `($a + $b) * 2`. It offers no quick-fix when
the result would not parse or would delete a comment. A metavariable inside a
string in `fix` is refused: it would not stand for the code it matched.

A statement metavariable (`$S;`) in `fix` stands for the statement it
matched. An expression `fix` may replace a statement match: it becomes an
expression statement. Statements cannot replace an expression.

An empty `fix` deletes the matched statements, or the statement made of the
matched expression, such as a `dump($x);` line. It offers none when the expression is part of a larger
one, or when the statement is the body of an `if` or a loop without braces.

## Testing rules

`custos test-rules` runs each rule over example files and checks what it
reports. Put them in `custos-tests/` (or the directory `customRuleTests`
names), one directory per rule ID:

```text
custos-tests/
└── NoDumpInSrc/
    ├── basic.php          # code, with the expected findings marked up
    ├── basic.fixed.php    # optional: the code after the quick-fixes
    └── basic.json         # optional: settings
```

In a fixture, wrap each expected finding in a tag named after its severity
(`error`, `warning` or `info`) with the exact message in `descr`; code that
is not wrapped must not be reported:

```php
<?php
<error descr="Remove the debug call">dump($order)</error>;
var_dump($order);
```

`basic.fixed.php` holds the expected result of applying the quick-fixes;
whitespace differences are ignored. `basic.json` may set `php`, the target
version, and `path`, the project path the rule sees: a fixture of a rule
limited to `src/**` needs `{"path": "src/Example.php"}`. `analyse` and `fix`
skip `custos-tests/`, whose fixtures hold markup; when `customRuleTests`
names another directory, add it to `exclude`. The command lists
each fixture as `ok` or `FAIL` with what differs, and exits with `1` when
one fails.

## Rules written in Go

When a pattern is not enough, write the rule in Go and build your own custos
binary. The package `github.com/janalis/custos/pkg/custos` runs the whole
command line (`analyse`, `fix`, `explain`, `rules`, `lsp`, `mcp`) with your
rules added to the built-in ones; `pkg/custos/ast` describes the syntax tree
they inspect.

```go
package main

import (
	"github.com/janalis/custos/pkg/custos"
	"github.com/janalis/custos/pkg/custos/ast"
)

var noEval = custos.MustRule(custos.Info{
	ID:       "NoEval",
	Severity: custos.SeverityError,
	Summary:  "eval() runs arbitrary code.",
}, []ast.NodeKind{ast.KEval}, func(ctx *custos.Context, n ast.Node) {
	ctx.ReportNode(n, "Remove eval()")
})

func main() {
	noDump, err := custos.Pattern(custos.PatternDefinition{
		ID:      "NoDump",
		Match:   custos.Patterns{`\dump(...)`},
		Message: "Remove the debug call",
	})
	if err != nil {
		panic(err)
	}
	custos.Main("1.0.0", noEval, noDump)
}
```

`custos.NewRule` calls the check function for every node of the listed
kinds. `custos.json` configures such rules like the others, by their ID. The context gives the file's source (`ctx.Text`), resolved names
(`ctx.GlobalFunctionName`, `ctx.Names()`), inferred types (`ctx.TypeOf`) and
reporting with quick-fixes (`ctx.ReportNode`). The repository's
`examples/custom-binary` is a complete program.

Such a program is not a custos release: it never offers to update itself,
and its `self-update` command refuses to replace it with the stock binary.
The Go API exposes the engine's own types and may change between releases
until custos 1.0.
