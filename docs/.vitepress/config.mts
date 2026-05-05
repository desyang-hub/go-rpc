import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Go-RPC Documentation',
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
      { text: 'Getting Started', link: '/getting-started/quick-start.html', activeMatch: '^/getting-started/' },
      { text: 'Architecture', link: '/architecture/overview.html', activeMatch: '^/architecture/' },
      { text: 'Guides', link: '/guides/service-registration.html', activeMatch: '^/guides/' },
      { text: 'API Reference', link: '/api/go-server.html', activeMatch: '^/api/' },
      { text: 'Deployment', link: '/deployment/docker.html', activeMatch: '^/deployment/' },
    ],

    sidebar: {
      '/': [
        {
          text: 'Getting Started',
          collapsed: true,
          items: [
            { text: 'Quick Start', link: '/getting-started/quick-start.html' },
            { text: 'Installation', link: '/getting-started/installation.html' },
            { text: 'Basic Usage', link: '/getting-started/basic-usage.html' },
          ],
        },
        {
          text: 'Architecture',
          collapsed: true,
          items: [
            { text: 'Overview', link: '/architecture/overview.html' },
            { text: 'Core Components', link: '/architecture/core-components.html' },
            { text: 'Service Discovery', link: '/architecture/service-discovery.html' },
            { text: 'Load Balancing', link: '/architecture/load-balancing.html' },
            { text: 'Circuit Breaker', link: '/architecture/circuit-breaker.html' },
          ],
        },
        {
          text: 'Guides',
          collapsed: true,
          items: [
            { text: 'Service Registration', link: '/guides/service-registration.html' },
            { text: 'Load Balancing Setup', link: '/guides/load-balancing.html' },
            { text: 'Observability', link: '/guides/observability.html' },
            { text: 'Authentication', link: '/guides/authentication.html' },
            { text: 'Rate Limiting', link: '/guides/rate-limiting.html' },
          ],
        },
        {
          text: 'API Reference',
          collapsed: true,
          items: [
            { text: 'Go Server', link: '/api/go-server.html' },
            { text: 'Go Client', link: '/api/go-client.html' },
            { text: 'rpc-gen CLI', link: '/api/rpc-gen.html' },
          ],
        },
        {
          text: 'Deployment',
          collapsed: true,
          items: [
            { text: 'Docker Deployment', link: '/deployment/docker.html' },
            { text: 'Kubernetes', link: '/deployment/kubernetes.html' },
          ],
        },
        {
          text: 'Contributing',
          items: [
            { text: 'Contributing Guide', link: '/contributing.html' },
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
      copyright: 'Copyright © 2026-present desyang-hub',
    },
  },

  markdown: {
    config(md) {
      md.linkify.set({ fuzzyEmail: false })
    },
  },
})