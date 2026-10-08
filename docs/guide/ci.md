# Continuous integration

`custos analyse` exits with `1` when a finding reaches `--fail-on` (default
`warning`) and with `2` on configuration errors, so it can gate a pipeline as
is. Pick an output format that your CI can display.

## GitHub Actions

Annotations on the pull request diff:

```yaml
- name: custos
  run: vendor/bin/custos analyse --format=github
```

With GitHub code scanning (alerts in the *Security* tab), upload SARIF and
then gate:

```yaml
- name: custos (SARIF)
  run: vendor/bin/custos analyse --format=sarif > custos.sarif || true
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: custos.sarif
- name: custos (gate)
  run: vendor/bin/custos analyse --format=github --fail-on=warning
```

These snippets assume `composer require --dev janalis/custos`. Without
Composer, download a release archive in a step and run `custos` from there.

## GitLab CI and others

```sh
custos analyse --format=checkstyle > custos-checkstyle.xml   # for Checkstyle-aware viewers
custos analyse --format=json > custos.json                   # for your own scripts
custos analyse --fail-on=error                               # gate on errors only
```

## Tips

- Commit a [baseline](./suppressing#baseline) to adopt custos on an existing
  code base without fixing everything first.
- Pin the target PHP version in `custos.json` (`"php": "8.2"`) so CI and
  local runs agree, whatever the PHP version installed.
- `custos fix --dry-run --diff` exits `0` and prints what would change, which
  is handy as an informational job.
