import React, { HTMLAttributes } from 'react';

// ─── Types ───────────────────────────────────────────────────────────────────

export type CardPadding = 'none' | 'sm' | 'md' | 'lg';

export interface SimpleCardProps extends HTMLAttributes<HTMLDivElement> {
  title?:     string;
  subtitle?:  string;
  actions?:   React.ReactNode;
  footer?:    React.ReactNode;
  padding?:   CardPadding;
  hoverable?: boolean;
  noBorder?:  boolean;
  children?:  React.ReactNode;
}

// ─── Padding map ─────────────────────────────────────────────────────────────

const paddingMap: Record<CardPadding, string> = {
  none: '0',
  sm:   'var(--spacing-3) var(--spacing-4)',
  md:   'var(--spacing-5) var(--spacing-6)',
  lg:   'var(--spacing-8) var(--spacing-8)',
};

// ─── Component ───────────────────────────────────────────────────────────────

export const SimpleCard = React.forwardRef<HTMLDivElement, SimpleCardProps>(
  (
    {
      title,
      subtitle,
      actions,
      footer,
      padding   = 'md',
      hoverable = false,
      noBorder  = false,
      children,
      style,
      onMouseEnter,
      onMouseLeave,
      ...rest
    },
    ref
  ) => {
    const [hovered, setHovered] = React.useState(false);

    const cardStyle: React.CSSProperties = {
      backgroundColor: 'var(--color-surface)',
      borderRadius:    'var(--radius-lg)',
      boxShadow:       hovered && hoverable
        ? '0 6px 20px rgba(0, 0, 0, 0.12)'
        : 'var(--shadow-card)',
      border:          noBorder ? 'none' : '1px solid var(--color-border)',
      overflow:        'hidden',
      transition:      'box-shadow var(--transition-base), transform var(--transition-base)',
      transform:       hovered && hoverable ? 'translateY(-2px)' : 'translateY(0)',
      cursor:          hoverable ? 'pointer' : 'default',
      ...style,
    };

    const hasHeader = title || subtitle || actions;

    return (
      <div
        ref={ref}
        style={cardStyle}
        onMouseEnter={(e) => { setHovered(true); onMouseEnter?.(e); }}
        onMouseLeave={(e) => { setHovered(false); onMouseLeave?.(e); }}
        {...rest}
      >
        {/* Header */}
        {hasHeader && (
          <div
            style={{
              display:        'flex',
              alignItems:     'flex-start',
              justifyContent: 'space-between',
              padding:        paddingMap[padding],
              paddingBottom:  children ? 'var(--spacing-4)' : undefined,
              borderBottom:   children
                ? '1px solid var(--color-border)'
                : 'none',
            }}
          >
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-1)' }}>
              {title && (
                <h3
                  style={{
                    fontFamily: 'var(--font-heading)',
                    fontSize:   'var(--font-size-lg)',
                    fontWeight: 600,
                    color:      'var(--color-text-primary)',
                    margin:     0,
                  }}
                >
                  {title}
                </h3>
              )}
              {subtitle && (
                <p
                  style={{
                    fontSize: 'var(--font-size-sm)',
                    color:    'var(--color-text-secondary)',
                    margin:   0,
                  }}
                >
                  {subtitle}
                </p>
              )}
            </div>
            {actions && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)', flexShrink: 0 }}>
                {actions}
              </div>
            )}
          </div>
        )}

        {/* Body */}
        {children && (
          <div
            style={{
              padding: paddingMap[padding],
            }}
          >
            {children}
          </div>
        )}

        {/* Footer */}
        {footer && (
          <div
            style={{
              padding:    paddingMap[padding],
              paddingTop: 'var(--spacing-4)',
              borderTop:  '1px solid var(--color-border)',
              display:    'flex',
              alignItems: 'center',
              justifyContent: 'flex-end',
              gap: 'var(--spacing-2)',
            }}
          >
            {footer}
          </div>
        )}
      </div>
    );
  }
);

SimpleCard.displayName = 'SimpleCard';

export default SimpleCard;
