import React, { useCallback, useMemo } from 'react';
import Link from '@docusaurus/Link';
import useIsBrowser from '@docusaurus/useIsBrowser';
import { useHistory, useLocation } from '@docusaurus/router';
import { usePluginData } from '@docusaurus/useGlobalData';
import classnames from 'classnames';
import styles from './styles.module.css';

// Filters live in the URL, never in component state. That is what keeps the
// server render (no filters applied) and the pre-hydration client render
// (useIsBrowser() still false) byte-for-byte identical - see the hydration
// note below in useFilters().
const EMPTY_FILTERS = { q: '' };

function parseFilters(search) {
  return { q: new URLSearchParams(search).get('q') || '' };
}

function buildSearch(filters) {
  return filters.q ? `?${new URLSearchParams({ q: filters.q })}` : '';
}

// Docusaurus renders <StaticRouter location={pathname}> on the server - no
// query string at all - and React 19 hydrates under BrowserRouter. Reading
// useLocation().search to seed the very first client render would produce a
// filtered list against the server's unfiltered list and blow away hydration.
// useIsBrowser() is false on the server AND on the first client render, so
// gating on it (rather than on window/document, which doesn't help here)
// keeps both renders identical. The filtered view appears one tick later,
// once useIsBrowser() flips to true.
function useFilters() {
  const isBrowser = useIsBrowser();
  const location = useLocation();
  const history = useHistory();

  const filters = useMemo(
    () => (isBrowser ? parseFilters(location.search) : EMPTY_FILTERS),
    [isBrowser, location.search],
  );

  const setFilters = useCallback(
    (next) => {
      history.replace({ search: buildSearch(next) });
    },
    [history],
  );

  return [filters, setFilters];
}

function matchesSearch(segment, q) {
  if (!q) {
    return true;
  }

  const needle = q.toLowerCase();
  return (
    segment.name.toLowerCase().includes(needle) ||
    segment.type.toLowerCase().includes(needle) ||
    segment.id.toLowerCase().includes(needle) ||
    segment.description.toLowerCase().includes(needle)
  );
}

function matchesAll(segment, filters) {
  return matchesSearch(segment, filters.q);
}

function normalize(value) {
  return value.toLowerCase().replace(/[\s_-]/g, '');
}

function showsSecondaryName(segment) {
  return normalize(segment.name) !== normalize(segment.type);
}

// href isn't part of the plugin payload (it's cheap to reconstruct from group + id, and shipping
// ~117 copies of it would just be dead weight in the shared chunk); build it here instead.
function segmentHref(segment) {
  return `/docs/segments/${segment.group}/${segment.id}`;
}

function SegmentCard({ segment, groupLabelById, authLabelById }) {
  return (
    <Link to={segmentHref(segment)} className={styles.item}>
      <span className={styles.nm}>{segment.type}</span>
      {showsSecondaryName(segment) && <span className={styles.name}>{segment.name}</span>}
      <span className={styles.ds}>{segment.description}</span>
      <div className={styles.tags}>
        <span className={styles.tag}>{groupLabelById[segment.group]}</span>
        {segment.auth && (
          <span
            className={classnames(styles.tag, styles.authTag)}
            title={authLabelById[segment.auth]}
          >
            needs auth
          </span>
        )}
      </div>
    </Link>
  );
}

function SegmentCatalog() {
  const data = usePluginData('oh-my-posh-segments');

  // Fail the build rather than degrade. Now that the plugin is registered in
  // docusaurus.config.js, undefined can only mean someone removed it, and a
  // page that quietly renders an apology is far worse than one that stops the
  // deploy: the catalog is the whole point of this route.
  if (!data) {
    throw new Error(
      "The 'oh-my-posh-segments' plugin returned no data. Check that './plugins/segments' is still listed in docusaurus.config.js.",
    );
  }

  const [filters, setFilters] = useFilters();

  // groupLabel isn't part of the plugin payload either (~117 copies of a value derivable from
  // data.groups, which already ships { id, label } per group) - build the lookup once here.
  const groupLabelById = useMemo(
    () => Object.fromEntries(data.groups.map((group) => [group.id, group.label])),
    [data],
  );

  const authLabelById = useMemo(
    () => Object.fromEntries(data.authTiers.map((tier) => [tier.id, tier.label])),
    [data],
  );

  const filtered = useMemo(
    () => data.segments.filter((s) => matchesAll(s, filters)),
    [data, filters],
  );

  const { total } = data;

  const hasActiveFilters = filters.q !== '';

  const resultsLine = hasActiveFilters
    ? `${filtered.length} of ${total} segments`
    : `${total} segments`;

  return (
    <div className={styles.cat}>
      <div className={styles.searchRow}>
        <label htmlFor="segment-catalog-search" className={styles.lbl}>
          Search
        </label>
        <input
          id="segment-catalog-search"
          type="text"
          className={styles.search}
          value={filters.q}
          placeholder="Search by name, type or description"
          onChange={(e) => setFilters({ ...filters, q: e.target.value })}
        />
      </div>

      <div className={styles.content}>
        <p className={styles.resultsLine}>{resultsLine}</p>

        {filtered.length === 0 ? (
          <div className={styles.empty}>
            <p>No segments match the current filters.</p>
            <button
              type="button"
              className="button button--secondary"
              onClick={() => setFilters(EMPTY_FILTERS)}
            >
              Clear filters
            </button>
          </div>
        ) : (
          <div className={styles.grid}>
            {filtered.map((segment) => (
              <SegmentCard
                key={segment.id}
                segment={segment}
                groupLabelById={groupLabelById}
                authLabelById={authLabelById}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default SegmentCatalog;
