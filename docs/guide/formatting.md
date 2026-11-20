# Formatting

`custos format` lays out PHP code: indentation, spaces between tokens, line
breaks, braces, blank lines, and the token spellings a style guide fixes,
such as `array()` versus `[]`, keyword case or trailing commas. It follows a
preset ([PER Coding Style](https://www.php-fig.org/per/coding-style/) unless
you choose another) that your `.editorconfig`, `custos.json` and the command
line can adjust.

The layout is *hybrid*: custos keeps the line breaks you wrote where the
style allows them, and only splits a list (arguments, parameters, array
items, closure `use`, `match` arms) when it is wider than the line length or
already split over several lines, putting one item per line. A list you
wrote on several lines stays on several lines unless you turn on
`join_fitting_lists`.

```diff
-$total = $calculator->invoiceTotal($order->lines(), $customer->discountRate(), $taxes->forCountry($country), $shippingCost);
+$total = $calculator->invoiceTotal(
+    $order->lines(),
+    $customer->discountRate(),
+    $taxes->forCountry($country),
+    $shippingCost,
+);
```

## Quick start

```sh
custos format                   # format the project in place
custos format src/Invoice.php   # one file or directory
custos format --check           # change nothing; exit 1 when a file would change
custos format --diff --dry-run  # show what would change as a unified diff
custos format --show-config     # every setting, its value and where it comes from
```

`custos format` checks the same files as `custos analyse`: the given paths,
else `paths` from `custos.json`, else the project root, minus the
[excluded directories](./configuration#excluded-directories) and the files
`format.exclude` lists. Use `--check` in CI so that a pull request fails when
its code is not formatted (see [Check and fix](#check-and-fix)).

For editors and scripts, format standard input to standard output:

```sh
custos format --stdin --stdin-path src/Invoice.php < src/Invoice.php
custos format - < src/Invoice.php
```

`--stdin-path` names the file the input stands for, so that its
`custos.json`, `.editorconfig` and `format.exclude` apply; it is not read.

## Presets

A preset is a complete set of setting values that follows a published style
guide. Each one builds on another and changes some of its values:

| Preset | Follows | Builds on |
|---|---|---|
| `psr1` | [PSR-1](https://www.php-fig.org/psr/psr-1/): only removes a byte order mark; the rest of the layout is kept as written. | — |
| `psr12` | [PSR-12](https://www.php-fig.org/psr/psr-12/) | `psr1` |
| `per` | [PER Coding Style](https://www.php-fig.org/per/coding-style/), the successor of PSR-12. **The default.** | `psr12` |
| `symfony` | The [Symfony coding standards](https://symfony.com/doc/current/contributing/code/standards.html) | `per` |
| `laravel` | The Laravel style, as [Laravel Pint](https://laravel.com/docs/pint) applies it by default | `per` |
| `drupal` | The [Drupal coding standards](https://project.pages.drupalcode.org/coding_standards/php/coding/): two-space indentation, braces on the same line, 80 columns. | `per` |

[Format settings](./format-settings) lists every setting with the value
each preset gives it. Choose a preset in `custos.json` or with `--preset`:

```json
{
  "format": { "preset": "symfony" }
}
```

When neither names one (the default `"preset": "auto"`), custos takes the
style of the project's other formatter (see
[Other tools' configuration](#other-tools-configuration)), else `per`.

### Compliance

Each preset is checked against the tool that defines its style guide on
real projects: PHP_CodeSniffer for `psr12` and `drupal`, PHP-CS-Fixer for
`per` and `symfony`, Laravel Pint for `laravel`. A preset lays code out; it
leaves alone what those tools also check beyond layout, such as the content
of docblocks, names and imports, and the risky rewrites custos applies only
on request. The rules each preset deliberately does not satisfy are listed,
with the reason, in
[`divergences.json`](https://github.com/janalis/custos/blob/main/internal/format/testdata/divergences.json).

## How settings resolve

Each setting takes the value of the last of these sources that sets it:

1. the preset;
2. the configuration of the project's other formatter, only when the preset
   is `auto` (in both `custos.json` and on the command line);
3. the `.editorconfig` files that apply to the file;
4. the `format` section of `custos.json`;
5. `--set id=value` on the command line.

`--show-config` prints the value of every setting for the first path (a
directory stands for a PHP file in it) and the source of each one. With
`"format": {"settings": {"line_length": 100}}` in `custos.json`, an
`.editorconfig` setting `indent_size = 4` and `max_line_length = 110`, and
`--set wrap=off`, an excerpt reads:

```text
$ custos format --show-config --set wrap=off
preset                            per                                                                        default
indent_style                      "space"                                                                    preset psr12
indent_size                       4                                                                          .editorconfig
tab_width                         4                                                                          .editorconfig
…
line_length                       100                                                                        custos.json
wrap                              "off"                                                                      command line
report_long_lines                 false                                                                      default
…
empty_body                        "inline"                                                                   preset per
```

The first line gives the preset and how it was chosen. A value marked
`default` is the neutral value no preset changes, which usually leaves the
code as written.

### EditorConfig

custos reads the `.editorconfig` files from the file's directory up to the
one marked `root = true`, as editors do, and maps these properties:

| Property | Setting |
|---|---|
| `indent_style` | `indent_style` |
| `indent_size` (`tab` takes `tab_width`) | `indent_size` |
| `tab_width` | `tab_width` |
| `end_of_line` | `line_ending` |
| `insert_final_newline` | `final_newline` |
| `trim_trailing_whitespace` | `trim_trailing_whitespace` |
| `max_line_length` (`off` leaves it unset) | `line_length` |

Other properties, and values custos does not understand, are ignored. Set
`"editorconfig": false` in the `format` section to ignore `.editorconfig`
files.

### Other tools' configuration

With the `auto` preset, custos reads the first of these at the project root
and starts from the preset closest to its style. It reads the files, never
runs them, and imports only what maps unambiguously onto its own settings;
the tools' other rules are left to the preset.

| File | Preset | Also imported |
|---|---|---|
| `pint.json` | its `preset`: `laravel`, `per`, `psr12`, `symfony`, or `empty` as `psr1`; `laravel` when it names none | — |
| `.php-cs-fixer.php`, `.php-cs-fixer.dist.php` | the last layout rule set enabled: `@PSR1` as `psr1`, `@PSR2` and `@PSR12` as `psr12`, `@PER-CS…` as `per`, `@Symfony` and `@PhpCsFixer` as `symfony`; `psr12` without one | `setIndent()` (`indent_style`, `indent_size`) and `setLineEnding()` (`line_ending`), written as string literals |
| `mago.toml`, `mago.dist.toml` | `[formatter]` `preset`: `default` as `per`, `psr-12` as `psr12`, `laravel` and `pint` as `laravel`, `drupal`; `per` without one | `print-width` (`line_length`), `tab-width` (`indent_size`, `tab_width`), `use-tabs` (`indent_style`) |

`--show-config` names the file the preset came from. Risky rule sets such as
`@Symfony:risky` are not imported: turn [risky settings](#risky-settings)
on in `custos.json`.

## Risky settings

Some settings change more than the layout: they rewrite code in ways that
are equivalent only under assumptions, such as no other code relying on a
removed import or a string's quoting. They are off unless you turn them on:

- `yoda`: the order of comparison operands, constant first
  (`null === $x`, value `yoda`) or last (`regular`);
- `single_quote`: write simple double-quoted strings in single quotes;
- `remove_unused_imports`: remove `use` imports the file never mentions;
- `declare_strict_types`: add `declare(strict_types=1);`, which changes how
  PHP checks scalar types at run time;
- `increment_style`: write standalone increments as `++$i` or `$i++`;
- `heredoc_to_nowdoc`: write heredocs without interpolation as nowdocs.

There are three ways to turn them on:

- `--allow-risky`, or `"allowRisky": true` in the `format` section, applies
  the risky values of the preset: `symfony` and `laravel` have some (the
  [Risky](./format-settings#risky) table shows them), the other presets
  none.
- `"risky": ["yoda", "increment_style"]` in the `format` section turns on
  the settings listed, whatever the preset: a `true`/`false` setting
  becomes `true`, another takes the preset's risky value, else its first
  value (`yoda`, `pre`).
- A value in `settings` or `--set`, such as `--set declare_strict_types=true`,
  always applies.

## Ignoring code

Comments keep code as written:

| Comment | Effect |
|---|---|
| `@custos-format-ignore` | The whole file. |
| `@custos-format-ignore-start` … `@custos-format-ignore-end` | The code between the two comments; without an end, up to the end of the file. |
| `@custos-format-ignore-next` | The next statement or class member. |

```php
// @custos-format-ignore-next
const MATRIX = [
    1, 0, 0,
    0, 1, 0,
    0, 0, 1,
];
```

custos also honors `@formatter:off` and `@formatter:on` (PhpStorm), and
Mago's `@mago-format-ignore`, `@mago-format-ignore-start`,
`@mago-format-ignore-end` and `@mago-format-ignore-next`, so code ignored
for those tools stays ignored.

## What is never changed

- The contents of strings, heredocs and nowdocs, including their trailing
  spaces and indentation. (Risky settings may change how a string or heredoc
  is quoted, never what it holds.)
- Inline HTML. A file that mixes PHP and HTML (a template) keeps its line
  breaks and indentation; only spaces and spellings inside the PHP code
  change.
- The data after `__halt_compiler();`.
- Code inside [ignore comments](#ignoring-code).
- Files with syntax errors: they are skipped with a warning.

Before writing a file, custos checks that the result holds the same code:
the same tokens apart from the intended rewrites, and no new syntax error.
If the check fails, the file is left as written and the run fails; please
[report it](https://github.com/janalis/custos/issues/new/choose) with the
file.

## Check and fix

Without `--check` or `--dry-run`, `custos format` rewrites the files that
change and prints, on stderr, a summary such as
`3 file(s) reformatted, 120 unchanged, 1 skipped`.

`--check` (or `--dry-run`, which is the same for files) writes nothing and
lists each change with its position, the setting behind it and a message,
then exits 1 when a file would change:

```text
src/Invoice.php
  3:8  format.class_brace  Move the class opening brace
  4:17  format.spacing  Fix the spaces between tokens
1 file(s) would be reformatted, 0 unchanged, 0 skipped
```

`--report` prints that listing as `json`, `checkstyle`, `github` or
`sarif`, the formats of [`analyse`](./cli#analyse); changes are `info` items
and skipped files `error` items (rule `format.skipped`). `--diff` prints a
unified diff instead of the listing: with `--check` or `--dry-run` nothing
is written, without them the files are written as well.

With `report_long_lines` on, lines that stay wider than `line_length` after
formatting (formatting cannot shorten every line) are listed too, and fail a
check.

| Exit code | Meaning |
|---|---|
| `0` | Success: files formatted, or nothing to change with `--check`. Files with syntax errors only warn. |
| `1` | With `--check` or `--dry-run`: a file would change, or a long line is reported. |
| `2` | Usage or configuration error, or a file could not be formatted (read or write error, or a result the safety check refused). |

Files already known to be formatted with the same settings are skipped on
the next run: custos keeps their fingerprints in its
[cache](./cli#paths). `--no-cache` neither reads nor writes it.

## Editors and agents

- The [language server](./editors) formats a document or a selection
  (`textDocument/formatting`, `textDocument/rangeFormatting`) with the
  project's configuration; it ignores the editor's tab size, so the editor
  writes what `custos format` does.
- The [MCP server](./mcp#formatting) offers `custos_format`, which checks,
  previews or applies the formatter.
- The [Formatting](/rules/code-style/Formatting) rule, off by default, brings
  formatting into `custos analyse`: each change becomes an `info` finding with
  a quick-fix, so it shows in editors and can be recorded in a
  [baseline](./suppressing#baseline). Enable it with
  `"rules": {"Formatting": {"enabled": true}}`.

## Migrating from other formatters

Start with `"preset": "auto"` (the default): custos picks the preset from
your `pint.json`, PHP-CS-Fixer or Mago configuration. Compare the result with
`custos format --check`, then set what differs in `format.settings` and
remove the old configuration.

Common PHP-CS-Fixer rules (also those of a `pint.json` `rules` object) and
the custos settings that cover them:

| PHP-CS-Fixer rule | custos setting |
|---|---|
| `indentation_type`, `setIndent()` | `indent_style`, `indent_size` |
| `line_ending`, `setLineEnding()` | `line_ending` |
| `single_blank_line_at_eof` | `final_newline` |
| `no_trailing_whitespace` | `trim_trailing_whitespace` |
| `no_closing_tag` | `closing_tag` |
| `blank_line_after_opening_tag`, `linebreak_after_opening_tag` | `blank_line_after_open_tag`, `open_tag_own_line` |
| `no_extra_blank_lines` | `max_blank_lines` |
| `braces_position` | `class_brace`, `function_brace`, `closure_brace`, `anonymous_class_brace`, `control_brace` |
| `control_structure_continuation_position` | `else_position`, `catch_position` |
| `blank_line_after_namespace`, `single_line_after_imports` | `blank_line_after_namespace`, `blank_line_after_imports` |
| `no_blank_lines_after_class_opening` | `class_open_blank_line` |
| `class_attributes_separation` | `method_separation`, `property_separation`, `constant_separation` |
| `blank_line_before_statement` | `blank_line_before` |
| `concat_space` | `concat_space` |
| `binary_operator_spaces` | `binary_operator_space` (`one`; `at_least_one` keeps alignment) |
| `cast_spaces` | `cast_space` |
| `not_operator_with_successor_space` | `not_operator_space` |
| `function_declaration` | `arrow_function_space`, `closure_space` |
| `lowercase_keywords`, `constant_case`, `magic_constant_casing` | `keyword_case`, `constant_case`, `magic_constant_case` |
| `lowercase_cast`, `short_scalar_cast` | `short_cast` |
| `trailing_comma_in_multiline`, `no_trailing_comma_in_singleline` | `trailing_comma`, `single_line_trailing_comma` |
| `array_syntax`, `list_syntax` | `array_syntax`, `list_syntax` |
| `elseif`, `standardize_not_equals` | `elseif`, `standardize_not_equals` |
| `new_with_parentheses` | `new_parentheses` |
| `single_line_comment_style` | `hash_comment` |
| `operator_linebreak` | `operator_linebreak` |
| `ordered_imports` | `import_order`, `import_groups` |
| `no_leading_import_slash`, `single_import_per_statement`, `blank_line_between_import_groups` | the settings of the same name |
| `method_argument_space` (fully multi-line) | `wrap` |
| `yoda_style`, `single_quote`, `no_unused_imports`, `declare_strict_types`, `increment_style`, `heredoc_to_nowdoc` | `yoda`, `single_quote`, `remove_unused_imports`, `declare_strict_types`, `increment_style`, `heredoc_to_nowdoc` ([risky](#risky-settings) in custos) |

Pint's and PHP-CS-Fixer's path exclusions become `format.exclude` globs.

From Mago's formatter: `print-width` is `line_length`, `tab-width` is
`indent_size` (and `tab_width`), `use-tabs` is `indent_style`, and
`single-quote` is the risky `single_quote` setting. The `@mago-format-ignore`
comments keep working.

Rules of these tools that go beyond layout and token spellings, such as
PHPDoc fixes or the order of class members, have no custos setting.
