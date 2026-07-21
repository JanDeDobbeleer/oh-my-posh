import Link from "@docusaurus/Link";
import useBaseUrl from "@docusaurus/useBaseUrl";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import Head from "@docusaurus/Head";
import classnames from "classnames";
import { useState } from "react";
import { motion, useMotionValue, useMotionTemplate } from "framer-motion";
import CubeScene from "../components/CubeScene";
import FlickeringGrid from "../components/FlickeringGrid";
import styles from "./styles.module.css";

const fadeUp = (delay) => ({
  initial: { opacity: 0, y: 24, filter: "blur(10px)" },
  animate: { opacity: 1, y: 0, filter: "blur(0px)" },
  transition: { duration: 1.1, ease: [0.22, 1, 0.36, 1], delay },
});

function BoxIcon() {
  return (
    <svg
      className={styles.cardIcon}
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z" />
      <path d="m3.3 7 8.7 5 8.7-5" />
      <path d="M12 22V12" />
    </svg>
  );
}

const features = [
  {
    title: <>Beautiful &amp; Intelligent</>,
    description: (
      <>
        Transform your terminal with stunning themes and intelligent segments that display
        Git status, cloud info, language versions, system metrics, and 180+ other contextual details.
        Your prompt adapts to what you're working on.
      </>
    ),
  },
  {
    title: <>Lightning Fast</>,
    description: (
      <>
        Built with Go for blazing performance. Smart caching and async operations ensure
        your prompt renders instantly, even with complex configurations and multiple segments.
        No more waiting for your terminal.
      </>
    ),
  },
  {
    title: <>Universal Compatibility</>,
    description: (
      <>
        One configuration works everywhere - PowerShell, Bash, Zsh, Fish, Nu Shell, and more.
        Windows, macOS, Linux, WSL, containers, SSH sessions. Write once, use everywhere
        with zero vendor lock-in.
      </>
    ),
  },
];

function Feature({ title, description }) {
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const [isHovering, setIsHovering] = useState(false);
  const mask = useMotionTemplate`radial-gradient(350px circle at ${mouseX}px ${mouseY}px, white, transparent 80%)`;

  function handleMouseMove({ currentTarget, clientX, clientY }) {
    const { left, top } = currentTarget.getBoundingClientRect();
    mouseX.set(clientX - left);
    mouseY.set(clientY - top);
  }

  return (
    <div className={classnames("col col--4", styles.feature)}>
      <div
        className={styles.featureCard}
        onMouseMove={handleMouseMove}
        onMouseEnter={() => setIsHovering(true)}
        onMouseLeave={() => setIsHovering(false)}
      >
        <motion.div
          className={styles.spotlight}
          style={{ maskImage: mask, WebkitMaskImage: mask }}
        >
          {isHovering && (
            <FlickeringGrid
              className={styles.spotlightGrid}
              squareSize={2}
              gridGap={1}
              colors={["#3b82f6", "#6366f1", "#8b5cf6", "#1e40af"]}
              maxOpacity={0.85}
              flickerChance={0.35}
            />
          )}
        </motion.div>
        <BoxIcon />
        <h3 className={styles.featureTitle}>{title}</h3>
        <p className={styles.featureDescription}>{description}</p>
      </div>
    </div>
  );
}

function Home() {
  const context = useDocusaurusContext();
  const { siteConfig = {} } = context;

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
    <Layout title="Home" description={`${siteConfig.tagline}`}>
      <Head>
        <script type="application/ld+json" dangerouslySetInnerHTML={{__html: websiteJsonLd}} />
        <script type="application/ld+json" dangerouslySetInnerHTML={{__html: organizationJsonLd}} />
      </Head>
      <header className={classnames("hero", styles.heroBanner)}>
        <div className={styles.heroScene}>
          <CubeScene />
        </div>
        <div className={classnames("container", styles.heroContent)}>
          <motion.p className={styles.eyebrow} {...fadeUp(0.1)}>
            Open source &middot; Cross-shell prompt engine
          </motion.p>
          <motion.h1 className={styles.heroTitle} {...fadeUp(0.25)}>
            {siteConfig.title}
          </motion.h1>
          <motion.p className={styles.heroSubtitle} {...fadeUp(0.4)}>
            {siteConfig.tagline}
          </motion.p>
          <motion.div className={styles.buttons} {...fadeUp(0.55)}>
            <Link className={styles.btnPrimary} to={useBaseUrl("docs/")}>
              Get Started
            </Link>
            <Link className={styles.btnGhost} to={useBaseUrl("docs/themes")}>
              Browse Themes &rarr;
            </Link>
          </motion.div>
          <motion.img
            className={styles.heroImage}
            src="https://res.cloudinary.com/dakrfj1oh/image/upload/v1784611431/s_j1e5sd.png"
            alt="Oh My Posh prompt"
            {...fadeUp(0.7)}
          />
        </div>
      </header>
      <main>
        {features && features.length > 0 && (
          <section className={styles.features}>
            <div className="container">
              <div className={styles.sectionHead}>
                <p className={styles.eyebrow}>Why Oh My Posh</p>
                <h2 className={styles.sectionTitle}>Built for the modern terminal.</h2>
              </div>
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
