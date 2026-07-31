import React, { useRef, useState } from 'react';
import classnames from 'classnames';
import IconWarning from '@theme/Admonition/Icon/Warning';
import styles from './styles.module.css';

// A small warning triangle rendered into ConfigEditor's actions slot (next to the
// download/copy/add-to-studio buttons) - visible whenever there's a parse error to report, so
// a reader whose cursor has scrolled away from the offending line still notices something needs
// fixing, without the top-of-preview Admonition banner shoving the svg up and down on every
// keystroke. Hovering (or focusing, for keyboard/touch) shows the same message the squiggle's
// own hover tooltip does (see ConfigEditor/index.js's errorMessage prop) - two entry points to
// one message, not two different ones.
function ErrorIndicator({ message }) {
  const [open, setOpen] = useState(false);
  // Anchored bottom-right by default (the triangle usually sits at the end of a right-aligned
  // action row), each flipped only when the bubble's default corner would actually run past
  // the matching viewport edge - rather than a fixed side, so this keeps working regardless of
  // where in the page the editor (and this row of actions) ends up.
  const [openLeft, setOpenLeft] = useState(false);
  const [openUpward, setOpenUpward] = useState(false);
  const wrapperRef = useRef(null);

  if (!message) {
    return null;
  }

  const show = () => {
    const rect = wrapperRef.current?.getBoundingClientRect();
    if (rect) {
      const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize) || 16;
      const bubbleWidth = 18 * rootFontSize;
      // No real measurement to go on before the bubble itself has rendered - a generous
      // estimate is enough here, this only decides which side of the flip to start on.
      const estimatedBubbleHeight = 6 * rootFontSize;
      setOpenLeft(rect.right - bubbleWidth < 0);
      setOpenUpward(rect.bottom + estimatedBubbleHeight > window.innerHeight);
    }
    setOpen(true);
  };
  const hide = () => setOpen(false);

  return (
    <span
      ref={wrapperRef}
      className={styles.errorIndicator}
      tabIndex={0}
      role="img"
      aria-label={`Config error: ${message}`}
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocus={show}
      onBlur={hide}
    >
      <IconWarning className={styles.errorIndicatorIcon} />
      {open && (
        <div
          className={classnames(styles.completionTooltip, styles.errorIndicatorTooltip, {
            [styles.errorIndicatorTooltipLeft]: openLeft,
            [styles.errorIndicatorTooltipUp]: openUpward,
          })}
          role="tooltip"
        >
          <div className={styles.completionTooltipTitle}>Config error</div>
          <div className={styles.completionTooltipText}>{message}</div>
        </div>
      )}
    </span>
  );
}

export default ErrorIndicator;
