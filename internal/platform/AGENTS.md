# Platform

`safeio` owns bounded filesystem reads: regular files only, size limits, and
checks on the opened descriptor as well as the path. Repository content is
untrusted, so every production read of project files goes through it.
`selfupdate` finds the latest release, tells how the running binary was
installed (Homebrew, the Composer launcher, a release download or a source
build) and replaces a release binary after checking `SHA256SUMS`; the CLI
decides when to check and runs package managers.
Packages here depend only on the standard library and each other (see
`docs/contributing/architecture.md`).
