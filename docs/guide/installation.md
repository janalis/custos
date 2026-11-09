# Installation

custos is a single static binary for Linux, macOS and Windows (amd64 and
arm64). It does not need PHP to run.

## Homebrew (macOS, Linux)

```sh
brew install janalis/tap/custos
```

## Composer (per project)

```sh
composer require --dev janalis/custos
vendor/bin/custos version
```

The Composer package is a small PHP launcher (PHP 7.4+). On first run it
downloads the binary for its version and your platform from the GitHub
release, verifies it against `SHA256SUMS`, and caches it inside the package
directory. Environment variables:

| Variable | Effect |
|---|---|
| `CUSTOS_BINARY` | Run this binary instead; nothing is downloaded. |
| `CUSTOS_DOWNLOAD_URL` | Base URL, or a local directory, holding the release assets (a mirror). |
| `HTTPS_PROXY` (or `https_proxy`) | Proxy for the download. |

When the download fails, the launcher prints why and exits 2.

## Release downloads

Each [release](https://github.com/janalis/custos/releases) has, for every
platform:

- an archive (`.tar.gz`, or `.zip` on Windows) holding `custos` (`custos.exe`
  on Windows) with the license and README;
- the bare binary, `custos_<os>_<arch>` (for example `custos_linux_arm64`),
  handy in CI.

Check the download against `SHA256SUMS`, make the binary executable if
needed, and put it on your `PATH` as `custos`.

## From source

With Go 1.27 or later:

```sh
git clone https://github.com/janalis/custos
cd custos
make build          # → bin/custos
```

## Updating

`custos self-update` updates custos the way it was installed:

| Install | What `self-update` does |
|---|---|
| Homebrew | Runs `brew upgrade janalis/tap/custos`. |
| Composer | Runs `composer update janalis/custos` in the project requiring custos. The constraint in `composer.json` still applies: to move past it, run `composer require --dev janalis/custos:^X.Y.Z`. A `CUSTOS_BINARY` override is updated the way that binary was installed instead. |
| Release download | Downloads the bare binary of the latest release, checks it against `SHA256SUMS` and replaces the running binary. `CUSTOS_DOWNLOAD_URL` and `HTTPS_PROXY` apply as for Composer. The binary's directory must be writable. |
| From source | Refuses: pull and rebuild. |

`custos self-update --check` only reports whether a newer release exists
and exits 1 when one does.

When stderr is a terminal, `analyse` and `fix` print a one-line notice
after their output when a newer release exists, with the command that
updates your install. They look the latest release up at most once a day,
while the command runs (a command that finishes first waits for it at most
one second), and keep the answer in the [cache directory](./cli#paths). The
check is skipped when `CI` is set, for development builds, and when
`CUSTOS_NO_UPDATE_CHECK` is set to any value. The language server and the
MCP server never check.

## Check the installation

```sh
custos version
custos rules | head
```

Next: [Getting started](./getting-started).
