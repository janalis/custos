# CLI adapter

Keep the executable entry in `cmd/custos` thin. Own arguments, configuration
precedence, profiling, output, exit codes and writes. Use project operations
for discovery/indexing/analysis/prepared fixing. Each command has its own file;
shared flags and profile handling live in options. Preserve observable text,
formats, errors, flags and version injection.
