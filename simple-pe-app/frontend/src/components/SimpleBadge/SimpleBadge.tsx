import React, { HTMLAttributes } from 'react';

// ─── Types ───────────────────────────────────────────────────────────────────

export type BadgeVariant = 'success' | 'error' | 'warning' | 'info' | 'neutral' | 'purple';
export type BadgeSize    = 'sm' | 'md';

export interface SimpleBadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?:     BadgeVariant;
  size?:        BadgeSize;
  dot?:         boolean;
  children:     React.ReactNode;
}

// ─── Token maps ───────────────────────────────────────────────────────────────

const variantTokens: Record<BadgeVariant, { bg: string; color: string; dot: string }> = {
  success: {
    bg:    'var(--color-success-bg)',
    color: 'var(--color-success)',
    dot:   'var(--color-success)',
  },
  error: {
    bg:    'var(--color-error-bg)',
    color: 'var(--color-error)',
    dot:   'var(--color-error)',
  },
  warning: {
    bg:    'var(--color-warning-bg)',
    color: 'var(--color-warning)',
    dot:   'var(--color-warning)',
  },
  info: {
    bg:    'var(--color-info-bg)',
    color: 'var(--color-info)',
    dot:   'var(--color-info)',
  },
  neutral: {
    bg:    'var(--color-border)',
    color: 'var(--color-text-secondary)',
    dot:   'var(--color-text-muted)',
  },
  purple: {
    bg:    'var(--color-purple-bg)',
    color: 'var(--color-brand)',
    dot:   'var(--color-brand)',
  },
};

const sizeTokens: Record<BadgeSize, { fontSize: string; padding: string; dotSize: number }> = {
  sm: { fontSize: 'var(--font-size-xs)', padding: '2px 8px',  dotSize: 6 },
  md: { fontSize: 'var(--font-size-sm)', padding: '4px 10px', dotSize: 7 },
};

// ─── Component ───────────────────────────────────────────────────────────────

export const SimpleBadge: React.FC<SimpleBadgeProps> = ({
  variant  = 'neutral',
  size     = 'md',
  dot      = false,
  children,
  style,
  ...rest
}) => {
  const tokens     = variantTokens[variant];
  const sizeConfig = sizeTokens[size];

  const badgeStyle: React.CSSProperties = {
    display:        'inline-flex',
    alignItems:     'center',
    gap:            dot ? '5px' : undefined,
    padding:        sizeConfig.padding,
    borderRadius:   'var(--radius-full)',
    fontSize:       sizeConfig.fontSize,
    fontWeight:     500,
    lineHeight:     1.4,
    backgroundColor: tokens.bg,
    color:           tokens.color,
    whiteSpace:     'nowrap',
    userSelect:     'none',
    ...style,
  };

  const dotStyle: React.CSSProperties = {
    width:           sizeConfig.dotSize,
    height:          sizeConfig.dotSize,
    borderRadius:    '50%',
    backgroundColor: tokens.dot,
    flexShrink:      0,
  };

  return (
    <span style={badgeStyle} {...rest}>
      {dot && <span style={dotStyle} aria-hidden="true" />}
      {children}
    </span>
  );
};

SimpleBadge.displayName = 'SimpleBadge';

export default SimpleBadge;
