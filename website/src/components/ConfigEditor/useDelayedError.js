import { useEffect, useRef, useState } from 'react';

// A parse error appears the instant a keystroke or an accepted completion leaves the config
// momentarily invalid (e.g. an autocompleted `{` with no body yet, or a completion popup that's
// still open) - showing it right away would flag the reader's own in-progress edit as a mistake
// before they've had a chance to finish it. Holding the error back for a short, silent grace
// period on its FIRST appearance means it only ever surfaces once the reader has actually
// paused, not on every transient mid-edit state. Once shown, further changes to the message
// (e.g. the reported line/column shifting as they keep typing) update immediately - only the
// initial appearance is delayed. Clearing is never delayed either: the moment the config parses
// again, the error disappears right away.
const ERROR_DELAY_MS = 600;

export function useDelayedError(errorValue, delay = ERROR_DELAY_MS) {
  const [visible, setVisible] = useState(Boolean(errorValue));
  const timeoutRef = useRef(null);

  useEffect(() => {
    if (!errorValue) {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }

      setVisible(false);
      return undefined;
    }

    if (!visible && !timeoutRef.current) {
      timeoutRef.current = setTimeout(() => {
        timeoutRef.current = null;
        setVisible(true);
      }, delay);
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [errorValue, delay]);

  return visible ? errorValue : null;
}
