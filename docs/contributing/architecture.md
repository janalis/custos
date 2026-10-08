# Architecture

custos is a single Go module (`custos`) with no runtime dependency outside
the standard library. A file goes through this pipeline:

```text
discover files ─► lex + parse ─► names ─► (lazy) index · types ─► one walk, rules by node kind ─► findings
   runner          syntax        names      index/stubs/infer       analysis + rules              report · fix · lsp
```

## Packages

| Package | Role |
|---|---|
| `cmd/custos` | The CLI: `analyse`, `fix`, `rules`, `explain`, `lsp`, `version`. |
| `internal/syntax` | Lexer, version-aware parser with error recovery, AST, walker and line index. Nodes are arena-allocated; positions are `uint32` byte offsets. |
| `internal/phpver` | The PHP version model (5.3 to 8.5). |
| `internal/names` | Namespace and `use` resolution. |
| `internal/phpdoc`, `internal/types` | PHPDoc tags and types (`@template`, aliases), and the type model (sets of atoms). |
| `internal/index` | Project symbol index: classes, members, functions, constants and inheritance. |
| `internal/stubs` | Embedded index of PHP builtins, generated from JetBrains phpstorm-stubs (`make stubs`). |
| `internal/infer` | Expression type inference and narrowing. |
| `internal/analysis` | The `Rule` interface, the engine, the per-file `Context` and suppressions. `analysis/util` holds shared AST helpers (calls, values, variable use, reachability, hierarchy…). |
| `internal/rules/<group>` | One file per rule, registered from its own `init()`. |
| `internal/fix` | Applies text edits and runs the fix loop. |
| `internal/runner` | File discovery, parallel analysis, project index build. |
| `internal/report` | Output formats: text, JSON, Checkstyle, GitHub, SARIF. |
| `internal/lsp` | The language server. |
| `internal/config` | `custos.json` and `composer.json` reading. |
| `internal/meta` | Rule facts (`rules.json`) and descriptions generated from the specs. |
| `internal/conformance` | Fixture markup parser and the own-fixture and upstream conformance runners. |
| `tools/` | Generators and checks: `extract`, `rulesdoc`, `rulesref`, `genexplain`, `genkinds`, `genstubs`, `cleanroom`, `covercheck`, `relprep`. |

## The engine

- Each rule declares the node kinds it wants (`Kinds()`). The engine builds
  a dispatch table and walks each file's syntax tree **once**, calling only
  the rules registered for each node kind.
- Semantic data (name resolution, the project index, types) is computed
  **lazily**, the first time a rule asks the `Context` for it, so syntax-only
  rules stay cheap.
- Files are analysed in parallel. The project index is built first, so that
  rules can resolve classes and members declared in other files.
- Rules report a span, a message and optional fixes. A fix is a function
  returning text edits on exact byte ranges, so a fix is only built when it
  is needed (`custos fix`, or an LSP code action).

```go
type Rule interface {
	ID() string                       // e.g. "UnnecessarySemicolon"
	Kinds() []syntax.NodeKind         // node kinds to visit
	Check(ctx *Context, n syntax.Node)
}
```

## Fixing

`custos fix` runs the analysis, applies the non-overlapping edits of every
finding, and runs again until nothing changes (at most 10 rounds), since a
fix can reveal or remove other findings. `make fixcheck` applies every fix
offered on a corpus of real projects and checks that the output still
parses.

## Language server

`custos lsp` keeps the project index in memory, re-analyses open documents
as they change, and updates the index when a file is saved. Quick-fixes are
offered as code actions and resolved lazily when the client supports it.

## Generated files

Run `make rules-doc` after changing a spec or a rule's fixtures. It rewrites:

- `internal/meta/descriptions.json` (`custos explain`), from the specs'
  *Summary* and *Options* sections;
- `docs/rules/**` and `docs/.vitepress/rules-sidebar.json` (this site's rule
  reference);
- `docs/internals/rules.md` (per-rule status).

CI checks that the committed files are up to date.
