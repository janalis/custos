---
id: SecurityAdvisories
group: Security
kind: syntax
needs: [composer]
php: { min: "", max: "" }
---

# SecurityAdvisories

## Summary

Works on a project's `composer.json` (not on PHP code). Two concerns:
development-only packages (test frameworks, debuggers, static analysers…)
listed under `require` end up in production; and an application that pulls
third-party packages should also require the advisory "firewall" meta-package
in `require-dev`, which blocks installing versions with known
vulnerabilities.

Throughout this spec:

- **ADV** is the advisory meta-package: vendor `roave`, package
  `security-advisories`; its full name is vendor, `/`, package.
- **CHK** is the alternative checker package: vendor `sensiolabs`, package
  `security-checker`.
- **DEV** is the effective development-package list (option
  `optionConfiguration`, see Options).

## Input

- **D0** The file's base name is exactly `composer.json` (case-sensitive)
  and its top-level JSON value is an object (the *manifest*). Any other
  file (e.g. `any.json`, `composer.lock`) or a manifest that is not an
  object → nothing.
- JSON terms used below: a *property* is a `"key": value` member; its *key
  literal* is the quoted key token including both `"`; "string value"
  means a JSON string (not number/bool/null/array/object). Key and string
  comparisons below use the decoded string contents.
- Top-level properties are looked up by exact key (first occurrence).

## Detection

### Skips (whole file)

- **D1 (library)** The manifest has a `type` property whose value is the
  string `library` (exactly) → nothing is reported.
- **D2 (owner)** Let `N` be the string value of the top-level `name`
  property (if any, and if it is a string). The *owner* `O` is:
  - `N` itself when `N` is an element of DEV (case-insensitive, like every
    DEV comparison: Composer package names are case-insensitive);
  - otherwise, when `N` contains `/`, the prefix of `N` up to and including
    the first `/` (e.g. `acme/`);
  - otherwise none.
  If `O` exists and is an element of DEV (case-insensitive) → nothing is
  reported (the
  manifest belongs to a development package itself, or to a vendor whose
  `vendor/` prefix was added to DEV).
- **D3** The manifest must have a `require` property whose value is an
  object; otherwise nothing is reported (no `require`, or `require` is a
  string/array/…).

### Production packages

Consider each member of the `require` object whose value is a **string**;
other members (numbers, booleans, null, arrays, objects) are ignored. Let
`p` = key lower-cased, `v` = value lower-cased; skip the member when `p` or
`v` is empty.

- **D4 (misplaced dev package)** If option `REPORT_MISPLACED_DEPENDENCIES`
  is on and `p` is an element of DEV (the lower-cased key against the
  lower-cased list entries, so `mikey179/vfsStream` and `PHPUnit/PHPUnit`
  match) → report the member's key literal
  (kind **M**).
- **D5 (third-party flag)** The manifest *has third-party packages* when at
  least one considered member has `p` containing `/` and either there is no
  owner `O` or `p` does not start with `O` lower-cased (package names are
  case-insensitive in Composer, so `Bakery/oven` owns `bakery/core`). Platform entries such as `php`, `ext-json` have no `/`.
- **D6 (secured flag)** The manifest is *secured* when some considered
  member has `p` equal to CHK.

### Advisory checks (only when option `REPORT_MISSING_ROAVE_ADVISORIES` is on)

- **D7** If the manifest has a `require-dev` property whose value is an
  object, scan its string-valued members (same `p`/`v` rules and empty
  skipping as above), in any order:
  - a member with `p` equal to CHK marks the manifest secured;
  - a member with `p` equal to ADV marks it secured, and if `v` is not
    exactly `dev-latest` (comparison after lower-casing, so
    `DEV-LATEST` is fine; `dev-master`, `dev-latest#abc123`, `^1.0` are
    not) → report that member's **value** literal (kind **L**). Stop the
    scan at the first ADV member.
- **D8** If the manifest is not secured and has third-party packages →
  report the key literal of the top-level `require` property (kind **A**).
  The fix F1/F2 is attached to this report.

## Exceptions (no report)

- **E1** Files not named `composer.json`; non-object manifests.
- **E2** `"type": "library"` manifests.
- **E3** Manifests whose owner (D2) is in DEV.
- **E4** No `require` object (no `require`, or `require` not an object):
  no check at all, not even in `require-dev`.
- **E5** Only platform packages, or only packages of the owner's own vendor
  (`acme/shop` requiring `acme/core`) → no kind A.
- **E6** CHK in `require` or `require-dev`, or ADV in `require-dev` → no
  kind A.
- **E7** With `REPORT_MISSING_ROAVE_ADVISORIES` off (the default), kinds A
  and L are never reported; with `REPORT_MISPLACED_DEPENDENCIES` off, kind M
  is never reported.
- **E8** A `require` object with no string values (D5/D6 cannot fire).

## Report

- Range: kind M — the key literal of the offending `require` member,
  including its quotes (`"phpunit/phpunit"`); kind L — the value literal of
  the ADV member in `require-dev`, including quotes; kind A — the key
  literal `"require"` of the top-level `require` property, including
  quotes.
- Severity: warning (all kinds).
- Messages:
  - M: `Development package in require; move it to require-dev.`
  - L: `Constrain the advisory package to dev-latest.`
  - A: `Add the security advisories package (dev-latest) to require-dev.`

## Fix

Only kind A has a fix. Let `RQ` be the top-level `require` property and
`PAIR` the text `"<ADV>": "dev-latest"` (ADV spelled out, one space after
the colon).

- **F1 (no usable require-dev)** The manifest has no `require-dev` property
  whose value is an object: insert, immediately after the end of `RQ` (right
  after the closing `}` of its object value, before whatever followed —
  whitespace, a comma, or the manifest's closing `}`), the text:
  whitespace, `,`, whitespace, `"require-dev": {`, whitespace, `PAIR`,
  whitespace, `}`. Recommended exact insertion (2-space indent):
  `\n  ,\n  "require-dev": {\n    PAIR\n  }`.
  Whitespace-collapsed, the region reads
  `} , "require-dev": { PAIR }` followed by the original continuation
  (e.g. ` }` when `RQ` was the last property, or `, "next": …` otherwise —
  the original comma now follows the inserted property).
- **F2 (existing require-dev object)** Insert immediately after the opening
  `{` of the `require-dev` object the text: whitespace, `PAIR`, whitespace,
  `,`. The original content (including its leading whitespace) follows
  unchanged. Recommended exact insertion: `\n    PAIR\n    ,`.
  Whitespace-collapsed: `"require-dev": { PAIR , <first existing member> …`.
- The conformance comparison collapses whitespace runs to one space, so a
  whitespace run is required at every place listed above (between `}` and
  `,`, between `,` and the next token, after `{`, before `}`, between
  `PAIR` and `,`) and no whitespace may be added elsewhere (e.g. inside
  `PAIR`, which has exactly one space after its colon).

## Options

| Option | Type | Default | Effect |
|---|---|---|---|
| `REPORT_MISSING_ROAVE_ADVISORIES` | bool | false | Enables kinds A (missing advisory package, with fix) and L (advisory package not on `dev-latest`). |
| `REPORT_MISPLACED_DEPENDENCIES` | bool | true | Enables kind M (development packages under `require`). |
| `optionConfiguration` | list of strings | the default list below | The development-package list DEV, used for kind M and for the owner skip (D2). |

Effective DEV: upstream always merges the default list into whatever the
user configured (user entries ∪ defaults, deduplicated). Recommendation for
custos: DEV = defaults ∪ user entries. Entries may also be vendor prefixes
ending in `/` (only meaningful for the owner skip D2).

Conformance note: upstream test cases that do not add the defaults run with
an *empty* DEV (the merge happens only when settings are loaded). With the
upstream fixtures this makes no difference to the outcome, except that the
two cases which explicitly add the defaults are the only ones where kind M
or the owner skip can fire. Using the defaults always is safe.

Default list (vendor / package; the list entry is `vendor/package`, compared
case-insensitively — the entry is stored as `mikey179/vfsStream`):

| Vendor | Package | | Vendor | Package |
|---|---|---|---|---|
| phpunit | phpunit | | phpspec | prophecy |
| johnkary | phpunit-speedtrap | | phpspec | phpspec |
| brianium | paratest | | humbug | humbug |
| mybuilder | phpunit-accelerator | | infection | infection |
| codedungeon | phpunit-result-printer | | mockery | mockery |
| spatie | phpunit-watcher | | satooshi | php-coveralls |
| symfony | phpunit-bridge | | mikey179 | vfsStream |
| symfony | debug | | filp | whoops |
| symfony | maker-bundle | | friendsofphp | php-cs-fixer |
| zendframework | zend-test | | phpstan | phpstan |
| zendframework | zend-debug | | vimeo | psalm |
| yiisoft | yii2-gii | | jakub-onderka | php-parallel-lint |
| yiisoft | yii2-debug | | squizlabs | php_codesniffer |
| orchestra | testbench | | slevomat | coding-standard |
| barryvdh | laravel-debugbar | | doctrine | coding-standard |
| codeception | codeception | | phpcompatibility | php-compatibility |
| behat | behat | | zendframework | zend-coding-standard |
| yiisoft | yii2-coding-standards | | wp-coding-standards | wpcs |
| phpmd | phpmd | | pdepend | pdepend |
| sebastian | phpcpd | | povils | phpmnd |
| phan | phan | | phpro | grumphp |
| wimg | php-compatibility | | sstalle | php7cc |
| phing | phing | | composer | composer |
| roave | security-advisories | | kalessil | production-dependencies-guard |

(48 entries.) ADV itself is in the list, so ADV under `require` is reported
as kind M.

## PHP versions

Not applicable (JSON input).

## Examples

All examples assume both bool options on and the default DEV list.

Misplaced packages plus missing advisory (new section, F1):

```json
{
  "name": "bakery/oven",
  <warning descr="Add the security advisories package (dev-latest) to require-dev.">"require"</warning>: {
    "php": ">=8.1",
    "guzzlehttp/guzzle": "^7.0",
    <warning descr="Development package in require; move it to require-dev.">"Mockery/Mockery"</warning>: "^1.6"
  },
  "license": "MIT"
}
```

```json
{
  "name": "bakery/oven",
  "require": {
    "php": ">=8.1",
    "guzzlehttp/guzzle": "^7.0",
    "Mockery/Mockery": "^1.6"
  }
  ,
  "require-dev": {
    "ADV": "dev-latest"
  },
  "license": "MIT"
}
```

Existing `require-dev` (F2):

```json
{
  "name": "bakery/oven",
  <warning descr="Add the security advisories package (dev-latest) to require-dev.">"require"</warning>: {
    "monolog/monolog": "^3.0"
  },
  "require-dev": {
    "bakery/fixtures": "*"
  }
}
```

```json
{
  "name": "bakery/oven",
  "require": {
    "monolog/monolog": "^3.0"
  },
  "require-dev": {
    "ADV": "dev-latest"
    ,
    "bakery/fixtures": "*"
  }
}
```

Advisory present but not constrained to `dev-latest` (kind L; the manifest
counts as secured, so no kind A):

```json
{
  "name": "",
  "require": { "nesbot/carbon": "^2.0" },
  "require-dev": { "ADV": <warning descr="Constrain the advisory package to dev-latest.">"dev-main"</warning> }
}
```

No reports: `{"name": "bakery/oven", "require": {"php": "^8.2", "bakery/core": "1.*"}}`;
`{"type": "library", "require": {"x/y": "*"}}`; `{"name": "phpstan/phpstan",
"require": {"phpunit/phpunit": "*"}}` (owner in DEV); `{"require": "x/y"}`;
`{"require": {"a": 1, "b": [], "c": null}}`; a manifest requiring CHK.

(In the examples `ADV` stands for the advisory package name spelled out in
full, `vendor/package`, exactly as the fix must produce it.)

## Divergences

- **Case-insensitive DEV list (custos diverges from upstream).** Upstream
  lower-cases the package keys but compares them with the list entries as
  stored, so the default entry `mikey179/vfsStream` can never match and a
  user-configured entry with capitals is ignored. Composer package names
  are case-insensitive; custos lower-cases both sides (D2, D4).
- F2 on an empty `require-dev` object (`{}`) produces `{ PAIR , }` (invalid
  JSON) upstream. Recommendation: omit the comma when the object has no
  members.
- F1 when a `require-dev` property exists but is not an object
  (e.g. `[]`) adds a second `require-dev` key upstream. Recommendation:
  offer no fix in that case.
- **Owner prefix compared case-insensitively — custos diverges from
  upstream** (D5). Upstream compares the lower-cased package key with the
  owner prefix as written, so a mixed-case `name` (`Bakery/oven`) makes the
  project's own packages (`bakery/core`) count as third-party and asks for
  the advisory package although only first-party code is required. custos
  lower-cases the owner prefix too, matching Composer's case-insensitive
  package names.
- **Metapackages (custos diverges).** A manifest with
  `"type": "metapackage"` has no code of its own: its `require` list is the
  package, typically pulled into another project's `require-dev` (a
  "dev tools" bundle). Reporting development packages in its `require`, or
  a missing advisories package, is wrong. custos skips metapackages like
  libraries (D1).
