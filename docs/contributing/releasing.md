# Releasing

Releases are cut from GitHub Actions in one click. Write user-visible changes
under `## [Unreleased]` in `CHANGELOG.md` as you go
([Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format, semantic
versioning).

## Cutting a release

Run **Actions → release → Run workflow** on `main` with a version: `X.Y.Z`,
or `patch` / `minor` / `major` to bump the latest tag. Tick *dry run* to
build everything without publishing. The workflow:

1. runs `make verify`;
2. runs `tools/relprep`, which turns `[Unreleased]` into `[X.Y.Z] - <date>`,
   pins the version in the Composer launcher (`composer/custos`) and extracts
   the release notes; then commits `Release vX.Y.Z`, tags it and pushes;
3. runs goreleaser (`.goreleaser.yaml`): archives, bare binaries and
   `SHA256SUMS` on the GitHub release, and the cask in
   `janalis/homebrew-tap`;
4. lets Packagist pick up the tag (and pings it when its secrets are set);
5. rebuilds and deploys this documentation site, so the changelog and the
   version in the navigation bar are current.

A failed publish can be retried with *Re-run failed jobs*; the tag is kept.

To build a snapshot locally:

```sh
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean --skip=publish
```

## Documentation site

The site is built with [VitePress](https://vitepress.dev) from `docs/` and
deployed to GitHub Pages by `.github/workflows/docs.yml`: on every push to
`main` that touches the docs, the specs or the generators, after each
release, and on demand (*Run workflow*). Pull requests build it without
deploying, which catches dead links. The build also fails when
`make rules-doc` output is not committed.

```sh
make docs-dev       # live preview on http://localhost:5173/custos/
make docs           # production build in docs/.vitepress/dist
```

## One-time setup

- Create the empty `janalis/homebrew-tap` repository and the
  `HOMEBREW_TAP_TOKEN` secret: a fine-grained token with *Contents: read and
  write* on that repository only.
- Submit `https://github.com/janalis/custos` on packagist.org; its GitHub hook
  then updates on every tag. Optionally add the `PACKAGIST_USERNAME` and
  `PACKAGIST_TOKEN` secrets.
- If `main` is protected, let GitHub Actions bypass the protection so the
  release commit can be pushed.
- **Settings → Pages → Build and deployment → Source: GitHub Actions.**
- Optionally upload `docs/public/og.png` as the repository's social preview
  (**Settings → General → Social preview**).
