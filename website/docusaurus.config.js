const path = require('path');

module.exports = {
  title: 'Oh My Posh',
  tagline: 'A prompt theme engine for any shell.',
  url: 'https://ohmyposh.dev',
  baseUrl: '/',
  favicon: 'img/favicons.svg',
  organizationName: 'jandedobbeleer',
  projectName: 'oh-my-posh',
  onBrokenLinks: 'ignore',
  plugins: [
    path.resolve(__dirname, 'plugins', 'appinsights')
  ],
  stylesheets: [
    "https://rsms.me/inter/inter.css",
    "https://fonts.googleapis.com/css2?family=Fira+Code&display=swap"
  ],
  themeConfig: {
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    prism: {
      additionalLanguages: ['powershell', 'lua', 'jsstacktrace', 'toml', 'json', 'yaml'],
    },
    docs: {
        sidebar: {
          hideable: true,
        },
    },
    navbar: {
      title: 'Oh My Posh',
      logo: {
        alt: 'Oh My Posh Logo',
        src: 'img/logo-dark.svg',
        srcDark: 'img/logo-light.svg',
      },
      items: [
        {
          to: 'docs',
          activeBasePath: 'docs',
          label: 'Docs',
          position: 'left',
        },
        {
          to: 'blog',
          label: 'Blog',
          position: 'left'
        },
        {
          href: 'https://github.com/sponsors/JanDeDobbeleer',
          label: 'Sponsor',
          position: 'left',
        },
        {
          href: 'https://polar.sh/oh-my-posh',
          label: 'Buy',
          position: 'left',
        },
        {
          href: 'https://swag.ohmyposh.dev',
          label: 'Swag',
          position: 'left',
        },
        {
          href: 'https://github.com/jandedobbeleer/oh-my-posh',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
          position: 'right',
        },
        {
          href: 'https://www.warp.dev/oh-my-posh',
          className: 'header-affiliate-link',
          'aria-label': 'Warp',
          position: 'right',
        },
        {
          href: 'https://discord.gg/n7E3DkXssv',
          className: 'header-discord-link',
          'aria-label': 'Discord',
          position: 'right',
        },
        {
          href: 'https://staging.bsky.app/profile/ohmyposh.dev',
          className: 'header-bluesky-link',
          'aria-label': 'Bluesky',
          position: 'right',
        }
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'How to',
          items: [
            {
              label: 'Getting started',
              to: 'docs/',
            },
            {
              label: 'Contributing',
              to: 'docs/contributing/started',
            },
          ],
        },
        {
          title: 'Social',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/jandedobbeleer/oh-my-posh',
            },
            {
              label: 'Discord',
              href: 'https://discord.gg/n7E3DkXssv',
            },
            {
              label: 'Bluesky',
              href: 'https://staging.bsky.app/profile/ohmyposh.dev',
            }
          ],
        },
        {
          title: 'Links',
          items: [
            {
              label: 'Sponsor',
              href: 'https://github.com/sponsors/JanDeDobbeleer',
            },
            {
              label: 'Buy',
              href: 'https://polar.sh/oh-my-posh',
            },
            {
              label: 'Warp',
              href: 'https://www.warp.dev/oh-my-posh',
            },
            {
              label: 'Docusaurus',
              href: 'https://github.com/facebook/docusaurus',
            },
            {
              label: 'Privacy',
              href: '/privacy',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href='https://github.com/sponsors/JanDeDobbeleer' target='_blank'>Jan De Dobbeleer</a> and <a href='/docs/contributors'>contributors</a>.`,
    },
    announcementBar: {
      id: 'support_us',
      content:
        'If you\'re enjoying Oh My Posh, consider becoming a <a target="_blank" rel="noopener noreferrer" href="https://github.com/sponsors/JanDeDobbeleer">sponsor</a> to keep the project going strong 💪',
      backgroundColor: '#2c7ae0',
      textColor: '#ffffff',
      isCloseable: false,
    },
    appInsights: {
      instrumentationKey: '51741aa7-e087-4e80-b7b0-0863d467462a',
    },
    algolia: {
      appId: 'XIR4RB3TM1',
      apiKey: '15c5f4340520612ed98fe21d15882029',
      indexName: 'ohmyposh',
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/jandedobbeleer/oh-my-posh/edit/main/website/',
        },
        theme: {
          customCss: [
            require.resolve('./src/css/prism-rose-pine-moon.css'),
            require.resolve('./src/css/custom.css')
          ],
        },
        blog: {
          onInlineAuthors: 'ignore'
        },
      },
    ],
  ],
};
