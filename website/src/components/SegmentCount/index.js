import React from 'react';
import { usePluginData } from '@docusaurus/useGlobalData';

// Reads the segments plugin's computed total instead of hardcoding a count in prose, so any MDX
// page (which can't call usePluginData directly) can drop this component in wherever the segment
// count is mentioned and never drift from the real number. Mirrors src/pages/index.js's own
// usePluginData('oh-my-posh-segments') read for the homepage feature card.
function SegmentCount() {
  const { total } = usePluginData('oh-my-posh-segments');
  return total;
}

export default SegmentCount;
