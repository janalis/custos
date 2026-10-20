---
layout: home

hero:
  name: custos
  text: Fast PHP inspector and fixer
  tagline: Rules with quick-fixes, from probable bugs to PHP 8.5 migration. One static binary for the command line, CI, your editor and AI agents.
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
    title: Rules for PHP
    details: From probable bugs and security issues to code style and language level migration, gated on your target PHP version, from 5.3 to 8.5.
    link: /rules/
    linkText: Rule reference
  - icon: 🛠️
    title: Quick-fixes
    details: Apply automatic fixes where the rule can establish a safe correction. Preview them as a diff with custos fix --dry-run --diff, or apply them all at once.
    link: /guide/cli#fix
    linkText: custos fix
  - icon: ⚡
    title: Fast, no PHP needed
    details: A single static Go binary with its own PHP parser and type inference. It indexes your project and vendor symbols, then analyzes files in parallel.
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
  - icon: 🤖
    title: For AI agents
    details: custos mcp lets coding agents analyze a workspace, read findings and apply quick-fixes through the Model Context Protocol, locally or remotely.
    link: /guide/mcp
    linkText: MCP setup
  - icon: 🤝
    title: PhpStorm-compatible
    details: Rules modeled on Php Inspections (EA Extended) keep their PhpStorm names, so existing @noinspection comments keep working. Native rules are designed for custos.
    link: /guide/suppressing
    linkText: Suppressing findings
---
