import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { it } from "node:test";

const source = await readFile(
  new URL("./exportPng.js", import.meta.url),
  "utf8"
);

for (const failure of ["fetch", "read"]) {
  it(`can retry a PNG export after a font ${failure} failure`, async (t) => {
    const { exportSvgAsPng } = await import(
      `data:text/javascript;base64,${Buffer.from(source).toString("base64")}#${failure}`
    );
    let attempts = 0;
    let downloads = 0;
    const error = new Error("Temporary font failure");

    t.mock.method(globalThis, "fetch", async () => {
      attempts++;
      if (failure === "fetch" && attempts === 1) {
        throw error;
      }
      return { blob: async () => new Blob(["font"]) };
    });

    const globals = {
      FileReader: class {
        readAsDataURL() {
          if (failure === "read" && attempts === 1) {
            this.error = error;
            this.onerror();
            return;
          }
          this.result = "data:font/ttf;base64,Zm9udA==";
          this.onload();
        }
      },
      Image: class {
        set src(value) {
          this.onload();
        }
      },
      document: {
        createElement(tag) {
          if (tag === "canvas") {
            return {
              getContext: () => ({ drawImage() {} }),
              toBlob: (callback) => callback(new Blob(["png"]))
            };
          }
          return {
            click() {
              downloads++;
            },
            remove() {}
          };
        },
        body: { appendChild() {} }
      }
    };

    for (const [name, value] of Object.entries(globals)) {
      const descriptor = Object.getOwnPropertyDescriptor(globalThis, name);
      Object.defineProperty(globalThis, name, { configurable: true, value });
      t.after(() => {
        if (descriptor) {
          Object.defineProperty(globalThis, name, descriptor);
          return;
        }
        delete globalThis[name];
      });
    }

    const svg = '<svg width="10" height="10"></svg>';
    await assert.rejects(exportSvgAsPng(svg), error);
    await exportSvgAsPng(svg);
    await exportSvgAsPng(svg);

    assert.equal(attempts, 2);
    assert.equal(downloads, 2);
  });
}
