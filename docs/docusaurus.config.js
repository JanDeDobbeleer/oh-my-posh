module.exports = {
  title: "Oh my Posh",
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
      title: "Oh my Posh",
      logo: {
        alt: "Oh my Posh Logo",
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
              label: "Support",
              href: "/docs/#-support-",
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
      copyright: `Copyright © ${new Date().getFullYear()} Jan De Dobbeleer.`,
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
