import React from "react";
import classnames from "classnames";
import BrowserOnly from "@docusaurus/BrowserOnly";

export default function FooterLayout({ style, links, logo, copyright }) {
  return (
    <footer
      className={classnames("footer", {
        "footer--dark": style === "dark",
      })}
    >
      <div className="container container-fluid">
        <div className="footerGridBanner">
          <div className="footerGridFade" />
          <BrowserOnly>
            {() => {
              const FlickeringGrid = require("@site/src/components/FlickeringGrid").default;
              const tablet = window.matchMedia("(max-width: 1024px)").matches;
              return (
                <FlickeringGrid
                  text="Oh My Posh"
                  fontSize={tablet ? 70 : 90}
                  squareSize={2}
                  gridGap={tablet ? 2 : 3}
                  color="#6B7280"
                  maxOpacity={0.3}
                  flickerChance={0.1}
                />
              );
            }}
          </BrowserOnly>
        </div>
        <div className="footerDivider" />
        {links}
        {(logo || copyright) && (
          <div className="footer__bottom text--center">
            {logo && <div className="margin-bottom--sm">{logo}</div>}
            {copyright}
          </div>
        )}
      </div>
    </footer>
  );
}
