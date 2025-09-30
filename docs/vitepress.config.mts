import {
  defineConfig,
} from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "pawbar",
  description: "Kat vibes for your desktop",
  srcExclude: ['**/README.md'],
  cleanUrls: true,
  head: [
    [
      'link',
      { rel: 'icon', type: 'image/svg+xml', href: '/pawbar.svg' }
    ],
    [
      'link',
      {
        rel: 'stylesheet',
        href: 'https://fonts.googleapis.com/css2?family=Karla:wght@400;500;700&family=Source+Code+Pro:wght@400;500;700&display=swap'
      }
    ]
  ],
  themeConfig: {
    logo: '/pawbar.svg',
    outline: [2, 3],
    search: { provider: 'local' },
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Docs', link: '/docs/getting-started', activeMatch: "/docs/" },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        link: '/docs/getting-started',
      },
      {
        text: 'Modules',
        collapsed: false,
        base: '/docs/modules/',
        items: [
          { text: 'Clock', link: 'clock' },
          { text: '' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/codelif/pawbar' }
    ],

    editLink: {
      pattern: 'https://github.com/codelif/pawbar/edit/main/docs/:path',
      text: 'Edit this page on Github'
    },

    footer: {
      message: 'Released under the BSD-3-Clause License.',
      copyright: 'Copyright © 2025 Harsh Sharma'
    }
  }
})



