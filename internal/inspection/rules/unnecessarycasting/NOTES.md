# UnnecessaryCasting implementation notes

## T-rules typer

The "Expression type rules (T-rules)" of the spec are implemented by
`infer.TRules` (`internal/semantic/infer/trules.go`), shared with
CallableParameterUseCaseInTypeContext. It is layered over `infer.Env`:
expressions the T-rules do not cover fall back to `Env.TypeOf`.

- Results drop unknown parts: `types.Unknown` (no atoms) means "empty".
  Atoms are raw: class names keep their leading backslash and `T[]` forms
  are kept.
- It stays separate from `Env` on purpose. Its rules come from the specs
  (partial sets, spec-defined arithmetic and overrides, an optional
  `SpecOnly` mode) and differ from the engine's typing; folding it into `Env`
  would change the engine's types or the two rules' findings.
- Outside `SpecOnly`, a variable is typed from the definitions reaching it:
  an unconditional reassignment of unknown type hides earlier assignments
  instead of leaving their types in the union. Method and static call
  results also pick up `null` from the engine's null-safe chain facts.

## Typer options

Set them before the first `TypeOf` call.

- `SpecOnly` evaluates the T-rules as literally written, without the
  engine's enrichments. Only CallableParameterUseCaseInTypeContext uses it
  (its spec prefers "unknown" when in doubt):
  - a variable's type comes only from its parameter types and plain `=`
    assignments (`=&` included, typed by T-rules); `foreach` bindings and
    compound assignments neither contribute nor hide earlier assignments;
  - a method call's receiver is typed by T-rules only, and a receiver class
    lacking the method contributes nothing (the others still count);
  - the R-function subject argument is typed by T-rules;
  - arithmetic normalises `T[]` to array and sees a parenthesised numeric
    literal as a numeric literal;
  - `c ?: b` needs a non-empty type for `c` minus null;
  - `self`/`static`/`parent` member contracts are kept as written for the
    consuming rule to translate; ordinary mode binds their class contexts.
- `DivisionIntOrFloat` types `int / int` as `int|float` instead of the spec
  heuristic's `int` (this rule: "Integer division" divergence).
- `SoundArithmetic` types `+ - * / **` by PHP's result rules instead of the
  spec heuristic: an unknown operand gives unknown, a float operand gives
  float, int with int gives int (`/` and `**` with a non-literal exponent may
  give a float), and any other operand (numeric strings, null, bool) gives
  `int|float`. Both rules set it.

## PHPDoc-free proof

The cast is reported only when the type also holds without user PHPDoc: the
check is repeated with a typer over `ctx.Types().Native()`, an `Env` that
ignores project docs (parameter, property and return docs, inline `@var`,
templates, conditional returns, assertions, `@param-out`) and index-time
inferred types, while builtin stub docs still count.
