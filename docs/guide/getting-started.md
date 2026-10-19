# Getting started

## 1. Analyze

From your project root (the directory holding `composer.json`):

```sh
custos analyse
```

custos checks the whole project, minus the
[default excludes](./configuration#excluded-directories), against your
[target PHP version](./configuration#target-php-version), usually taken from
`composer.json`. Each finding gives its position, severity, message and rule
ID, and says when a quick-fix is available:

```text
src/Invoice.php:3:5: warning: Variable $total is used only once; inline its value. [OneTimeUseVariables] (fixable)
src/Invoice.php:6:7: info: Stray semicolon; remove it. [UnnecessarySemicolon] (fixable)
src/Invoice.php:7:6: error: Restrict the classes unserialize() may create via its second argument. [UnserializeExploits]

1 file(s) analysed: 1 error(s), 1 warning(s), 1 info
```

The command fails when a finding is a warning or worse; `--fail-on` changes
the threshold (see [exit codes](./cli#exit-codes)).

Want to know more about a rule? Run `custos explain OneTimeUseVariables`, or
open its page in the [rule reference](/rules/).

## 2. Fix

Preview the quick-fixes as a diff, then apply them:

```sh
custos fix --dry-run --diff src
custos fix src
```

```diff
--- a/src/Invoice.php
+++ b/src/Invoice.php
@@ -1,6 +1,5 @@
 <?php
 function total($a) {
-    $total = $a + 1;
-    return $total;
+    return $a + 1;
 }
-foo();;
+foo();
```

Fix one rule at a time if you want small, reviewable commits:

```sh
custos fix --rule UnnecessarySemicolon src
```

## 3. Configure

Create a `custos.json` at the project root to pin the PHP version, choose
paths, and turn rules on or off:

```json
{
  "php": "8.3",
  "paths": ["src", "tests"],
  "rules": {
    "StaticClosureCanBeUsed": { "enabled": false }
  }
}
```

See [Configuration](./configuration).

## 4. Adopt on legacy code

On an existing code base, record the current findings in a
[baseline](./suppressing#baseline) so that only new ones are reported:

```sh
custos analyse --generate-baseline custos-baseline.json
```

## 5. Wire it in

- CI: [GitHub Actions, GitLab and others](./ci).
- Editor: [diagnostics and quick-fixes as you type](./editors).
