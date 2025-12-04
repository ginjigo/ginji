import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Ginji',
  description: 'Go Framework for Modern Web Services',
  base: '/ginji/',
  
  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API Reference', link: '/api/' },
      { text: 'Examples', link: '/examples/' }
    ],

    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'What is Ginji?', link: '/guide/introduction' },
          { text: 'Getting Started', link: '/guide/getting-started' },
          { text: 'Quick Start', link: '/guide/quick-start' }
        ]
      },
      {
        text: 'Core Concepts',
        items: [
          { text: 'Routing', link: '/guide/routing' },
          { text: 'Context', link: '/guide/context' },
          { text: 'Middleware', link: '/guide/middleware' },
          { text: 'Error Handling', link: '/guide/error-handling' },
          { text: 'Validation', link: '/guide/validation' }
        ]
      },
      {
        text: 'Middleware',
        items: [
          { text: 'Overview', link: '/middleware/' },
          { text: 'Body Limit', link: '/middleware/body-limit' },
          { text: 'Security Headers', link: '/middleware/security' },
          { text: 'Health Checks', link: '/middleware/health' },
          { text: 'Rate Limiting', link: '/middleware/rate-limit' },
          { text: 'Authentication', link: '/middleware/auth' },
          { text: 'Timeout', link: '/middleware/timeout' }
        ]
      },
      {
        text: 'Real-Time',
        items: [
          { text: 'WebSocket', link: '/realtime/websocket' },
          { text: 'Server-Sent Events', link: '/realtime/sse' },
          { text: 'Streaming', link: '/realtime/streaming' }
        ]
      },
      {
        text: 'Advanced',
        items: [
          { text: 'Testing', link: '/advanced/testing' },
          { text: 'Performance', link: '/advanced/performance' },
          { text: 'Best Practices', link: '/advanced/best-practices' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/ginjigo/ginji' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present Ginji'
    },

    search: {
      provider: 'local'
    }
  }
})