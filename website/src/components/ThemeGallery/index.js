import React, { useCallback, useState } from 'react';
import Link from '@docusaurus/Link';
import { useHistory } from '@docusaurus/router';
import themes from '../../../generated/themes.json';
import { parseConfig } from '../ConfigEditor/serialize';
import { LOAD_KEY, trySessionStorageSet } from '../ConfigEditor/studioHandoff';
import styles from './styles.module.css';

// Fetches a theme's actual config from GitHub (rawConfigUrl - see export_themes.mjs's own
// comment on why the manifest carries a URL rather than the file content) and, on success, hands
// it to the studio via LOAD_KEY before navigating there - mirroring Config.js's "Add to Studio"
// but replacing the whole editor instead of appending one segment (see studioHandoff.js).
//
// The fetch happens *before* navigating, not after, so a network failure - GitHub down, offline,
// blocked by an extension/firewall - leaves the reader on this page with a clear error message
// and the existing GitHub link as a fallback, rather than dropping them onto a studio that either
// silently kept its previous contents or shows nothing.
function OpenInStudioButton({ name, rawConfigUrl, format }) {
  const history = useHistory();
  const [status, setStatus] = useState('idle'); // 'idle' | 'loading' | 'error'

  const handleClick = useCallback(async () => {
    setStatus('loading');

    try {
      const response = await fetch(rawConfigUrl);

      if (!response.ok) {
        throw new Error(`request failed with status ${response.status}`);
      }

      const text = await response.text();

      // Guards against a "successful" fetch that isn't actually the config - GitHub serving an
      // HTML error/rate-limit page with a 200, for instance. A config that doesn't parse must
      // not be handed to the studio as though it were real.
      parseConfig(format, text);

      if (!trySessionStorageSet(LOAD_KEY, JSON.stringify({ format, text }))) {
        throw new Error('sessionStorage is unavailable');
      }

      history.push('/docs/studio');
    } catch {
      setStatus('error');
    }
  }, [rawConfigUrl, format, history]);

  return (
    <div className={styles.studioAction}>
      <button
        type="button"
        className={styles.studioButton}
        onClick={handleClick}
        disabled={status === 'loading'}
      >
        {status === 'loading' ? 'Loading…' : 'Open in Studio'}
      </button>

      {status === 'error' && (
        <p className={styles.studioError}>
          Couldn&apos;t fetch {name}&apos;s config from GitHub. Try again, or{' '}
          <Link to={rawConfigUrl}>view it directly</Link>.
        </p>
      )}
    </div>
  );
}

// ThemeCard inlines one theme's SVG verbatim via dangerouslySetInnerHTML - the
// markup never passes through MDX/JSX, so it can't be mangled by the MDX
// compiler and it doesn't need to be escaped/quoted for JSX attribute rules.
// It renders inline (not via <img>), which is the entire point: an inlined
// SVG inherits the page's own @font-face (see custom.css's "Victor Mono"
// declaration), so the icons and powerline glyphs render without any font
// embedding, subsetting, or data URI of their own.
function ThemeCard({ name, githubUrl, rawConfigUrl, format, svg, svgLight }) {
  return (
    <div className={styles.card}>
      <div className={styles.headingRow}>
        <h3 className={styles.heading} id={name.toLowerCase()}>
          <Link to={githubUrl}>{name}</Link>
        </h3>
        <OpenInStudioButton name={name} rawConfigUrl={rawConfigUrl} format={format} />
      </div>
      <Link to={githubUrl} className={styles.render}>
        {/* Two renders (dark svg, light svgLight - see export_themes.mjs's LIGHT_BACKGROUND)
            ship in the static HTML; custom.css's shared omp-light-only/omp-dark-only classes
            pick between them off Docusaurus's own html[data-theme] attribute, the same instant
            CSS-only switch the homepage hero uses (src/pages/index.js). */}
        <span
          className={`${styles.svgWrapper} omp-dark-only`}
          dangerouslySetInnerHTML={{ __html: svg }}
        />
        <span
          className={`${styles.svgWrapper} omp-light-only`}
          dangerouslySetInnerHTML={{ __html: svgLight }}
        />
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
        <ThemeCard
          key={theme.name}
          name={theme.name}
          githubUrl={theme.githubUrl}
          rawConfigUrl={theme.rawConfigUrl}
          format={theme.format}
          svg={theme.svg}
          svgLight={theme.svgLight}
        />
      ))}
    </div>
  );
}

export default ThemeGallery;
