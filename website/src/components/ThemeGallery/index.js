import React from 'react';
import Link from '@docusaurus/Link';
import themes from '../../../generated/themes.json';
import styles from './styles.module.css';

// ThemeCard inlines one theme's SVG verbatim via dangerouslySetInnerHTML - the
// markup never passes through MDX/JSX, so it can't be mangled by the MDX
// compiler and it doesn't need to be escaped/quoted for JSX attribute rules.
// It renders inline (not via <img>), which is the entire point: an inlined
// SVG inherits the page's own @font-face (see custom.css's "Victor Mono"
// declaration), so the icons and powerline glyphs render without any font
// embedding, subsetting, or data URI of their own.
function ThemeCard({ name, githubUrl, svg }) {
  return (
    <div className={styles.card}>
      <h3 className={styles.heading} id={name.toLowerCase()}>
        <Link to={githubUrl}>{name}</Link>
      </h3>
      <Link to={githubUrl} className={styles.render}>
        <span className={styles.svgWrapper} dangerouslySetInnerHTML={{ __html: svg }} />
      </Link>
    </div>
  );
}

// A plain static import, deliberately not usePluginData/setGlobalData: that
// mechanism bundles a plugin's entire payload into Docusaurus's shared
// main.js runtime (fetched on every page, forever), because usePluginData
// has to be callable from anywhere. This JSON file, by contrast, is only
// ever reachable through this one module, which is only ever reachable
// through docs/themes.mdx - so webpack's own chunking (no special plugin
// machinery needed) keeps it inside that page's own already-code-split
// chunk instead of hoisting it into the shared bundle. Being a plain
// synchronous import (not a dynamic import()), there is nothing async to
// resolve at render time either, so ReactDOMServer's build-time SSR pass
// renders the real markup directly - no Loadable/preloadAll trick, no
// hydration-only content, no loading flash.
function ThemeGallery() {
  return (
    <div className={styles.gallery}>
      {themes.map((theme) => (
        <ThemeCard key={theme.name} name={theme.name} githubUrl={theme.githubUrl} svg={theme.svg} />
      ))}
    </div>
  );
}

export default ThemeGallery;
