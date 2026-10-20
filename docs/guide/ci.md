# Continuous integration

`custos analyse` fails when a finding reaches `--fail-on` (default `warning`)
or on a configuration error (see [exit codes](./cli#exit-codes)), so it can
gate a pipeline as is. Files that do not parse count as `error` findings
(see [tool findings](./cli#tool-findings)). Pick an
[output format](./cli#analyse) that your CI can display.

Run `composer install` before custos: it reads the symbols under `vendor/`
to know the types your code uses.

## GitHub Actions

Annotations on the pull request diff:

```yaml
- name: custos
  run: vendor/bin/custos analyse --format=github
```

With GitHub code scanning (alerts in the *Security* tab), upload SARIF and
then gate. This runs the analysis twice, once per format:

```yaml
- name: custos (SARIF)
  run: vendor/bin/custos analyse --format=sarif --fail-on=never > custos.sarif
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: custos.sarif
- name: custos (gate)
  run: vendor/bin/custos analyse --format=github --fail-on=warning
```

`--fail-on=never` keeps the report step from failing on findings while
still failing on configuration errors, which `|| true` would hide.

These snippets assume `composer require --dev janalis/custos`. Without
Composer, download the bare binary of a release (replace `latest/download`
with `download/<tag>` to pin a version):

```yaml
- name: Install custos
  run: |
    base=https://github.com/janalis/custos/releases/latest/download
    curl -fsSLo custos "$base/custos_linux_amd64"
    curl -fsSLo SHA256SUMS "$base/SHA256SUMS"
    grep ' custos_linux_amd64$' SHA256SUMS | sed 's/custos_linux_amd64/custos/' | sha256sum -c -
    chmod +x custos
- name: custos
  run: ./custos analyse --format=github
```

## GitLab CI and others

```sh
custos analyse --format=checkstyle > custos-checkstyle.xml   # for Checkstyle-aware viewers
custos analyse --format=json > custos-report.json            # for your own scripts
custos analyse --fail-on=error                               # gate on errors only
```

Do not redirect a report to `custos.json`: that is the configuration file.

## Tips

- Commit a [baseline](./suppressing#baseline) to adopt custos on an existing
  code base without fixing everything first.
- Pin the target PHP version in `custos.json` (`"php": "8.2"`) so CI and
  local runs agree, whatever the PHP version installed.
- `custos fix --dry-run --diff` exits `0` and prints what would change, which
  is handy as an informational job.
