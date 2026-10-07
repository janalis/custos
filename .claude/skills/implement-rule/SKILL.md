---
name: implement-rule
description: Implement one or more custos rules in Go from their clean-room spec (specs/<RuleID>.md), with own fixtures and a green conformance run. Use when asked to implement/port a rule whose spec exists.
---

# implement-rule

## Hard rule
Work **only** from `specs/<ID>.md`, `internal/meta/rules.json` and custos'
own code. Do **not** open any file under the EA checkout
(`~/Sites/phpinspectionsea`) — not the Java sources, not the fixtures. The
conformance runner reads EA fixtures for you and only reports
missing/unexpected ranges and fix diffs.

## Prerequisites
Phases 2–3 (parser, analysis engine, fixer) must exist; semantic rules
(`kind: semantic`) also need Phase 6 for their `needs`. If missing, stop and
report which prerequisite blocks the rule.

## Steps
1. Read the spec and the rule's facts (`internal/meta/rules.json`).
2. Create `internal/rules/<group-slug>/<rule_snake>.go` implementing the
   `analysis.Rule` interface: `Meta()` from the catalogue, `Kinds()` listing
   only the node kinds it needs, `Check(ctx, node)` following D/E items in
   order. Fixes as text edits per F items. Register it with
   `func init() { register(myRule{}) }` in the rule's own file (never edit a
   shared list).
3. Reuse helpers in `internal/analysis/util` before writing new ones. New
   generic helpers go in a NEW file `internal/analysis/util/<topic>.go`
   (+ `_test.go`); rule-specific helpers stay in the rule file. Other agents
   work in parallel: never modify files you did not create in this task
   (except to fix a genuine bug, which you must report). If the build breaks
   because of someone else's in-progress file, wait a minute and retry.
4. Write own fixtures in `testdata/rules/<ID>/`: `basic.php` (positives),
   `false-positives.php`, `basic.fixed.php` if fixable, `<name>.json` for
   options/PHP version. Use the spec's examples as a starting point.
5. `make fixtures RULE=<ID>` then `make conformance RULE=<ID>`. On EA
   mismatches, re-read the spec; if the spec is wrong or incomplete, report it
   (spec fixes go through `spec-rule`) instead of guessing from EA files.
6. If a rule touches a hot path, add/extend a benchmark.
7. `make verify && make rules-doc`.

## Output
Per rule: files added, conformance result (pass / fail with reasons /
documented divergence), spec gaps found.

## Building

Build the CLI with `make build` (→ `bin/custos`, run as `./bin/custos`).
Never run a bare `go build ./cmd/custos` — it writes a stray binary into the
current directory. To only check compilation use `go build ./...`.

## API cheat sheet

- AST: `internal/syntax` (`ast.go` node types, `walk.go` `Children`/`Inspect`,
  every node has `Span()`, `Parent()`, `Kind()`; `Paren` nodes are kept;
  `ExprStmt`, `Block{Alt}`, `Nop` for empty `;`). Tokens (incl. comments) in
  `File.Tokens`; find tokens inside a span with a binary search on `Start`.
- Rule: implement `analysis.Rule` — `ID() string`, `Kinds() []syntax.NodeKind`,
  `Check(ctx *analysis.Context, n syntax.Node)`; whole-file rules also
  implement `analysis.FileRule`. Register via `init()` + `register(...)`.
  Conventions: see `internal/rules/codestyle/unnecessary_semicolon.go`.
- Names: `ctx.Names()` (namespace/use resolution: `Class`, `Function`,
  `Const`), `ctx.IsGlobalFunctionCall(call, "strlen")`, `ctx.FunctionName(call)`
  (lower-case resolved name). No type inference / symbol index yet — rules
  needing them are not assigned yet.
- Recovery nodes: error recovery may create zero-width nodes (`Nop`,
  `BadExpr`); ignore nodes whose span is empty.
- Context: `ctx.PHP` (target version, `phpver.PHP71` …), `ctx.ComparisonStyle`,
  `ctx.Text(n)`, `ctx.Report(span, msg, fixes...)`, `ctx.ReportNode`,
  `ctx.ReportSeverity`, options `ctx.Bool/Int/String/List("OPTION_NAME")`
  (EA list-option calls arrive as `ctx.List("@calls")`), `ctx.IsTestFile()`,
  `ctx.Memo(key, fn)` to compute per-file derived data (parsed option lists,
  lookup tables) once instead of on every node.
- Fix: `analysis.Fix{Title: "...", Edits: func() []analysis.TextEdit {...}}`
  — edits are byte-range replacements on the original source; build them lazily.
- Fixtures: `testdata/rules/<ID>/*.php` use `<warning descr="msg">code</warning>`
  (`<error>`, `<weak_warning>` = info); optional `<name>.fixed.php` and
  `<name>.json` (`{"php": "7.4", "comparisonStyle": "yoda", "options": {"OPT": "true"}}`).
  Own fixtures default to PHP 8.4; messages are compared exactly.
- EA conformance parses with PHP 8.5 permissive mode and gates rules at the
  case level (default 5.6 when the EA test sets none).

## Semantic API (Phase 6)

- `ctx.Index()` — symbols of this file layered over the project index and the
  embedded PHP stubs (`internal/index`): `Class(fqn, ver)`, `Function`,
  `Constant`, `ResolveFunction(fqn, fallback, ver)`, `Ancestors`,
  `ParentChain`, `IsSubtype`, `FindMethod/FindProperty/FindConst` (inheritance
  aware), `Children`. Use `ctx.PHP` as the version argument.
- `ctx.Types()` / `ctx.TypeOf(expr)` — type inference (`internal/infer`,
  types in `internal/types`): `types.Type` is a set of atoms (`int`, `string`,
  `null`, `\Foo\Bar`, `\Foo[]`…); `IsUnknown()`, `Has`, `OnlyOf`, `Classes()`,
  `Elem()`, `IsArrayLike()`. Unknown means "no information": rules must stay
  silent on unknown unless the spec says otherwise. Helpers on the env:
  `ResolveFunction(call)`, `ClassFQN(classLike)`, `infer.EnclosingClass(n)`.
- `index.DocComment(file, node)` + `internal/phpdoc` (`Parse`, `Tag`, `Params`,
  `ReturnType`, `VarType`) for doc tags.
- Rules needing symbols from OTHER files must implement the marker
  `Semantic()` (interface `analysis.SemanticRule`) so the CLI builds the
  project index first. EA conformance runs single-file (file + stubs).
- The inference engine is young: when a rule needs a missing capability
  (a builtin return override, a construct not inferred), add it to
  `internal/infer` with a test in `infer_test.go` — small, focused edits only,
  since other agents may edit it too; run `go test ./internal/infer` after.
- Flow-type needs (reads/writes of variables in order, "used later") are
  implemented per rule on the AST unless a shared helper already exists in
  `internal/analysis/util`.
