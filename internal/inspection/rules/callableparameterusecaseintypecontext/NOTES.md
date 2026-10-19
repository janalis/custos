# CallableParameterUseCaseInTypeContext implementation notes

## Expression typing

Expression types (the spec's T-rules, defined in
`specs/UnnecessaryCasting.md`) come from the shared typer `infer.TRules`,
also used by UnnecessaryCasting; see `../unnecessarycasting/NOTES.md`.
This rule configures it differently:

- `SpecOnly`: the T-rules as written, without the general engine's
  enrichments, so doubtful cases stay unknown. Variables are typed only from
  parameter types and plain `=`/`=&` assignments (foreach bindings and
  compound assignments neither contribute nor hide earlier assignments); a
  method call's receiver is typed by T-rules only, and a receiver class
  lacking the method contributes nothing; `c ?: b` needs a non-empty type for
  `c` minus null. The `SpecOnly` field comment in `infer/trules.go` lists the
  full set.
- `SoundArithmetic`: arithmetic follows PHP's actual result rules (see the
  spec's "Sound arithmetic" divergence), so `time() + $untyped` is not known
  to be a float.
- `DivisionIntOrFloat` is left off.
