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
| `HTTPS_PROXY` | Proxy for the download. |

## Release archives

Download the archive for your platform from the
[releases page](https://github.com/janalis/custos/releases), check it
against `SHA256SUMS`, extract `custos`, and put it on your `PATH`.

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
