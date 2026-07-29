// The sessionStorage handshake between a segment doc's editor and the studio (see Config.js's
// "Add to Studio" action and Studio/index.js's mount effect). Two keys:
//
//  - APPEND_KEY is a one-shot mailbox: a segment doc writes { segment } to it right before
//    navigating to /docs/studio, and the studio reads-then-deletes it on mount. sessionStorage
//    (rather than a URL param) because a segment can be an arbitrarily large object with no
//    length limit or encoding to worry about, and because it survives the navigation without
//    showing up in the address bar.
//  - SESSION_KEY is the studio's own "resume where I left off": it saves { format, text } on
//    every change and restores it on mount. Without this, walking from one segment doc to
//    another and back to the studio would find a fresh, pristine studio each time - APPEND_KEY's
//    segment would have nothing accumulated to append to, and "walk several segment pages and
//    accumulate a prompt" (the whole point of the hand-off) would only ever hold the last one.
//
// Every access goes through these try/catch wrappers: some browsers (Safari private mode,
// locked-down embeds, sandboxed iframes) throw on *any* sessionStorage touch, not just quota
// errors. A blocked hand-off should degrade to "just navigate" / "just show the pristine
// starter" rather than take the page down.
export const APPEND_KEY = 'omp-studio-pending-segment';
export const SESSION_KEY = 'omp-studio-session';

export function trySessionStorageGet(key) {
  try {
    return window.sessionStorage.getItem(key);
  } catch {
    return null;
  }
}

export function trySessionStorageSet(key, value) {
  try {
    window.sessionStorage.setItem(key, value);
    return true;
  } catch {
    return false;
  }
}

export function trySessionStorageRemove(key) {
  try {
    window.sessionStorage.removeItem(key);
  } catch {
    // Nothing sensible to do - the one-shot mailbox just won't be there next time either.
  }
}
