# Clean-room process

custos is released under the MIT license. Its rule catalogue is modelled on
[Php Inspections (EA Extended)](https://github.com/kalessil/phpinspectionsea),
which is LGPL-2.1. To keep custos free of upstream code and text, every rule
goes through a two-step **clean-room** process. This page explains it.

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
   sources.

The spec is also the source of the user documentation: its *Summary* and
*Options* sections become `custos explain` and the rule's page on this site,
and its first example is shown on that page.

## Checking against upstream without copying it

`make conformance` reads the upstream fixtures directly from a local checkout
at test time (`EA_PATH`). It compares only the rule, the range and the
severity of each finding, and the result of applying fixes. Upstream message
text is never compared or stored. The repository's own fixtures
(`testdata/rules/`) are the CI gate.

`make cleanroom` scans the repository (specs, code, test data, docs) for long
strings that appear verbatim in the local upstream checkout and fails on any
hit.

## Divergences

custos fixes upstream bugs instead of reproducing them. Each intentional
difference is documented in the spec's *Divergences* section and listed in
`testdata/ea-divergences.json`, so conformance runs know about it.
