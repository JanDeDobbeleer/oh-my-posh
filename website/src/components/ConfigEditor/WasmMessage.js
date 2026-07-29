import React from 'react';

// Shared between the studio and a segment doc's editor (see useWasmRenderer.js's `eager` doc
// comment for why their status starts differently). 'idle' - the segment editor's status before
// a reader's first edit - renders nothing, same as 'ready': there is no download in flight to
// report on, and the page is already showing its build-time static preview.
function WasmMessage({ status, progress, errorMessage, className }) {
  if (status === 'idle' || status === 'ready') {
    return null;
  }

  if (status === 'error') {
    return <p className={className}>Could not load the renderer: {errorMessage}</p>;
  }

  const pct = progress > 0 ? ` (${Math.round(progress * 100)}%)` : '';

  return (
    <p className={className}>
      Downloading the renderer{pct}. It is a ~29 MB WebAssembly build of oh-my-posh, fetched once
      per visit.
    </p>
  );
}

export default WasmMessage;
