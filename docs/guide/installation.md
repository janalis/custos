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

## Check the installation

```sh
custos version
custos rules | head
```

Next: [Getting started](./getting-started).
