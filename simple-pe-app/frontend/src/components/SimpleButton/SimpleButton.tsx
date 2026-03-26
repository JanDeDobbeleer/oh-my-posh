import React, { ButtonHTMLAttributes } from 'react';

// ─── Types ───────────────────────────────────────────────────────────────────

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize    = 'sm' | 'md' | 'lg';

export interface SimpleButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?:  ButtonVariant;
  size?:     ButtonSize;
  loading?:  boolean;
  fullWidth?: boolean;
  leftIcon?:  React.ReactNode;
  rightIcon?: React.ReactNode;
}

// ─── Spinner ─────────────────────────────────────────────────────────────────

const Spinner: React.FC<{ size: number }> = ({ size }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    style={{
      animation: 'spin 0.7s linear infinite',
      flexShrink: 0,
    }}
    aria-hidden="true"
  >
    <circle
      cx="12"
      cy="12"
      r="10"
      stroke="currentColor"
      strokeWidth="3"
      strokeLinecap="round"
      strokeDasharray="40 20"
    />
    <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
  </svg>
);

// ─── Style helpers ────────────────────────────────────────────────────────────

const variantStyles: Record<ButtonVariant, React.CSSProperties & { '--btn-hover-bg'?: string }> = {
  primary: {
    backgroundColor: 'var(--color-brand)',
    color: '#ffffff',
    border: '2px solid transparent',
  },
  secondary: {
    backgroundColor: 'transparent',
    color: 'var(--color-brand)',
    border: '2px solid var(--color-brand)',
  },
  ghost: {
    backgroundColor: 'transparent',
    color: 'var(--color-text-secondary)',
    border: '2px solid transparent',
  },
  danger: {
    backgroundColor: 'var(--color-error)',
    color: '#ffffff',
    border: '2px solid transparent',
  },
};

const sizeStyles: Record<ButtonSize, React.CSSProperties> = {
  sm: {
    padding: '6px 14px',
    fontSize: 'var(--font-size-sm)',
    gap: '6px',
  },
  md: {
    padding: '10px 20px',
    fontSize: 'var(--font-size-base)',
    gap: '8px',
  },
  lg: {
    padding: '14px 28px',
    fontSize: 'var(--font-size-lg)',
    gap: '10px',
  },
};

// ─── Component ───────────────────────────────────────────────────────────────

export const SimpleButton = React.forwardRef<HTMLButtonElement, SimpleButtonProps>(
  (
    {
      variant  = 'primary',
      size     = 'md',
      loading  = false,
      fullWidth = false,
      leftIcon,
      rightIcon,
      disabled,
      children,
      style,
      onMouseEnter,
      onMouseLeave,
      ...rest
    },
    ref
  ) => {
    const [hovered, setHovered] = React.useState(false);

    const isDisabled = disabled || loading;

    const hoverOverlay: React.CSSProperties = hovered && !isDisabled
      ? variant === 'primary'
        ? { backgroundColor: 'var(--color-brand-dark)' }
        : variant === 'danger'
        ? { backgroundColor: '#DC2626' }
        : variant === 'secondary'
        ? { backgroundColor: 'var(--color-purple-bg, #EDE9FE)' }
        : { backgroundColor: 'var(--color-border)' }
      : {};

    const spinnerSize = size === 'sm' ? 14 : size === 'lg' ? 20 : 16;

    const baseStyle: React.CSSProperties = {
      display:        'inline-flex',
      alignItems:     'center',
      justifyContent: 'center',
      borderRadius:   'var(--radius-full)',
      fontFamily:     'var(--font-body)',
      fontWeight:     600,
      lineHeight:     1,
      cursor:         isDisabled ? 'not-allowed' : 'pointer',
      opacity:        isDisabled ? 0.55 : 1,
      transition:     'background-color var(--transition-fast), box-shadow var(--transition-fast), opacity var(--transition-fast)',
      outline:        'none',
      width:          fullWidth ? '100%' : undefined,
      whiteSpace:     'nowrap',
      userSelect:     'none',
      boxShadow:      hovered && !isDisabled && variant === 'primary'
        ? '0 4px 12px rgba(108, 60, 225, 0.35)'
        : 'none',
      ...variantStyles[variant],
      ...sizeStyles[size],
      ...hoverOverlay,
      ...style,
    };

    return (
      <button
        ref={ref}
        disabled={isDisabled}
        style={baseStyle}
        onMouseEnter={(e) => { setHovered(true); onMouseEnter?.(e); }}
        onMouseLeave={(e) => { setHovered(false); onMouseLeave?.(e); }}
        aria-busy={loading}
        {...rest}
      >
        {loading ? (
          <Spinner size={spinnerSize} />
        ) : leftIcon ? (
          <span style={{ display: 'flex', alignItems: 'center', flexShrink: 0 }}>
            {leftIcon}
          </span>
        ) : null}

        {children && (
          <span style={{ display: 'inline-flex', alignItems: 'center' }}>
            {children}
          </span>
        )}

        {!loading && rightIcon && (
          <span style={{ display: 'flex', alignItems: 'center', flexShrink: 0 }}>
            {rightIcon}
          </span>
        )}
      </button>
    );
  }
);

SimpleButton.displayName = 'SimpleButton';

export default SimpleButton;
