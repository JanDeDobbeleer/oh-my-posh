import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import { usePluginData } from "@docusaurus/useGlobalData";
import Layout from "@theme/Layout";
import Head from "@docusaurus/Head";
import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";
import CodeBlock from "@theme/CodeBlock";
import classnames from "classnames";
// A plain static import, the same way <ThemeGallery/> pulls in the full manifest: webpack
// code-splits it into this page's own chunk, and being synchronous it resolves during the SSR
// build, so the static HTML ships with the prompt already inlined. This file holds only the
// default theme - see export_themes.mjs's HERO_FILE for why it is not the full manifest.
import hero from "../../generated/hero.json";
import styles from "./styles.module.css";

// The first two cards are fixed copy; the third's title is built from the segments plugin's
// computed total (see usePluginData below) instead of a hardcoded "117 segments" that would
// silently drift as segments are added or removed.
function buildFeatures(segmentTotal) {
  return [
    {
      title: <>One config, every shell</>,
      description: (
        <>
          Write your prompt once and run it in PowerShell, Bash, Zsh, Fish, Nu and more, on
          Windows, macOS, Linux, WSL, in containers and over SSH.
        </>
      ),
    },
    {
      title: <>Built for speed</>,
      description: (
        <>
          Written in Go. Segments run concurrently and cache what they find, so a prompt with a
          dozen of them still renders in milliseconds.
        </>
      ),
    },
    {
      title: <>{segmentTotal} segments</>,
      description: (
        <>
          Git status, cloud context, language versions, system metrics and more, composed with a
          template language so your prompt follows the work you are doing.
        </>
      ),
    },
  ];
}

const shells = [
  "bash",
  "cmd",
  "elvish",
  "fish",
  "nu",
  "powershell",
  "xonsh",
  "yash",
  "zsh",
];

function Feature({ title, description }) {
  return (
    <div className={classnames("col col--4", styles.feature)}>
      <div className={styles.card}>
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig } = context;
  const { total: segmentTotal } = usePluginData("oh-my-posh-segments");
  const features = buildFeatures(segmentTotal);

  const websiteJsonLd = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "WebSite",
    "url": "https://ohmyposh.dev/",
    "name": "Oh My Posh",
    "description": siteConfig.tagline,
  });

  const organizationJsonLd = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "Organization",
    "name": "Oh My Posh",
    "url": "https://ohmyposh.dev/",
    "logo": "https://ohmyposh.dev/img/logo.png",
    "description": siteConfig.tagline,
    "sameAs": [
      "https://github.com/JanDeDobbeleer/oh-my-posh",
    ],
  });

  return (
    <Layout title="Home" description={siteConfig.tagline}>
      <Head>
        <script type="application/ld+json" dangerouslySetInnerHTML={{__html: websiteJsonLd}} />
        <script type="application/ld+json" dangerouslySetInnerHTML={{__html: organizationJsonLd}} />
      </Head>
      <header className={styles.heroBanner}>
        <div className="container">
          <h1 className={classnames("hero__title", styles.heroTitle)}>{siteConfig.title}</h1>
          <p className={classnames("hero__subtitle", styles.subtitle)}>{siteConfig.tagline}</p>

          {/* The prompt oh-my-posh renders with no config of its own, drawn by the same SVG
              encoder the gallery and every segment doc use - not a screenshot, and not one of
              the bundled themes. It inherits the page's own @font-face, so the icons and
              powerline glyphs draw without embedding a font, and the text stays selectable. */}
          <div className={styles.heroPrompt}>
            <span
              className={styles.svgWrapper}
              role="img"
              aria-label="The prompt Oh My Posh renders out of the box"
              dangerouslySetInnerHTML={{ __html: hero.svg }}
            />
          </div>

          <div className={styles.installBox}>
            <Tabs
              groupId="install-os"
              defaultValue="windows"
              values={[
                { label: "Windows", value: "windows" },
                { label: "macOS", value: "macos" },
                { label: "Linux", value: "linux" },
              ]}
            >
              <TabItem value="windows">
                <CodeBlock language="powershell">
                  winget install JanDeDobbeleer.OhMyPosh --source winget
                </CodeBlock>
              </TabItem>
              <TabItem value="macos">
                <CodeBlock language="bash">
                  brew install jandedobbeleer/oh-my-posh/oh-my-posh
                </CodeBlock>
              </TabItem>
              <TabItem value="linux">
                <CodeBlock language="bash">
                  curl -s https://ohmyposh.dev/install.sh | bash -s
                </CodeBlock>
              </TabItem>
            </Tabs>
            <p className={styles.installNote}>
              Other options and package managers in the{" "}
              <Link to={useBaseUrl("docs/installation/windows")}>installation guide</Link>.
            </p>
          </div>

          <div className={styles.buttons}>
            <Link
              className="button button--primary button--lg"
              to={useBaseUrl("docs/")}
            >
              Get started &rarr;
            </Link>
            <Link
              className="button button--outline button--primary button--lg"
              to={useBaseUrl("docs/themes")}
            >
              Browse themes &rarr;
            </Link>
          </div>

        </div>
      </header>
      <main>
        <section className={styles.features}>
          <div className="container">
            <div className="row">
              {features.map((props, idx) => (
                <Feature key={idx} {...props} />
              ))}
            </div>
          </div>
        </section>
        <section className={styles.shellRow}>
          <div className="container">
            <p className={styles.shellList}>Works with {shells.join(" · ")}</p>
          </div>
        </section>
      </main>
    </Layout>
  );
}

export default Home;
