import React, { InputHTMLAttributes, forwardRef, useState } from 'react';

// ─── Types ───────────────────────────────────────────────────────────────────

export type InputVariant = 'default' | 'filled';
export type InputSize    = 'sm' | 'md' | 'lg';

export interface SimpleInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label?:      string;
  error?:      string;
  helperText?: string;
  leftIcon?:   React.ReactNode;
  rightIcon?:  React.ReactNode;
  variant?:    InputVariant;
  size?:       InputSize;
  fullWidth?:  boolean;
}

// ─── Size map ─────────────────────────────────────────────────────────────────

const sizeMap: Record<InputSize, { height: number; fontSize: string; iconSize: number; paddingX: string }> = {
  sm: { height: 34, fontSize: 'var(--font-size-sm)',   iconSize: 14, paddingX: 'var(--spacing-3)' },
  md: { height: 42, fontSize: 'var(--font-size-base)', iconSize: 16, paddingX: 'var(--spacing-4)' },
  lg: { height: 50, fontSize: 'var(--font-size-lg)',   iconSize: 18, paddingX: 'var(--spacing-5)' },
};

// ─── Component ───────────────────────────────────────────────────────────────

export const SimpleInput = forwardRef<HTMLInputElement, SimpleInputProps>(
  (
    {
      label,
      error,
      helperText,
      leftIcon,
      rightIcon,
      variant    = 'default',
      size       = 'md',
      fullWidth  = true,
      disabled,
      style,
      id,
      ...rest
    },
    ref
  ) => {
    const [focused, setFocused] = useState(false);

    const inputId = id ?? (label ? `input-${label.toLowerCase().replace(/\s+/g, '-')}` : undefined);
    const sz      = sizeMap[size];

    const hasError = Boolean(error);

    const wrapperStyle: React.CSSProperties = {
      display:       'flex',
      flexDirection: 'column',
      gap:           'var(--spacing-1)',
      width:         fullWidth ? '100%' : 'auto',
    };

    const labelStyle: React.CSSProperties = {
      fontSize:   'var(--font-size-sm)',
      fontWeight: 500,
      color:      hasError
        ? 'var(--color-error)'
        : focused
        ? 'var(--color-brand)'
        : 'var(--color-text-secondary)',
      transition: 'color var(--transition-fast)',
    };

    const containerStyle: React.CSSProperties = {
      position:        'relative',
      display:         'flex',
      alignItems:      'center',
      height:          sz.height,
      borderRadius:    'var(--radius-md)',
      border:          `1.5px solid ${
        hasError
          ? 'var(--color-error)'
          : focused
          ? 'var(--color-brand)'
          : 'var(--color-border)'
      }`,
      backgroundColor: variant === 'filled'
        ? 'var(--color-bg)'
        : 'var(--color-surface)',
      boxShadow:        focused && !hasError
        ? '0 0 0 3px rgba(108, 60, 225, 0.12)'
        : focused && hasError
        ? '0 0 0 3px rgba(239, 68, 68, 0.12)'
        : 'none',
      transition: 'border-color var(--transition-fast), box-shadow var(--transition-fast)',
      opacity:    disabled ? 0.55 : 1,
    };

    const iconPad = leftIcon  ? `calc(${sz.paddingX} + ${sz.iconSize}px + 8px)` : sz.paddingX;
    const rPad    = rightIcon ? `calc(${sz.paddingX} + ${sz.iconSize}px + 8px)` : sz.paddingX;

    const inputStyle: React.CSSProperties = {
      flex:            1,
      height:          '100%',
      border:          'none',
      outline:         'none',
      background:      'transparent',
      fontSize:        sz.fontSize,
      color:           'var(--color-text-primary)',
      paddingLeft:     iconPad,
      paddingRight:    rPad,
      fontFamily:      'var(--font-body)',
      cursor:          disabled ? 'not-allowed' : 'text',
      ...style,
    };

    const iconStyle: (side: 'left' | 'right') => React.CSSProperties = (side) => ({
      position:    'absolute',
      [side]:      sz.paddingX,
      top:         '50%',
      transform:   'translateY(-50%)',
      display:     'flex',
      alignItems:  'center',
      color:       hasError
        ? 'var(--color-error)'
        : focused
        ? 'var(--color-brand)'
        : 'var(--color-text-muted)',
      pointerEvents: 'none',
      transition:  'color var(--transition-fast)',
      fontSize:    sz.iconSize,
    });

    const helperStyle: React.CSSProperties = {
      fontSize: 'var(--font-size-xs)',
      color:    hasError ? 'var(--color-error)' : 'var(--color-text-muted)',
      marginTop: 'var(--spacing-1)',
    };

    return (
      <div style={wrapperStyle}>
        {label && (
          <label htmlFor={inputId} style={labelStyle}>
            {label}
          </label>
        )}

        <div style={containerStyle}>
          {leftIcon && (
            <span style={iconStyle('left')}>{leftIcon}</span>
          )}

          <input
            ref={ref}
            id={inputId}
            disabled={disabled}
            style={inputStyle}
            onFocus={(e) => { setFocused(true); rest.onFocus?.(e); }}
            onBlur={(e)  => { setFocused(false); rest.onBlur?.(e); }}
            aria-invalid={hasError}
            aria-describedby={
              hasError
                ? `${inputId}-error`
                : helperText
                ? `${inputId}-helper`
                : undefined
            }
            {...rest}
          />

          {rightIcon && (
            <span style={iconStyle('right')}>{rightIcon}</span>
          )}
        </div>

        {(error || helperText) && (
          <p
            id={hasError ? `${inputId}-error` : `${inputId}-helper`}
            style={helperStyle}
            role={hasError ? 'alert' : undefined}
          >
            {error ?? helperText}
          </p>
        )}
      </div>
    );
  }
);

SimpleInput.displayName = 'SimpleInput';

export default SimpleInput;
