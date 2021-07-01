module.exports = {
  title: "Oh My Posh",
  tagline: "A prompt theme engine for any shell.",
  url: "https://ohmyposh.dev",
  baseUrl: "/",
  favicon: "img/favicon.ico",
  organizationName: "jandedobbeleer",
  projectName: "oh-my-posh",
  onBrokenLinks: "ignore",
  themeConfig: {
    sidebarCollapsible: false,
    prism: {
      theme: require("prism-react-renderer/themes/duotoneLight"),
      darkTheme: require("prism-react-renderer/themes/oceanicNext"),
    },
    navbar: {
      title: "Oh My Posh",
      logo: {
        alt: "Oh My Posh Logo",
        src: "img/logo.svg",
      },
      items: [
        {
          to: "docs/",
          activeBasePath: "docs",
          label: "Docs",
          position: "left",
        },
        {
          href: "https://github.com/sponsors/JanDeDobbeleer",
          label: "Sponsor",
          position: "left",
        },
        {
          href: "https://www.gitkraken.com/invite/nQmDPR9D",
          label: "GitKraken",
          position: "left",
        },
        {
          href: "https://github.com/jandedobbeleer/oh-my-posh",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "How to",
          items: [
            {
              label: "Getting started",
              to: "docs/",
            },
            {
              label: "Contributing",
              to: "docs/contributing_started",
            },
          ],
        },
        {
          title: "Social",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/jandedobbeleer/oh-my-posh",
            },
            {
              label: "Twitter",
              href: "https://twitter.com/jandedobbeleer",
            },
          ],
        },
        {
          title: "Links",
          items: [
            {
              label: "Sponsor",
              href: "https://github.com/sponsors/JanDeDobbeleer",
            },
            {
              label: "GitKraken",
              href: "https://www.gitkraken.com/invite/nQmDPR9D",
            },
            {
              label: "Docusaurus",
              href: "https://github.com/facebook/docusaurus",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href="https://github.com/sponsors/JanDeDobbeleer" target="_blank">Jan De Dobbeleer</a> and <a href="/docs/contributors">contributors</a>.`,
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/jandedobbeleer/oh-my-posh/edit/main/docs/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
