import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import classnames from "classnames";
import React from "react";
import styles from "./styles.module.css";

const features = [
  {
    title: <>🎨 Beautiful & Intelligent</>,
    description: (
      <>
        Transform your terminal with stunning themes and intelligent segments that display
        Git status, cloud info, language versions, system metrics, and 180+ other contextual details.
        Your prompt adapts to what you're working on.
      </>
    ),
  },
  {
    title: <>⚡ Lightning Fast</>,
    description: (
      <>
        Built with Go for blazing performance. Smart caching and async operations ensure
        your prompt renders instantly, even with complex configurations and multiple segments.
        No more waiting for your terminal.
      </>
    ),
  },
  {
    title: <>🌍 Universal Compatibility</>,
    description: (
      <>
        One configuration works everywhere - PowerShell, Bash, Zsh, Fish, Nu Shell, and more.
        Windows, macOS, Linux, WSL, containers, SSH sessions. Write once, use everywhere
        with zero vendor lock-in.
      </>
    ),
  },
];

function Feature({ imageUrl, title, description }) {
  const imgUrl = useBaseUrl(imageUrl);
  return (
    <div className={classnames("col col--4", styles.feature)}>
      {imgUrl && (
        <div className="text--center">
          <img className={styles.featureImage} src={imgUrl} alt={title} />
        </div>
      )}
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;
  return (
    <Layout title="Home" description={`${siteConfig.tagline}`}>
      <header className={classnames("hero hero--primary", styles.heroBanner)}>
        <div className="container">
          <h1 className="hero__title">{siteConfig.title}</h1>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link
              className={classnames(
                "button button--primary button--lg",
                styles.getStarted
              )}
              to={useBaseUrl("docs/")}
            >
              Get Started &rarr;
            </Link>
            <Link
              className={classnames(
                "button button--outline button--lg",
                styles.getStarted
              )}
              to={useBaseUrl("docs/themes")}
            >
              See themes &rarr;
            </Link>
          </div>
          <img className="hero--image" src="/img/hero.png" alt="Oh My Posh prompt"></img>
        </div>
      </header>
      <main>
        {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container">
              <div className="row">
                {features.map((props, idx) => (
                  <Feature key={idx} {...props} />
                ))}
              </div>
            </div>
          </section>
        )}
      </main>
    </Layout>
  );
}

export default Home;
