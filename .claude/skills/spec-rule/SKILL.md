---
name: spec-rule
description: Write the clean-room behavioural spec (specs/<RuleID>.md) for one or more custos rules by studying the upstream Php Inspections EA inspection. Use when asked to spec a rule, a batch of rules, or a rule group.
---

# spec-rule

Produce `specs/<RuleID>.md` for each requested rule. You are the **only** step
allowed to read the upstream EA sources; your output must let an implementer
reproduce the behaviour **without** ever opening EA.

## Inputs
- Rule IDs (e.g. `UnnecessarySemicolon`), or a group name from `docs/rules.md`.
- `.cache/ea/index.json` (run `make extract` if missing):
  - `rules.<ID>.source` — inspection Java file (relative to `eaPath`)
  - `rules.<ID>.description`, `rules.<ID>.tests`
  - `cases[]` with `rules` containing the ID — fixture path, `.fixed` path,
    PHP level, options, comparison style
- Facts: `internal/meta/rules.json` entry for the ID.

## Steps
1. Read the inspection class and every helper it relies on (strategy classes,
   `utils/*`, `fixers/*`) until you understand every branch. Note which ones
   need semantic data (reference resolution, types, class hierarchy, index,
   control flow, stubs) → `kind` / `needs` in frontmatter.
2. Read all its fixtures (and `.fixed` files). For each highlight, work out
   **exactly which node/token range** is highlighted and under which
   severity — conformance compares ranges and severities exactly. Note
   messages' *meaning*, not their text.
3. Read the test class to learn option values and PHP levels used.
4. Copy `specs/_TEMPLATE.md` to `specs/<ID>.md` and fill every section:
   numbered detection conditions (D1…), exceptions (E1…), report range,
   fix behaviour (F1…; the fix output is compared whitespace-normalised to EA's
   `.fixed` file, so be precise about produced code), options (names/defaults
   from `rules.json`, effect in your words), PHP-version gating.
5. Write **new** examples: different identifiers, structure and values than
   any EA fixture; cover each D/E/F item at least once.
6. Run `make rules-doc`.

## Clean-room checklist (verify before finishing)
- [ ] No Java code, pseudo-Java, or method/class names of EA internals.
- [ ] No EA message, display name or description text (not even translated
      or lightly reworded) — write our own message wording.
- [ ] No snippet copied or adapted from EA fixtures.
- [ ] Every detection/exception/fix condition observed in EA is captured
      (behaviour is not copyrightable — be exhaustive about it).
- [ ] Ambiguities or likely upstream bugs listed under "Divergences" with a
      recommendation.

## Output
Report per rule: kind, needs, number of D/E/F items, open questions.
