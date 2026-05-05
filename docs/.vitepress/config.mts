import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'go-rpc Documentation',
  description: 'Enterprise-grade cross-language RPC framework documentation',

  base: '/go-rpc/',
  lang: 'en-US',
  cleanUrls: false,
  ignoreDeadLinks: true,

  outDir: '.vitepress/dist',

  head: [
    ['link', { rel: 'icon', href: '/logo.svg', type: 'image/svg+xml' }],
  ],

  themeConfig: {
    logo: '/logo.svg',
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Getting Started', link: '/getting-started/' },
      { text: 'Architecture', link: '/architecture/' },
      { text: 'Guides', link: '/guides/' },
      { text: 'API Reference', link: '/api/' },
      { text: 'Deployment', link: '/deployment/' },
    ],

    sidebar: {
      '/': [
        {
          text: 'Getting Started',
          collapsed: true,
          items: [
            { text: 'Quick Start', link: '/getting-started/quick-start/' },
            { text: 'Installation', link: '/getting-started/installation/' },
            { text: 'Basic Usage', link: '/getting-started/basic-usage/' },
          ],
        },
        {
          text: 'Architecture',
          collapsed: true,
          items: [
            { text: 'Overview', link: '/architecture/overview/' },
            { text: 'Core Components', link: '/architecture/core-components/' },
            { text: 'Service Discovery', link: '/architecture/service-discovery/' },
            { text: 'Load Balancing', link: '/architecture/load-balancing/' },
            { text: 'Circuit Breaker', link: '/architecture/circuit-breaker/' },
          ],
        },
        {
          text: 'Guides',
          collapsed: true,
          items: [
            { text: 'Service Registration', link: '/guides/service-registration/' },
            { text: 'Load Balancing Setup', link: '/guides/load-balancing/' },
            { text: 'Observability', link: '/guides/observability/' },
            { text: 'Authentication', link: '/guides/authentication/' },
            { text: 'Rate Limiting', link: '/guides/rate-limiting/' },
          ],
        },
        {
          text: 'API Reference',
          collapsed: true,
          items: [
            { text: 'Go Server', link: '/api/go-server/' },
            { text: 'Go Client', link: '/api/go-client/' },
            { text: 'rpc-gen CLI', link: '/api/rpc-gen/' },
          ],
        },
        {
          text: 'Deployment',
          collapsed: true,
          items: [
            { text: 'Docker Deployment', link: '/deployment/docker/' },
            { text: 'Kubernetes', link: '/deployment/kubernetes/' },
          ],
        },
        {
          text: 'Contributing',
          items: [
            { text: 'Contributing Guide', link: '/contributing/' },
          ],
        },
      ],
    },

    lastUpdated: true,
    editLink: {
      pattern: 'https://github.com/desyang-hub/go-rpc/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/desyang-hub/go-rpc' },
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025-present desyang-hub',
    },
  },

  markdown: {
    config(md) {
      md.linkify.set({ fuzzyEmail: false })
    },
  },
})