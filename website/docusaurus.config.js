import {themes as prismThemes} from 'prism-react-renderer';

export default {
  title: 'Oh My Posh',
  tagline: 'The most customizable and fastest prompt engine for any shell.',
  url: 'https://ohmyposh.dev',
  baseUrl: '/',
  favicon: 'img/favicons.svg',
  organizationName: 'jandedobbeleer',
  projectName: 'oh-my-posh',
  onBrokenLinks: 'throw',
  plugins: [
    './plugins/appinsights',
    './plugins/segments',
    './plugins/themes'
  ],
  headTags: [
    {
      tagName: 'link',
      attributes: {
        rel: 'alternate',
        type: 'text/plain',
        title: 'llms.txt',
        href: '/llms.txt',
      },
    },
  ],
  stylesheets: [
    "https://rsms.me/inter/inter.css"
  ],
  themeConfig: {
    metadata: [
      // A raster, unlike everything else the site shows: link previews will not render an SVG.
      // Composed from the same rendered default prompt the homepage inlines, by
      // scripts/render-og-image.mjs - which is run by hand, not by the build, since it needs a
      // headless browser to pick up the real font. Re-run it when the default config changes.
      {property: 'og:image', content: 'https://ohmyposh.dev/img/og-image.png'},
      {property: 'og:image:width', content: '1200'},
      {property: 'og:image:height', content: '630'},
      {property: 'og:type', content: 'website'},
      // Without this the card renders as a thumbnail beside the text rather than full width.
      {name: 'twitter:card', content: 'summary_large_image'},
      {name: 'twitter:image', content: 'https://ohmyposh.dev/img/og-image.png'},
    ],
    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      // false, so the toggle is light <-> dark and nothing else. With it on, Docusaurus adds a
      // third "system" step to the cycle that looks identical to whichever of the two the OS is
      // already set to - a click that appears to do nothing.
      respectPrefersColorScheme: false,
    },
    prism: {
      // Matches ConfigEditor's own DARK_CODE_THEME/LIGHT_CODE_THEME pair (see
      // src/components/ConfigEditor/index.js) so every fenced code block in the docs follows
      // the site's light/dark switch the same way the config editor does, instead of staying on
      // one theme regardless of colour mode.
      theme: prismThemes.github,
      darkTheme: prismThemes.palenight,
      additionalLanguages: ['powershell', 'lua', 'jsstacktrace', 'toml'],
    },
    docs: {
        sidebar: {
          hideable: false,
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
          to: 'docs/segments/overview',
          label: 'Segments',
          position: 'left',
        },
        {
          to: 'docs/themes',
          label: 'Themes',
          position: 'left',
        },
        {
          to: 'docs/studio',
          label: 'Studio',
          position: 'left',
        },
        {
          to: 'blog',
          label: 'Blog',
          position: 'left'
        },
        {
          href: 'https://github.com/jandedobbeleer/oh-my-posh',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
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
      // No `style: 'dark'`: that adds .footer--dark, which pins the footer to one dark palette
      // in both colour modes, so in light mode the page ended on a slab that matched nothing
      // else on it. The footer now follows the navbar in whichever mode is active - see .footer
      // in src/css/custom.css.
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
              label: 'Docusaurus',
              href: 'https://github.com/facebook/docusaurus',
            },
            {
              label: 'Privacy',
              href: '/privacy',
            },
          ],
        },
        {
          title: 'Support',
          items: [
            {
              label: 'GitHub Sponsors',
              href: 'https://github.com/sponsors/JanDeDobbeleer',
            },
            {
              label: 'Product spotlight',
              href: 'https://buy.polar.sh/polar_cl_qnmZxboq1IDUJo03mk2Jue6ktqZrCXElnzH2s2xbV2R',
            },
            {
              label: 'Swag',
              href: 'https://swag.ohmyposh.dev',
            },
          ],
        },
        {
          title: 'Our sponsors',
          items: [
            {
              label: 'CodeRabbit',
              href: 'https://coderabbit.link/posh',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href='https://github.com/sponsors/JanDeDobbeleer' target='_blank'>Jan De Dobbeleer</a> and <a href='/docs/contributors'>contributors</a>.`,
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
          sidebarPath: './sidebars.js',
          editUrl: 'https://github.com/jandedobbeleer/oh-my-posh/edit/main/website/',
        },
        theme: {
          customCss: [
            './src/css/custom.css'
          ],
        },
        blog: {
          onInlineAuthors: 'ignore'
        },
      },
    ],
  ],
};
