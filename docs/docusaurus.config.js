const path = require('path');

module.exports = {
  title: 'Oh My Posh',
  tagline: 'A prompt theme engine for any shell.',
  url: 'https://ohmyposh.dev',
  baseUrl: '/',
  favicon: 'img/favicon.ico',
  organizationName: 'jandedobbeleer',
  projectName: 'oh-my-posh',
  onBrokenLinks: 'ignore',
  plugins: [
    path.resolve(__dirname, 'plugins', 'appinsights'),
    'docusaurus-node-polyfills'
  ],
  themeConfig: {
    prism: {
      additionalLanguages: ['powershell', 'lua', 'jsstacktrace'],
    },
    navbar: {
      title: 'Oh My Posh',
      logo: {
        alt: 'Oh My Posh Logo',
        src: 'img/logo.svg',
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
          href: 'https://www.gitkraken.com/invite/nQmDPR9D',
          className: 'header-gk-link',
          'aria-label': 'GitKraken',
          position: 'right',
        },
        {
          href: 'https://twitter.com/jandedobbeleer',
          className: 'header-twitter-link',
          'aria-label': 'Twitter',
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
              to: 'docs/contributing_started',
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
              label: 'Twitter',
              href: 'https://twitter.com/jandedobbeleer',
            },
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
              label: 'GitKraken',
              href: 'https://www.gitkraken.com/invite/nQmDPR9D',
            },
            {
              label: 'Docusaurus',
              href: 'https://github.com/facebook/docusaurus',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href='https://github.com/sponsors/JanDeDobbeleer' target='_blank'>Jan De Dobbeleer</a> and <a href='/docs/contributors'>contributors</a>.`,
    },
    appInsights: {
      instrumentationKey: '72804848-dc30-4856-8245-4fa1450b041f',
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
          editUrl: 'https://github.com/jandedobbeleer/oh-my-posh/edit/main/docs/',
        },
        theme: {
          customCss: [
            require.resolve('./src/css/custom.css'),
            require.resolve('./src/css/prism-rose-pine-moon.css')
          ],
        },
      },
    ],
  ],
};
