# Platform

`safeio` owns bounded filesystem reads: regular files only, size limits, and
checks on the opened descriptor as well as the path. Repository content is
untrusted, so every production read of project files goes through it.
Packages here depend only on the standard library and each other (see
`docs/contributing/architecture.md`).
