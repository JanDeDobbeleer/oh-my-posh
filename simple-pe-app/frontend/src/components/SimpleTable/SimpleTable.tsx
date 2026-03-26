import React, { useState } from 'react';

// ─── Types ───────────────────────────────────────────────────────────────────

export type SortDirection = 'asc' | 'desc';

export interface TableColumn<T> {
  key:       keyof T | string;
  header:    string;
  sortable?: boolean;
  width?:    string;
  align?:    'left' | 'center' | 'right';
  render?:   (value: unknown, row: T, index: number) => React.ReactNode;
}

export interface SimpleTableProps<T extends Record<string, unknown>> {
  columns:       TableColumn<T>[];
  data:          T[];
  loading?:      boolean;
  emptyMessage?: string;
  onRowClick?:   (row: T, index: number) => void;
  rowKey?:       keyof T | ((row: T, index: number) => string);
  stickyHeader?: boolean;
  maxHeight?:    string;
}

// ─── Skeleton Row ─────────────────────────────────────────────────────────────

const SkeletonCell: React.FC<{ width?: string }> = ({ width = '80%' }) => (
  <td style={{ padding: 'var(--spacing-3) var(--spacing-4)' }}>
    <div
      style={{
        height:           14,
        width,
        borderRadius:     'var(--radius-sm)',
        backgroundColor:  'var(--color-border)',
        animation:        'shimmer 1.4s ease-in-out infinite',
      }}
    />
  </td>
);

// ─── Sort Icon ────────────────────────────────────────────────────────────────

const SortIcon: React.FC<{ direction: SortDirection | null }> = ({ direction }) => (
  <span
    style={{
      display:       'inline-flex',
      flexDirection: 'column',
      marginLeft:    'var(--spacing-1)',
      gap:           1,
      verticalAlign: 'middle',
    }}
  >
    <svg width="8" height="5" viewBox="0 0 8 5" fill="none">
      <path
        d="M4 0L7.46 4.5H.54L4 0Z"
        fill={direction === 'asc' ? 'var(--color-brand)' : 'var(--color-text-muted)'}
      />
    </svg>
    <svg width="8" height="5" viewBox="0 0 8 5" fill="none">
      <path
        d="M4 5L.54.5H7.46L4 5Z"
        fill={direction === 'desc' ? 'var(--color-brand)' : 'var(--color-text-muted)'}
      />
    </svg>
  </span>
);

// ─── Component ───────────────────────────────────────────────────────────────

function SimpleTableInner<T extends Record<string, unknown>>(
  {
    columns,
    data,
    loading      = false,
    emptyMessage = 'No hay datos disponibles',
    onRowClick,
    rowKey,
    stickyHeader = false,
    maxHeight,
  }: SimpleTableProps<T>,
  _ref: React.ForwardedRef<HTMLDivElement>
) {
  const [sortKey, setSortKey]           = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');
  const [hoveredRow, setHoveredRow]      = useState<number | null>(null);

  const handleSort = (key: string) => {
    if (sortKey === key) {
      setSortDirection((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDirection('asc');
    }
  };

  const sortedData = React.useMemo(() => {
    if (!sortKey) return data;
    return [...data].sort((a, b) => {
      const av = a[sortKey];
      const bv = b[sortKey];
      if (av === undefined || av === null) return 1;
      if (bv === undefined || bv === null) return -1;
      const cmp = av < bv ? -1 : av > bv ? 1 : 0;
      return sortDirection === 'asc' ? cmp : -cmp;
    });
  }, [data, sortKey, sortDirection]);

  const getRowKey = (row: T, index: number): string => {
    if (!rowKey) return String(index);
    if (typeof rowKey === 'function') return rowKey(row, index);
    return String(row[rowKey] ?? index);
  };

  const getCellValue = (row: T, col: TableColumn<T>, index: number): React.ReactNode => {
    const value = row[col.key as keyof T];
    if (col.render) return col.render(value as unknown, row, index);
    if (value === null || value === undefined) return '—';
    return String(value);
  };

  const thStyle: React.CSSProperties = {
    padding:         'var(--spacing-3) var(--spacing-4)',
    textAlign:       'left',
    fontSize:        'var(--font-size-xs)',
    fontWeight:      600,
    textTransform:   'uppercase',
    letterSpacing:   '0.05em',
    color:           'var(--color-text-secondary)',
    backgroundColor: 'var(--color-bg)',
    borderBottom:    '2px solid var(--color-border)',
    whiteSpace:      'nowrap',
    position:        stickyHeader ? 'sticky' : 'static',
    top:             0,
    zIndex:          stickyHeader ? 1 : 'auto',
  };

  const skeletonRows = Array.from({ length: 5 });

  return (
    <>
      <style>{`
        @keyframes shimmer {
          0%   { opacity: 1; }
          50%  { opacity: 0.4; }
          100% { opacity: 1; }
        }
      `}</style>
      <div
        ref={_ref}
        style={{
          width:        '100%',
          overflowX:    'auto',
          overflowY:    maxHeight ? 'auto' : 'visible',
          maxHeight,
          borderRadius: 'var(--radius-lg)',
          border:       '1px solid var(--color-border)',
        }}
      >
        <table
          style={{
            width:           '100%',
            borderCollapse:  'collapse',
            fontSize:        'var(--font-size-sm)',
            backgroundColor: 'var(--color-surface)',
          }}
        >
          <thead>
            <tr>
              {columns.map((col) => (
                <th
                  key={String(col.key)}
                  style={{
                    ...thStyle,
                    width:    col.width,
                    textAlign: col.align ?? 'left',
                    cursor:    col.sortable ? 'pointer' : 'default',
                    userSelect: 'none',
                  }}
                  onClick={col.sortable ? () => handleSort(String(col.key)) : undefined}
                >
                  {col.header}
                  {col.sortable && (
                    <SortIcon
                      direction={sortKey === String(col.key) ? sortDirection : null}
                    />
                  )}
                </th>
              ))}
            </tr>
          </thead>

          <tbody>
            {loading ? (
              skeletonRows.map((_, ri) => (
                <tr key={ri}>
                  {columns.map((col) => (
                    <SkeletonCell key={String(col.key)} width={ri % 2 === 0 ? '70%' : '85%'} />
                  ))}
                </tr>
              ))
            ) : sortedData.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  style={{
                    padding:   'var(--spacing-10) var(--spacing-4)',
                    textAlign: 'center',
                    color:     'var(--color-text-muted)',
                    fontSize:  'var(--font-size-sm)',
                  }}
                >
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 'var(--spacing-2)' }}>
                    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="var(--color-border)" strokeWidth="1.5">
                      <rect x="3" y="3" width="18" height="18" rx="2" />
                      <path d="M3 9h18M9 21V9" />
                    </svg>
                    <span>{emptyMessage}</span>
                  </div>
                </td>
              </tr>
            ) : (
              sortedData.map((row, rowIndex) => (
                <tr
                  key={getRowKey(row, rowIndex)}
                  onClick={onRowClick ? () => onRowClick(row, rowIndex) : undefined}
                  onMouseEnter={() => setHoveredRow(rowIndex)}
                  onMouseLeave={() => setHoveredRow(null)}
                  style={{
                    backgroundColor: hoveredRow === rowIndex
                      ? 'var(--color-bg)'
                      : 'transparent',
                    cursor:      onRowClick ? 'pointer' : 'default',
                    transition:  'background-color var(--transition-fast)',
                    borderBottom: '1px solid var(--color-border)',
                  }}
                >
                  {columns.map((col) => (
                    <td
                      key={String(col.key)}
                      style={{
                        padding:    'var(--spacing-3) var(--spacing-4)',
                        textAlign:  col.align ?? 'left',
                        color:      'var(--color-text-primary)',
                        verticalAlign: 'middle',
                        whiteSpace: col.align === 'right' ? 'nowrap' : 'normal',
                      }}
                    >
                      {getCellValue(row, col, rowIndex)}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </>
  );
}

export const SimpleTable = React.forwardRef(SimpleTableInner) as <T extends Record<string, unknown>>(
  props: SimpleTableProps<T> & { ref?: React.ForwardedRef<HTMLDivElement> }
) => React.ReactElement;

(SimpleTable as unknown as { displayName: string }).displayName = 'SimpleTable';

export default SimpleTable;
