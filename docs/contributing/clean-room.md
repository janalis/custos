# Clean-room process

custos is released under the MIT license. Part of its rule catalogue is
modeled on
[Php Inspections (EA Extended)](https://github.com/kalessil/phpinspectionsea),
which is LGPL-2.1. To keep custos free of upstream code and text, every
EA-modeled rule goes through a two-step **clean-room** process. This page is
the reference for it; other pages only summarize it.

## What may be reused

Only **facts**, all of them already captured in `internal/inspection/meta/rules.json` by
`make extract`:

- rule IDs (the upstream short name without `Inspection`), kept so that
  `@noinspection` comments keep working;
- groups, default severity and enabled flags;
- option names and defaults;
- PHP-version thresholds.

Never copy, translate or closely paraphrase upstream Java code, messages,
descriptions, documentation or test fixtures. No upstream file is ever copied
into this repository.

## Two steps, two people (or sessions)

1. **Specify.** Someone reads the upstream inspection and writes
   `specs/<ID>.md` **in their own words**, with **new** PHP examples: what is
   detected, the exceptions, the options, the PHP-version gating and what the
   fix produces. The spec must not contain upstream code or text.
2. **Implement.** Someone else implements the rule **from the spec only**.
   They do not open the upstream Java sources or fixtures. When the spec is
   wrong or incomplete, the gap goes back to step 1, not to the upstream
   sources. Rule implementation notes (`NOTES.md`) are written by the
   implementer, in their own words, and never quote upstream code or text.

The spec is also the source of the user documentation: its *Summary* and
*Options* sections become `custos explain` and the rule's page on this site,
and its first example is shown on that page.

## Checking against upstream without copying it

`make conformance` reads the upstream fixtures directly from a local checkout
at test time (`EA_PATH`). For each case it runs the rule under test and
compares the severity and range of each finding, and the result of applying
fixes. Upstream message text is never compared or stored. The repository's
own fixtures (`internal/inspection/rules/<id>/testdata/`) are the CI gate.

`make cleanroom` scans `specs/`, `internal/`, `cmd/` and `docs/` and fails
on any hit:

- forbidden tokens, always: the upstream message prefix (`EA` in square
  brackets) and names of upstream Java classes (IDE API types and the
  plugin's helper classes), listed in `tools/cleanroom`;
- when the local upstream checkout exists, long strings that appear verbatim
  in its Java sources or fixtures.

The conformance harness under `internal/testing/conformance/` is exempt from
the token check, since it parses that prefix.

Conformance failures must not expose upstream messages or expected fixed source
to implementers. Missing expectations are reported by severity and range; fix
mismatches show the divergence offset and custos' own output. Keep raw upstream
fixtures and locally extracted case data outside version control.

## Divergences

custos fixes upstream bugs instead of reproducing them. Each intentional
difference is documented in the spec's *Divergences* section and listed in
`internal/testing/conformance/testdata/ea-divergences.json`, so conformance runs know about it.
