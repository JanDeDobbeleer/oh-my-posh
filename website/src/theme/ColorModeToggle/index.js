import React from 'react';
import OriginalColorModeToggle from '@theme-original/ColorModeToggle';
import { useColorMode } from '@docusaurus/theme-common';

export default function ColorModeToggle(props) {
  const { colorMode } = useColorMode();

  return (
    <OriginalColorModeToggle
      {...props}
      respectPrefersColorScheme={false}
      value={colorMode}
    />
  );
}
