# AI agents (MCP)

Run `custos mcp` to expose PHP analysis and navigation to an MCP client.
The default transport is stdio. Custos reads PHP as data; it never executes
PHP, Composer scripts or shell commands requested by source or PHPDoc.
Conversation history belongs to the agent host. Checkpoints retain analysis
evidence across server restarts.

## Local client setup

Configure your client's MCP process entry, adapting its configuration format:

```json
{
  "mcpServers": {
    "custos": {
      "command": "/absolute/path/to/custos",
      "args": ["mcp", "--root", "/absolute/path/to/project"]
    }
  }
}
```

With Composer, use the consuming project's launcher:

```json
{
  "mcpServers": {
    "custos": {
      "command": "/absolute/path/to/project/vendor/bin/custos",
      "args": ["mcp"]
    }
  }
}
```

Composer's proxy path and manifest identify the consuming project even when
started from a nested or unrelated working directory. Custom vendor/bin paths
and symlinked package installations are supported. A binary override through
`CUSTOS_BINARY` does not change discovery. Global installations discover from
the working directory. The launcher preserves the working directory.

Root precedence is explicit `--workspaces` configuration or `--root`, then the
verified Composer launcher hint, then existing working-directory discovery.
Automatic discovery registers the root as `default`. Explicit configuration
may omit a default; callers must then provide `workspaceId`.

Custos tests SDK client connections over stdio and Streamable HTTP for the
SDK-supported revisions. These are protocol tests, not certification of a
particular agent application's integration.

## Tool workflow

| Tool | Use |
|---|---|
| `custos_workspace` | Authorized workspaces, effective settings, provenance and checkpoints. |
| `custos_analyze` | Relative paths or a PHP buffer; returns a checkpoint ID and counts by severity, rule and file. |
| `custos_findings` | Filter by rule, severity and file or directory; page findings, excerpts, byte ranges and fix choices. |
| `custos_rules` | Clean-room explanations, options, PHP thresholds and suppressions, searched by words or rule ID. |
| `custos_symbols` | Project/vendor declarations with PHP-style signatures, filtered by kind and scope. |
| `custos_type` | Native and PHPDoc-aware expression types at a line and column, or a byte offset. |
| `custos_prepare_fixes` | Selected fixes or bounded iterative fixes; previews and a plan ID. |
| `custos_apply_fixes` | Apply a disk plan once and return file outcomes and fresh analysis. |
| `custos_checkpoints` | List, resume or delete the caller's checkpoints. |

Analyze first, then pass `checkpointId` to navigation and findings tools.
Positions use zero-based byte offsets and one-based lines and rune columns;
`custos_type` accepts either a `line` and `column` or an `offset`. Each result
is one JSON text block. Pages default to 20 items, accept up to 100 and stay
below 32 KiB. Keep filters unchanged when using `nextCursor` as the next
call's `cursor`.

Rule IDs are matched without regard to case. A findings `path` selects one
file or everything below a directory. Symbol search ranks an exact name first,
then a whole trailing segment such as a short class or method name, then other
matches, with project declarations before vendor ones; `scope` is `project`,
`vendor` or `all`, and defaults to `project` when no query is given.

Select findings by their `id` and fixes by their zero-based index:

```json
{
  "checkpointId": "returned-checkpoint-id",
  "choices": [{ "findingId": "0", "fix": 0 }]
}
```

Alternatively, set `iterative` to true, optionally restricting `path` or `rule`.
Pass the returned `planId` to `custos_apply_fixes`. Plans live only in memory,
expire after ten minutes and become unusable after an application attempt.
Preview pages expose a diff per file and its exact edits, as line replacements
for iterative plans; large replacements split into text chunks with byte offsets. `textBase64` is the authoritative chunk payload, including
for source bytes that are not valid UTF-8. `text` is a readable preview. Supplied buffers permit previews but never disk writes.
Their declarations replace those of the same disk file; other semantic context
comes from the captured disk snapshot; runtime declarations and
ambiguous symbols can limit type precision.

Project configuration, excludes, suppression comments, baselines and PHP
version gates apply. A symlink under an analyzed path is rejected. Symlinks
under the vendor directory are never followed: they are skipped, like vendor
files too large or impossible to parse, and `custos_analyze` lists them under
`limitations` because types declared there stay unknown. The Composer-installed
`janalis/custos` package is omitted from semantic indexing.

Project inputs are captured in memory, bounded to 10,000 files and 64 MiB per
snapshot. Vendor files are read for their declarations and a hash but not
retained, bounded to 100,000 files and 1 GiB. One tool operation runs at a
time; up to 16 calls wait their turn, further ones are told to retry.
Cancellation is checked between bounded file operations.

Every tool call checks that its checkpoint still matches the workspace. While
the server holds the checkpoint's snapshot, a file whose size and modification
time are unchanged is trusted, unless it was modified within two seconds of
the analysis; any other file is hashed again. A changed, added or removed
input makes the checkpoint stale and the error names the first paths: analyze
again. Unchanged vendor declarations are reused by later analyses of the same
workspace.

## Checkpoint storage

The default is `custos/mcp` under the operating system's user cache directory.
Override it with `--storage` or the workspace configuration. The directory must
be private and outside workspace roots. Evidence includes findings,
fingerprints, input hashes and manifests. ASTs, supplied buffers, credentials
and fix closures are not persisted.

Retention is seven days, including checkpoints written for other workspace
roots that share the directory, which are removed at startup. At most 20 checkpoints and 64 MiB of serialized
evidence per identity/workspace, with global bounds of 1,000 checkpoints and 64 MiB of evidence, and up to 64
configured workspaces. At most two source snapshots, totaling
64 MiB, remain in memory; the least recently used one is evicted first. Evicted disk snapshots can be reconstructed; evicted
buffers require a new analysis. Fix plans are bounded to 20 globally, 100 files
per plan and 64 MiB of captured source/output globally.

Resume after a restart hashes configuration, discovered files and all semantic inputs,
including vendor declarations. Changed inputs mark evidence stale and require
fresh analysis. Rebuilt findings also detect changed rule behavior after a
server update. Corrupt records are ignored. Checkpoints do not restore an
agent's conversation.

## Remote deployment

HTTP requires configured authentication. A private bearer credential is stored
as a SHA-256 digest in a private configuration file. Generate a strong random
credential outside the repository, calculate its digest and deliver the token
through your client's secret storage. Never put it in a URL.

```json
{
  "default": "app",
  "storage": "/private/custos-checkpoints",
  "hosts": ["127.0.0.1:8765"],
  "origins": [],
  "workspaces": [{
    "id": "app",
    "root": "/srv/app",
    "readers": ["agent-a"],
    "writers": ["agent-a"]
  }],
  "credentials": [{
    "sha256": "replace-with-64-hex-character-token-digest",
    "identity": "agent-a",
    "expires": "2027-01-01T00:00:00Z",
    "scopes": ["custos:read", "custos:write"]
  }]
}
```

```sh
custos mcp --workspaces /private/mcp.json --http 127.0.0.1:8765
```

Send `Authorization: Bearer YOUR_TOKEN` to `/mcp`. No unauthenticated HTTP
mode exists. Host headers are allowlisted; nonempty Origin headers must match
configured origins. HTTP sessions are stateless, so a session ID cannot grant
access. Checkpoint and plan access always requires the same stable identity
and workspace. Writes additionally require workspace write permission and
`custos:write` scope. Stdio uses the `local` identity.

For public binding, supply `--tls-cert` and `--tls-key`, an HTTPS `publicUrl`
ending in `/mcp`, and explicit `hosts` and `origins`. The default deployment
example binds loopback. TLS termination, certificate management and network
access policy belong to the deployment operator.

External-provider OAuth is also supported:

```json
{
  "publicUrl": "https://custos.example/mcp",
  "hosts": ["custos.example"],
  "origins": ["https://agent.example"],
  "oauth": {
    "issuer": "https://identity.example",
    "introspection": "https://identity.example/oauth/introspect",
    "audience": "https://custos.example/mcp",
    "clientIdEnv": "CUSTOS_INTROSPECTION_CLIENT",
    "clientSecretEnv": "CUSTOS_INTROSPECTION_SECRET"
  },
  "workspaces": [{
    "id": "app",
    "root": "/srv/app",
    "readers": ["https://identity.example#subject-123"],
    "writers": []
  }]
}
```

The provider owns login, consent, client registration and token issuance.
Custos publishes protected-resource metadata at
`/.well-known/oauth-protected-resource` and its `/mcp` variant. Introspection
uses HTTPS, a bounded response and no redirects. Tokens must be active,
unexpired, audience-matching and scoped for `custos:read`. Identity is the
configured issuer followed by `#` and the returned stable subject. The example
omits a default, requiring explicit workspace selection.

Fix application checks that the checkpoint inputs are unchanged, rejects conflicts and unparsable
output, serializes server writes and atomically replaces each file with its
permissions preserved. Per-file failures are explicit; application is not a
multi-file transaction. External editors do not share the server's write lock,
so coordinate edits while applying a plan. An applied plan's checkpoint is
replaced by the analysis returned with the outcomes. Custos rechecks content immediately
before replacement; the operating system offers no portable compare-and-swap
rename against an independently writing editor.
