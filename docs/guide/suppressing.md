# Suppressing findings

## In the code

Put a comment right before the statement or declaration:

```php
// @custos-ignore OneTimeUseVariables
$total = compute();

/** @noinspection OneTimeUseVariablesInspection */
function legacy() { /* … */ }
```

- `@custos-ignore` and `@noinspection` work the same way and accept a rule ID
  or, for EA rules, the PhpStorm name (`<ID>Inspection`). Comments written for
  Php Inspections (EA Extended) in PhpStorm therefore keep working. Native
  rules have no PhpStorm name: use their rule ID.
- Several names can be listed: `// @custos-ignore UnusedGotoLabel, UnnecessarySemicolon`.
- `ALL` suppresses every rule: `// @custos-ignore ALL`.
- Put the comment before the **first statement of a file** to suppress the
  rules for the whole file.
- A comment right before a parameter, a property hook, a call argument or an
  array item suppresses the rules for that element only, even inside a
  statement that spans several lines:

  ```php
  return new SettingsRow(
      'Layout',
      // @noinspection JsonEncodingApiUsageInspection -- the value is display text
      value: (string) \json_encode($layout),
      hint: \json_encode($hint), // still reported
  );
  ```

In an editor, the language server offers a *Suppress … for this
statement* action that writes the comment for you. See [Editors](./editors).

## Turning rules off

To stop a rule everywhere, disable it in `custos.json`:

```json
{ "rules": { "StaticClosureCanBeUsed": { "enabled": false } } }
```

## Baseline

On a legacy code base, a baseline records today's findings so that only new
ones are reported:

```sh
custos analyse --generate-baseline custos-baseline.json
```

Commit the file and reference it from `custos.json`, relative to the project
root:

```json
{ "baseline": "custos-baseline.json" }
```

`analyse` then hides the recorded findings. A baseline named in `custos.json`
that does not exist yet suppresses nothing. You can also pass
`--baseline FILE` (relative to the current directory), which must exist.
`--generate-baseline` writes its file relative to the current directory too.
`--stats` prints how many findings the baseline suppressed.

Baseline entries match on file, rule, message and the text of the flagged
line, not on line numbers, so they survive code moving around. Regenerate the
baseline after fixing old findings to keep it small.
