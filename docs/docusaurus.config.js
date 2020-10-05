module.exports = {
  title: "Oh my Posh 3",
  tagline: "A prompt theme engine for any shell.",
  url: "https://ohmyposh.dev",
  baseUrl: "/",
  favicon: "img/favicon.ico",
  organizationName: "jandedobbeleer",
  projectName: "oh-my-posh3",
  onBrokenLinks: "ignore",
  themeConfig: {
    sidebarCollapsible: false,
    prism: {
      theme: require("prism-react-renderer/themes/github"),
      darkTheme: require("prism-react-renderer/themes/dracula"),
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
          href: "https://github.com/jandedobbeleer/oh-my-posh3",
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
              label: "Packages",
              to: "docs/powershell",
            },
          ],
        },
        {
          title: "Social",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/jandedobbeleer/oh-my-posh3",
            },
            {
              label: "Twitter",
              href: "https://twitter.com/jandedobbeleer",
            },
          ],
        },
        {
          title: "Built with",
          items: [
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
          editUrl: "https://github.com/jandedobbeleer/oh-my-posh3/edit/master/docs/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
