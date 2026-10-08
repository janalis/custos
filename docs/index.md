---
layout: home

hero:
  name: custos
  text: Fast PHP inspector and fixer
  tagline: 178 inspections with quick-fixes, from probable bugs to PHP 8.5 migration. One static binary for the command line, CI and your editor.
  image:
    src: /logo.png
    alt: custos, an elephant in a police uniform
  actions:
    - theme: brand
      text: Get started
      link: /guide/getting-started
    - theme: alt
      text: Browse the rules
      link: /rules/
    - theme: alt
      text: GitHub
      link: https://github.com/janalis/custos

features:
  - icon: 🔎
    title: 178 inspections
    details: Probable bugs, performance, security, control flow, code style, unused code, PHPUnit and language-level migration from PHP 5.3 to 8.5.
    link: /rules/
    linkText: Rule reference
  - icon: 🛠️
    title: Quick-fixes
    details: Over a hundred rules come with an automatic fix. Preview them as a diff with custos fix --dry-run --diff, or apply them all at once.
    link: /guide/cli#fix
    linkText: custos fix
  - icon: ⚡
    title: Fast, no PHP needed
    details: A single static Go binary with its own PHP parser and type inference. Files are analysed in parallel, each in one pass over its syntax tree.
    link: /guide/installation
    linkText: Install
  - icon: 🧩
    title: In your editor
    details: custos lsp shows diagnostics as you type and offers quick-fixes, fix-all and suppressions. It runs alongside your main PHP language server.
    link: /guide/editors
    linkText: Editor setup
  - icon: 🚦
    title: Built for CI
    details: Text, JSON, Checkstyle, GitHub annotations and SARIF output. A baseline lets legacy projects report only new findings.
    link: /guide/ci
    linkText: CI recipes
  - icon: 🤝
    title: PhpStorm-compatible
    details: Rule IDs match Php Inspections (EA Extended), so your existing @noinspection comments keep working.
    link: /guide/suppressing
    linkText: Suppressing findings
---
