# Suppressing findings

## In the code

Put a comment right before the statement or declaration:

```php
// @custos-ignore OneTimeUseVariables
$total = compute();

/** @noinspection OneTimeUseVariablesInspection */
function legacy() { /* … */ }
```

- `@noinspection <ID>Inspection` is the PhpStorm syntax, so comments written
  for Php Inspections (EA Extended) keep working. `@custos-ignore <ID>` is the
  custos spelling. Both accept either form of the name.
- Several names can be listed: `// @custos-ignore UnusedGotoLabel, UnnecessarySemicolon`.
- `ALL` suppresses every rule: `// @custos-ignore ALL`.
- Put the comment before the **first statement of a file** to suppress the
  rules for the whole file.

In an editor, the language server offers a *Suppress … for this
statement* action that writes the comment for you. See [Editors](./editors).

## Turning rules off

To stop a rule everywhere, disable it in `custos.json`:

```json
{ "rules": { "MultipleReturnStatements": { "enabled": false } } }
```

## Baseline

On a legacy code base, a baseline records today's findings so that only new
ones are reported:

```sh
custos analyse --generate-baseline custos-baseline.json
```

Commit the file and reference it from `custos.json` (or pass `--baseline`):

```json
{ "baseline": "custos-baseline.json" }
```

Baseline entries match on file, rule, message and the text of the flagged
line, not on line numbers, so they survive code moving around. Regenerate the
baseline after fixing old findings to keep it small.
