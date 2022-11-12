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
  stylesheets: [
    "https://rsms.me/inter/inter.css",
    "https://fonts.googleapis.com/css2?family=Fira+Code&display=swap"
  ],
  themeConfig: {
    prism: {
      additionalLanguages: ['powershell', 'lua', 'jsstacktrace', 'toml'],
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
          href: 'https://social.ohmyposh.dev/@jan',
          className: 'header-mastodon-link',
          'aria-label': 'Mastodon',
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
              html: "<a class=\"footer__link-item\" rel=\"me\" href=\"https://social.ohmyposh.dev/@jan\">Mastodon <svg width=\"13.5\" height=\"13.5\" aria-hidden=\"true\" viewBox=\"0 0 24 24\" class=\"iconExternalLink_node_modules-@docusaurus-theme-classic-lib-theme-Icon-ExternalLink-styles-module\"><path fill=\"currentColor\" d=\"M21 13v10h-21v-19h12v2h-10v15h17v-8h2zm3-12h-10.988l4.035 4-6.977 7.07 2.828 2.828 6.977-7.07 4.125 4.172v-11z\"></path></svg></a>",
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
          editUrl: 'https://github.com/jandedobbeleer/oh-my-posh/edit/main/website/',
        },
        theme: {
          customCss: [
            require.resolve('./src/css/prism-rose-pine-moon.css'),
            require.resolve('./src/css/custom.css')
          ],
        },
      },
    ],
  ],
};
