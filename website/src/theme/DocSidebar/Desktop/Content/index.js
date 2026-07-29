/**
 * Ejected (swizzled) from @docusaurus/theme-classic 3.10.2's
 * `theme/DocSidebar/Desktop/Content/index.js`. Re-check this file against the
 * upstream component whenever Docusaurus is upgraded to a new major version -
 * the upstream source lives at
 * node_modules/@docusaurus/theme-classic/lib/theme/DocSidebar/Desktop/Content/index.js
 * inside `website/`.
 *
 * The only addition over upstream is a filter input above the `<ul>` that
 * narrows the sidebar tree to items whose label matches the query, so the
 * 118-entry Segments category can be searched without leaving the page.
 */
import React, {useState, useMemo} from 'react';
import clsx from 'clsx';
import {ThemeClassNames} from '@docusaurus/theme-common';
import {
  useAnnouncementBar,
  useScrollPosition,
} from '@docusaurus/theme-common/internal';
import {translate} from '@docusaurus/Translate';
import DocSidebarItems from '@theme/DocSidebarItems';
import styles from './styles.module.css';

function useShowAnnouncementBar() {
  const {isActive} = useAnnouncementBar();
  const [showAnnouncementBar, setShowAnnouncementBar] = useState(isActive);
  useScrollPosition(
    ({scrollY}) => {
      if (isActive) {
        setShowAnnouncementBar(scrollY === 0);
      }
    },
    [isActive],
  );
  return isActive && showAnnouncementBar;
}

/**
 * Recursively narrows a sidebar tree to items whose label matches `query`
 * (an already-lowercased, already-trimmed substring). A category is kept if
 * its own label matches (in which case it's kept whole, with its original
 * children) or if any descendant matches (in which case only the matching
 * descendants are kept). Every kept category is forced open (`collapsed:
 * false`) so a match is never left hidden inside a collapsed category. Items
 * with no `label` (e.g. raw `html` sidebar items) have nothing to match
 * against and are always kept as-is. The input tree is never mutated - kept
 * categories are shallow copies.
 */
/**
 * What a sidebar entry can be found by. The label alone is not enough: a
 * segment's sidebar label is its display name, but the thing a reader knows -
 * and the thing they type into their config - is its `type`, which lives in
 * the doc id. The kubectl segment is labelled "Kubernetes", so searching
 * "kubectl", the only string that appears in a config, found nothing at all.
 * The last path component of the doc id is that type, so match on both.
 */
function searchableText(item) {
  const label = item.label ?? '';
  const id = item.docId ?? item.id ?? '';
  const type = id.slice(id.lastIndexOf('/') + 1);

  return `${label} ${type}`.toLowerCase();
}

function filterSidebarItems(items, query) {
  return items.reduce((acc, item) => {
    if (item.type === 'category') {
      const labelMatches = searchableText(item).includes(query);
      if (labelMatches) {
        acc.push({...item, collapsed: false});
        return acc;
      }
      const filteredChildren = filterSidebarItems(item.items, query);
      if (filteredChildren.length > 0) {
        acc.push({...item, items: filteredChildren, collapsed: false});
      }
      return acc;
    }
    if (item.type === 'link') {
      if (searchableText(item).includes(query)) {
        acc.push(item);
      }
      return acc;
    }
    // 'html' (and any other future item type) has no label to filter on.
    acc.push(item);
    return acc;
  }, []);
}

export default function DocSidebarDesktopContent({path, sidebar, className}) {
  const showAnnouncementBar = useShowAnnouncementBar();
  const [query, setQuery] = useState('');
  const trimmedQuery = query.trim().toLowerCase();
  const hasQuery = trimmedQuery.length > 0;

  // Recomputed on every keystroke; SSR never sees a non-empty query since
  // `query` starts at '' and nothing browser-only feeds it.
  const visibleSidebar = useMemo(
    () => (hasQuery ? filterSidebarItems(sidebar, trimmedQuery) : sidebar),
    [sidebar, hasQuery, trimmedQuery],
  );
  const hasResults = visibleSidebar.length > 0;

  return (
    <nav
      aria-label={translate({
        id: 'theme.docs.sidebar.navAriaLabel',
        message: 'Docs sidebar',
        description: 'The ARIA label for the sidebar navigation',
      })}
      className={clsx(
        'menu thin-scrollbar',
        styles.menu,
        showAnnouncementBar && styles.menuWithAnnouncementBar,
        className,
      )}>
      <label className={styles.srOnly} htmlFor="doc-sidebar-filter-input">
        {translate({
          id: 'theme.docs.sidebar.filterInputLabel',
          message: 'Filter sidebar items',
          description:
            'The accessible label for the docs sidebar filter input',
        })}
      </label>
      <input
        id="doc-sidebar-filter-input"
        // type="text", not "search": the latter gets native browser search
        // styling (rounded pill shape, built-in clear button) that doesn't
        // match SegmentCatalog's own `.search` input, which this is styled
        // after (see styles.module.css's .filterInput comment).
        type="text"
        className={styles.filterInput}
        placeholder={translate({
          id: 'theme.docs.sidebar.filterInputPlaceholder',
          message: 'Filter...',
          description: 'The placeholder for the docs sidebar filter input',
        })}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      {hasQuery && !hasResults ? (
        <p className={styles.noResults}>
          {translate({
            id: 'theme.docs.sidebar.filterNoResults',
            message: 'No matching items.',
            description: 'Shown when the docs sidebar filter matches nothing',
          })}
        </p>
      ) : (
        <ul
          // Remounts the whole tree whenever filtering turns on/off (but not
          // on every keystroke while it stays on), so freshly-matched
          // categories always mount with their forced-open `collapsed` value
          // instead of inheriting a stale open/closed state from a sidebar
          // category component instance reused across renders (categories
          // are keyed by index in upstream `DocSidebarItems`, and `collapsed`
          // is only read once, at mount, via a lazy `useState` initializer).
          // Clearing the query flips the key back and remounts with the
          // original, untouched `sidebar`, restoring original collapsed
          // states.
          key={hasQuery ? 'filtered' : 'full'}
          className={clsx(ThemeClassNames.docs.docSidebarMenu, 'menu__list')}>
          <DocSidebarItems items={visibleSidebar} activePath={path} level={1} />
        </ul>
      )}
    </nav>
  );
}
