import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Lightning Multitool",
  description: "Easy-to-use multitool for bitcoin lightning node operators",
  ignoreDeadLinks: true,

  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Features', link: '/features/introduction' }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' }
          ]
        }
      ],
      '/features/': [
        {
          text: 'Features',
          items: [
            { text: 'Introduction', link: '/features/introduction' },
            { text: 'Lightning Address', link: '/features/lightning-address' },
            { text: 'Nostr Support', link: '/features/nostr' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/asheswook/lnurl' }
    ]
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en'
    },
    ko: {
      label: '한국어',
      lang: 'ko',
      link: '/ko/',

      themeConfig: {
        nav: [
          { text: '홈', link: '/ko/' },
          { text: '가이드', link: '/ko/guide/getting-started' },
          { text: '기능', link: '/ko/features/introduction' }
        ],

        sidebar: {
          '/ko/guide/': [
            {
              text: '가이드',
              items: [
                { text: '시작하기', link: '/ko/guide/getting-started' },
                { text: '설치', link: '/ko/guide/installation' },
                { text: '설정', link: '/ko/guide/configuration' }
              ]
            }
          ],
          '/ko/features/': [
            {
              text: '기능',
              items: [
                { text: '소개', link: '/ko/features/introduction' },
                { text: '라이트닝 주소', link: '/ko/features/lightning-address' },
                { text: 'Nostr 지원', link: '/ko/features/nostr' }
              ]
            }
          ]
        }
      }
    }
  }
})
