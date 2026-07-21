import React from "react";
import BrowserOnly from "@docusaurus/BrowserOnly";

export default function CubeScene() {
  return (
    <BrowserOnly>
      {() => {
        const UnicornScene = require("unicornstudio-react").default;
        return (
          <UnicornScene
            jsonFilePath="/unicorn/cube-unicorn.json"
            sdkUrl="https://cdn.jsdelivr.net/gh/hiunicornstudio/unicornstudio.js@v2.2.5/dist/unicornStudio.umd.js"
            width="100%"
            height="100%"
            lazyLoad={true}
          />
        );
      }}
    </BrowserOnly>
  );
}
