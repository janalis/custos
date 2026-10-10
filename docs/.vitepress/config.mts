import { execFileSync } from 'node:child_process'
import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'
import rulesSidebar from './rules-sidebar.json'

const repo = 'https://github.com/janalis/custos'
const base = '/custos/'
const site = 'https://janalis.github.io' + base

// Latest release tag, shown in the nav ("dev" outside a git checkout).
function version(): string {
  try {
    return execFileSync('git', ['describe', '--tags', '--abbrev=0'], { stdio: ['ignore', 'pipe', 'ignore'] }).toString().trim()
  } catch {
    return 'dev'
  }
}

export default withMermaid(defineConfig({
  title: 'custos',
  description: 'Fast PHP inspector and fixer: PHP inspections with quick-fixes, a single static binary, CLI and LSP.',
  lang: 'en-US',
  base,
  cleanUrls: true,
  lastUpdated: true,
  srcExclude: ['internals/**'],
  sitemap: { hostname: site },
  head: [
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: base + 'favicon-32.png' }],
    ['link', { rel: 'apple-touch-icon', href: base + 'apple-touch-icon.png' }],
    ['meta', { name: 'theme-color', content: '#4752B5' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:site_name', content: 'custos' }],
    ['meta', { property: 'og:image', content: site + 'og.png' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
  ],
  themeConfig: {
    logo: { src: '/logo.png', alt: 'custos' },
    nav: [
      { text: 'Guide', link: '/guide/introduction', activeMatch: '/guide/' },
      { text: 'Rules', link: '/rules/', activeMatch: '/rules/' },
      { text: 'Contributing', link: '/contributing/', activeMatch: '/contributing/' },
      {
        text: version(),
        items: [
          { text: 'Changelog', link: '/changelog' },
          { text: 'Releases', link: repo + '/releases' },
        ],
      },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is custos?', link: '/guide/introduction' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Getting started', link: '/guide/getting-started' },
          ],
        },
        {
          text: 'Usage',
          items: [
            { text: 'Command line', link: '/guide/cli' },
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Native inspections', link: '/guide/native-rules' },
            { text: 'Suppressing findings', link: '/guide/suppressing' },
            { text: 'Continuous integration', link: '/guide/ci' },
            { text: 'Editors (LSP)', link: '/guide/editors' },
          ],
        },
        { text: 'Rule reference', link: '/rules/' },
      ],
      '/rules/': [{ text: 'All rules', link: '/rules/' }, ...rulesSidebar],
      '/contributing/': [
        {
          text: 'Contributing',
          items: [
            { text: 'Getting set up', link: '/contributing/' },
            { text: 'Clean-room process', link: '/contributing/clean-room' },
            { text: 'Architecture', link: '/contributing/architecture' },
            { text: 'Adding a rule', link: '/contributing/adding-a-rule' },
            { text: 'Fixtures', link: '/contributing/fixtures' },
            { text: 'Releasing', link: '/contributing/releasing' },
          ],
        },
      ],
    },
    search: { provider: 'local' },
    editLink: {
      // Serialised into the client bundle: no closure variables. Rule pages
      // are generated from their spec, so they link to it.
      pattern: ({ filePath }) =>
        filePath.startsWith('rules/') && filePath !== 'rules/index.md'
          ? 'https://github.com/janalis/custos/edit/main/specs/' + filePath.split('/').pop()
          : 'https://github.com/janalis/custos/edit/main/docs/' + filePath,
      text: 'Edit this page on GitHub',
    },
    socialLinks: [{ icon: 'github', link: repo }],
    outline: { level: [2, 3] },
    footer: {
      message:
        'Released under the MIT License. Rule catalogue modelled on Php Inspections (EA Extended); independent clean-room implementation.',
      copyright: 'Builtin symbol data from JetBrains phpstorm-stubs (Apache-2.0).',
    },
  },
}))
