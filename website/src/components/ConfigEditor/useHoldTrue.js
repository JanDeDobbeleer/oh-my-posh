import { useEffect, useRef, useState } from 'react';

// Mirrors `value` the instant it turns true, but keeps returning true for `graceMs` after it
// turns false. Used to suppress the parse-error banner not just while the completion popup is
// open, but for a short buffer afterwards too: accepting a completion edits the config text
// through the same debounced render pipeline as normal typing (see DEBOUNCE_MS in
// Studio/index.js and Config.js), so without this buffer the still-stale error from before the
// edit flashes back into view for that debounce gap before the re-render clears it.
export function useHoldTrue(value, graceMs) {
  const [held, setHeld] = useState(value);
  const timeoutRef = useRef(null);

  useEffect(() => {
    if (value) {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }

      setHeld(true);
      return undefined;
    }

    timeoutRef.current = setTimeout(() => {
      timeoutRef.current = null;
      setHeld(false);
    }, graceMs);

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
    };
  }, [value, graceMs]);

  return held;
}
