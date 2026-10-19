---
name: implement-rule
description: Implement one or more custos rules in Go from their clean-room spec (specs/<RuleID>.md), with own fixtures and a green conformance run. Use when asked to implement/port a rule whose spec exists.
---

# implement-rule

## Hard rule (clean room)

Work **only** from `specs/<ID>.md`, the rule's facts
(`internal/inspection/meta/rules.json` or `native-rules.json`) and custos'
own code. Do **not** open any file under the EA checkout
(`~/Sites/phpinspectionsea`): not the Java sources, not the fixtures. The
conformance runner reads EA fixtures for you and only reports
missing/unexpected ranges and fix diffs. Process:
`docs/contributing/clean-room.md`.

## Steps

1. Read the spec and the rule's facts. If the spec is missing or does not
   describe a behavior you need, stop and report the gap.
2. Create `internal/inspection/rules/<lowercase-ID>/rule.go` with the package
   comment (`// Package <id> implements the <RuleID> inspection.`) and
   `New() analysis.Rule`. Put the private implementation type there or, for
   larger rules, in `<rule_snake>.go`. Implement `ID()`, `Kinds()` and
   `Check` following the spec's D/E/F items.
3. Add exactly one constructor to `internal/inspection/catalogue/catalogue.go`,
   preserving group execution order and ID sorting within each group. There
   is no `init()` registration.
4. Keep rule-specific helpers local. Reuse `astquery`, `flowquery` or
   `semanticquery` (lexical, local flow, resolved behavior); new shared
   helpers need focused tests. Run `make architecture`.
5. Write own fixtures in `internal/inspection/rules/<lowercase-ID>/testdata/`
   starting from the spec's examples; format in
   `docs/contributing/fixtures.md`. They must cover every statement of the
   rule (`make coverage` lists uncovered blocks): remove unreachable code;
   cover parser-recovery guards with broken PHP (`broken*.php`).
6. `make fixtures RULE=<ID>`, then for EA-modeled rules
   `make conformance RULE=<ID>`. On mismatches, re-read the spec; if it is
   wrong or incomplete, report it (spec fixes go through `spec-rule`) instead
   of guessing from EA files. Intentional divergences go in the spec and
   `internal/testing/conformance/testdata/ea-divergences.json`.
7. If the rule touches a hot path, add or extend a benchmark.
8. If something is worth keeping that the spec must not say (non-obvious
   algorithm, budgets/limits, reliance on shared helpers, known gaps), write
   `internal/inspection/rules/<lowercase-ID>/NOTES.md`, titled
   `# <RuleID> implementation notes`, with free `##` sections in your own
   words. Skip it when there is nothing to add.
9. `make rules-doc`, then `make verify`.

## Output

Per rule: files added, conformance result (pass / fail with reasons /
documented divergence), spec gaps found.

## Building

Build the CLI with `make build` (→ `bin/custos`). Never run a bare
`go build ./cmd/custos`; it writes a stray binary into the current directory.
To only check compilation use `go build ./...`.

## API cheat sheet

Grep the packages before writing a private helper; names below exist today.

- Rule contract (`internal/inspection/analysis`): `Rule` (`ID`, `Kinds`,
  `Check(ctx, n)`); `FileRule` (`CheckFile(ctx)`) for whole-file rules.
  Marker methods: `Semantic()` (`SemanticRule`) when the rule needs
  declarations from other files, so the project index is built first;
  `Flow()` (`FlowRule`) when it needs project-local callable summaries.
  Example: `internal/inspection/rules/unnecessarysemicolon/`.
- Context: `ctx.File`, `ctx.Src`, `ctx.PHP` (target version),
  `ctx.ComparisonStyle`, `ctx.Text(n)`, `ctx.Report(span, msg, fixes...)`,
  `ctx.ReportNode`, `ctx.ReportSeverity`, options
  `ctx.Bool/Int/String/List("OPTION")` (list-option calls arrive as
  `ctx.List("@calls")`), `ctx.IsTestFile()`, `ctx.Memo(key, fn)` for
  per-file derived data, `ctx.Gap(reason, span)` when a bounded query gives
  up.
- Semantics (lazy): `ctx.Names()` (namespace/use resolution),
  `ctx.IsGlobalFunctionCall(call, "strlen")`, `ctx.FunctionName(call)`,
  `ctx.Index()` (`Class`, `Function`, `Constant`, `ResolveFunction`,
  `Ancestors`, `IsSubtype`, `FindMethod/FindProperty/FindConst`),
  `ctx.Types()` / `ctx.TypeOf(expr)` (unknown means no information: stay
  silent), `ctx.Flow()` (`internal/semantic/flow` values and state).
  PHPDoc: `index.DocComment(file, node)` + `internal/php/phpdoc`.
- AST (`internal/php/syntax`): `Children`, `Inspect`, `UnwrapParens`,
  `IsFuncLike`, `EnclosingFuncLike`, `EnclosingClass`, `FuncLikeParams`,
  `FuncLikeBody`, `IsNullConst`. Error recovery may create zero-width
  nodes; ignore nodes whose span is empty.
- `astquery` (lexical): `ParentSkipParens`, `CallArgValues`, `ArgValue`,
  `CallLastName`, `LastNamePart`, `BoolConst`, `IsStringLiteral`,
  `IsNumberLiteral`, `StringLiteralValue`, `QuotedStringContent`,
  `IsLogicalOperand`, `MentionsVariable`, `Equivalent`, `TokenBefore`,
  `TokenAfter`, `IndentBefore`, `LineIndent`, `PrevStmt`, `NextStmt`,
  `IsTestClassFQN`, `ReplaceFix`.
- `flowquery` (local flow): `Reachable`, `ReachingAssignments`,
  `VarAccesses`, `MayHaveSideEffects`, `PossibleValues[Known]`,
  `PossibleValuesReaching`, `PossibleValuesComplete`; see its package
  comment for the completeness contract.
- `semanticquery` (resolved): `ResolvesToGlobalFunction`, `QualifiedBuiltin`,
  `NativeBuiltin`, `CallArgument`, `ArgBindsByRef`, `ResolveCallee`,
  `DiscoverValues[Known]`, `ClassDecl`, `MethodDecl`, `MethodInChain`,
  `DocHasTag`, `DocHasAnnotation`, `GlobalConstName`, `InTestContext`,
  `NativeCallHistory` (`CallHistory.Certain()`).
- Fix: `diagnostic.Fix{Title: "...", Edits: func() []diagnostic.TextEdit {...}}`;
  edits are byte-range replacements on the original source, built lazily.
- Missing inference capability: add it to `internal/semantic/infer` with a
  focused test; keep edits small and run `go test ./internal/semantic/infer`.
