# AI agents (MCP)

Run `custos mcp` to expose PHP analysis and navigation to an MCP client.
The default transport is stdio. custos reads PHP as data; it never executes
PHP, Composer scripts or shell commands requested by source or PHPDoc.
Conversation history belongs to the agent host; checkpoints keep analysis
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

The workspace root is chosen by the first of these that applies:

1. an explicit `--workspaces` file or `--root` flag;
2. the Composer launcher: its own path and the project's manifest identify
   the consuming project, even when the client starts it from a nested or
   unrelated working directory. Custom `vendor/bin` paths and symlinked
   package installations work, and a `CUSTOS_BINARY` override does not
   change this;
3. the [project root](./configuration#project-root) found from the working
   directory, as for a global installation.

The launcher preserves the working directory. A root found automatically is
registered as the workspace `default`. A `--workspaces` file may omit a
default; callers must then pass `workspaceId`.

custos tests SDK client connections over stdio and Streamable HTTP for the
SDK-supported protocol revisions. These are protocol tests, not
certification of a particular agent application's integration.

## Tools

| Tool | Use |
|---|---|
| `custos_workspace` | Authorized workspaces, effective settings, provenance and checkpoints. |
| `custos_analyze` | Relative paths or a PHP buffer; returns a checkpoint ID and counts by severity, rule and file. |
| `custos_findings` | Filter by rule, severity and file or directory; page findings, excerpts, byte ranges and fix choices. |
| `custos_rules` | Rule explanations, options, PHP thresholds and suppressions, searched by words or rule ID. |
| `custos_symbols` | Project and vendor declarations with PHP-style signatures, filtered by kind and scope. |
| `custos_type` | Native and PHPDoc-aware expression types at a line and column, or a byte offset. |
| `custos_prepare_fixes` | Selected fixes or bounded iterative fixes; previews and a plan ID. |
| `custos_apply_fixes` | Apply a disk plan once and return file outcomes and fresh analysis. |
| `custos_checkpoints` | List, resume or delete the caller's checkpoints. |

Analyze first, then pass `checkpointId` to the findings and navigation
tools. Positions use zero-based byte offsets and one-based lines and rune
columns; `custos_type` accepts either a `line` and `column` or an `offset`.
Each result is one JSON text block. Keep filters unchanged when passing
`nextCursor` as the next call's `cursor`.

Unlike the command line and `custos.json`, the MCP tools match rule IDs
without regard to case. A findings `path` selects one file or everything
below a directory. Symbol search ranks an exact name first, then a whole
trailing segment such as a short class or method name, then other matches,
with project declarations before vendor ones; `scope` is `project`,
`vendor` or `all`, and defaults to `project` when no query is given.

### What is analyzed

Project configuration, excludes, suppression comments, the baseline and PHP
version gates apply, as on the command line. A symlink under an analyzed
path is rejected. Symlinks under the vendor directory are never followed:
they are skipped, like vendor files too large or impossible to parse, and
`custos_analyze` lists them under `limitations` because types declared there
stay unknown. The Composer-installed `janalis/custos` package is left out of
the index.

A supplied buffer is analyzed instead of the disk file of the same path: its
declarations replace that file's, and the rest of the semantic context comes
from the captured disk snapshot. Runtime declarations and ambiguous symbols
can limit type precision.

### Fixes

Select findings by their `id` and fixes by their zero-based index:

```json
{
  "checkpointId": "returned-checkpoint-id",
  "choices": [{ "findingId": "0", "fix": 0 }]
}
```

Alternatively, set `iterative` to true, optionally restricting `path` or
`rule`. Pass the returned `planId` to `custos_apply_fixes`. A plan built
from a supplied buffer can be previewed but never written to disk.

Preview pages show a diff per file and its exact edits, as line replacements
for iterative plans; large replacements are split into text chunks with byte
offsets. `textBase64` is the authoritative chunk payload, including for
source bytes that are not valid UTF-8; `text` is a readable preview.

Applying a plan checks that the checkpoint inputs are unchanged, rejects
conflicts and unparsable output, serializes the server's writes, and
atomically replaces each file with its permissions preserved. Failures are
reported per file; an application is not a multi-file transaction. The
checkpoint of an applied plan is replaced by the analysis returned with the
outcomes.

External editors do not share the server's write lock, so coordinate edits
while applying a plan. custos rechecks each file right before replacing it,
but operating systems offer no portable compare-and-swap rename against an
editor writing at the same time.

## Limits

| Limit | Value |
|---|---|
| Page size | 20 items by default, up to 100, and under 32 KiB |
| Project snapshot | 10,000 files and 64 MiB, held in memory |
| Vendor files | 100,000 files and 1 GiB, read for their declarations and a hash, not kept |
| Concurrent calls | one operation at a time; 16 calls in all, running or waiting, then callers are told to retry |
| Call duration | 2 minutes; cancellation is checked between bounded file operations |
| Fix plans | 20 at a time, 100 files each, 64 MiB of captured source and output in all; a plan expires after ten minutes or after one application attempt |
| Checkpoints | 20 and 64 MiB of evidence per identity and workspace; 1,000 and 64 MiB in all; kept seven days |
| Source snapshots in memory | two, 64 MiB in all; the least recently used is evicted first |
| Workspaces | 64 configured workspaces, 256 bearer credentials |

## Checkpoints

Every tool call checks that its checkpoint still matches the workspace.
While the server holds the checkpoint's snapshot, a file whose size and
modification time are unchanged is trusted, unless it was modified within
two seconds of the analysis; any other file is hashed again. A changed,
added or removed input makes the checkpoint stale, and the error names the
first paths: analyze again. Unchanged vendor declarations are reused by
later analyses of the same workspace.

Checkpoints are stored under `custos/mcp` in the operating system's user
cache directory, or in the directory given by `--storage` or the workspaces
file. It must be private and outside every workspace root. The evidence
includes findings, fingerprints, input hashes and manifests; syntax trees,
supplied buffers, credentials and fixes are not stored. Checkpoints older
than seven days are removed at startup, including those of other workspace
roots sharing the directory.

An evicted disk snapshot is rebuilt when needed; an evicted buffer needs a
new analysis. Resuming after a restart hashes the configuration, the
discovered files and every semantic input, vendor declarations included.
Changed inputs make the evidence stale; rebuilt findings also reveal rules
that behave differently after a server update. Corrupt records are ignored.
Checkpoints do not restore an agent's conversation.

## Remote deployment

HTTP requires authentication: bearer credentials, OAuth, or both, set in a
private `--workspaces` file. Stdio callers use the identity `local`.

A bearer credential is stored as the SHA-256 digest of its token. Generate a
strong random token outside the repository, compute its digest, and give the
token to your client through its secret storage, never in a URL.

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

Send `Authorization: Bearer YOUR_TOKEN` to `/mcp`. There is no
unauthenticated HTTP mode. Host headers are checked against `hosts`, and a
nonempty Origin header must match `origins`. HTTP sessions are stateless, so
a session ID cannot grant access. Checkpoints and plans are only available to
the identity and workspace that created them. Writes also require the
workspace's write permission and the `custos:write` scope.

### Workspaces file

Unknown keys are rejected.

| Key | Meaning |
|---|---|
| `workspaces` | List of `{id, root, readers, writers}`; `readers` and `writers` are identities. |
| `default` | ID of the workspace used when a call gives no `workspaceId`. |
| `storage` | Checkpoint directory, as `--storage`. |
| `hosts` | Allowed Host headers. |
| `origins` | Allowed Origin headers. |
| `publicUrl` | HTTPS URL of the server, ending in `/mcp`; required for public binding and OAuth. |
| `credentials` | Bearer credentials: `sha256` (64 hex characters), `identity`, `expires` (required), `scopes` (must include `custos:read`). |
| `oauth` | External-provider OAuth; see below. |

For public binding, supply `--tls-cert` and `--tls-key`, a `publicUrl`, and
explicit `hosts` and `origins`. The example above binds loopback. TLS
termination, certificate management and network access policy belong to the
deployment operator.

### OAuth

custos can accept tokens from an external OAuth provider:

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
`issuer`, `introspection` and `publicUrl` must be HTTPS URLs, `audience` must
equal `publicUrl`, and the introspection client ID and secret are read from
the environment variables named by `clientIdEnv` and `clientSecretEnv`.
custos publishes protected-resource metadata at
`/.well-known/oauth-protected-resource` and its `/mcp` variant.
Introspection uses HTTPS, a bounded response and no redirects. Tokens must be
active, unexpired, issued for the audience and scoped for `custos:read`. The
identity is the issuer followed by `#` and the token's subject. The example
has no default workspace, so calls must name one.
