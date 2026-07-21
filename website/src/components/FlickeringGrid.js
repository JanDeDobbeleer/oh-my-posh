import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

function toRGB(color) {
  if (color.startsWith("#")) {
    let h = color.slice(1);
    if (h.length === 3) h = h.split("").map((c) => c + c).join("");
    const n = parseInt(h, 16);
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
  }
  const m = color.match(/(\d+)[, ]+(\d+)[, ]+(\d+)/);
  if (m) return { r: +m[1], g: +m[2], b: +m[3] };
  return { r: 180, g: 180, b: 180 };
}

export default function FlickeringGrid({
  squareSize = 3,
  gridGap = 3,
  flickerChance = 0.2,
  color = "#B4B4B4",
  colors,
  width,
  height,
  className,
  maxOpacity = 0.15,
  text = "",
  fontSize = 140,
  fontWeight = 600,
  ...props
}) {
  const canvasRef = useRef(null);
  const containerRef = useRef(null);
  const [isInView, setIsInView] = useState(false);
  const [canvasSize, setCanvasSize] = useState({ width: 0, height: 0 });

  const rgbColors = useMemo(
    () => (colors && colors.length ? colors : [color]).map(toRGB),
    [colors, color]
  );

  const setupCanvas = useCallback(
    (canvas, w, h) => {
      const dpr = window.devicePixelRatio || 1;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      canvas.style.width = `${w}px`;
      canvas.style.height = `${h}px`;
      const cols = Math.ceil(w / (squareSize + gridGap));
      const rows = Math.ceil(h / (squareSize + gridGap));
      const squares = new Float32Array(cols * rows);
      const colorIdx = new Uint8Array(cols * rows);
      for (let i = 0; i < squares.length; i++) {
        squares[i] = Math.random() * maxOpacity;
        colorIdx[i] = Math.floor(Math.random() * rgbColors.length);
      }

      // Pre-compute the text mask once per resize (visually identical, cheaper per frame)
      let hasText = null;
      if (text) {
        const maskCanvas = document.createElement("canvas");
        maskCanvas.width = canvas.width;
        maskCanvas.height = canvas.height;
        const mctx = maskCanvas.getContext("2d", { willReadFrequently: true });
        if (mctx) {
          mctx.save();
          mctx.scale(dpr, dpr);
          mctx.fillStyle = "white";
          mctx.font = `${fontWeight} ${fontSize}px Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`;
          mctx.textAlign = "center";
          mctx.textBaseline = "middle";
          mctx.fillText(text, w / 2, h / 2);
          mctx.restore();
          const data = mctx.getImageData(0, 0, maskCanvas.width, maskCanvas.height).data;
          hasText = new Uint8Array(cols * rows);
          for (let i = 0; i < cols; i++) {
            for (let j = 0; j < rows; j++) {
              const px = Math.min(
                maskCanvas.width - 1,
                Math.floor((i * (squareSize + gridGap) + squareSize / 2) * dpr)
              );
              const py = Math.min(
                maskCanvas.height - 1,
                Math.floor((j * (squareSize + gridGap) + squareSize / 2) * dpr)
              );
              if (data[(py * maskCanvas.width + px) * 4] > 0) {
                hasText[i * rows + j] = 1;
              }
            }
          }
        }
      }

      return { cols, rows, squares, colorIdx, hasText, dpr };
    },
    [squareSize, gridGap, maxOpacity, text, fontSize, fontWeight, rgbColors]
  );

  const updateSquares = useCallback(
    (squares, deltaTime) => {
      for (let i = 0; i < squares.length; i++) {
        if (Math.random() < flickerChance * deltaTime) {
          squares[i] = Math.random() * maxOpacity;
        }
      }
    },
    [flickerChance, maxOpacity]
  );

  const drawGrid = useCallback(
    (ctx, w, h, cols, rows, squares, colorIdx, hasText, dpr) => {
      ctx.clearRect(0, 0, w, h);
      for (let i = 0; i < cols; i++) {
        for (let j = 0; j < rows; j++) {
          const idx = i * rows + j;
          const opacity = squares[idx];
          const finalOpacity =
            hasText && hasText[idx] ? Math.min(1, opacity * 3 + 0.4) : opacity;
          const c = rgbColors[colorIdx[idx] % rgbColors.length];
          ctx.fillStyle = `rgba(${c.r},${c.g},${c.b},${finalOpacity})`;
          ctx.fillRect(
            i * (squareSize + gridGap) * dpr,
            j * (squareSize + gridGap) * dpr,
            squareSize * dpr,
            squareSize * dpr
          );
        }
      }
    },
    [rgbColors, squareSize, gridGap]
  );

  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let animationFrameId;
    let gridParams;

    const updateCanvasSize = () => {
      const newWidth = width || container.clientWidth;
      const newHeight = height || container.clientHeight;
      setCanvasSize({ width: newWidth, height: newHeight });
      gridParams = setupCanvas(canvas, newWidth, newHeight);
    };

    updateCanvasSize();

    let lastTime = 0;
    const animate = (time) => {
      if (!isInView) return;
      const deltaTime = (time - lastTime) / 1000;
      lastTime = time;
      updateSquares(gridParams.squares, deltaTime);
      drawGrid(
        ctx,
        canvas.width,
        canvas.height,
        gridParams.cols,
        gridParams.rows,
        gridParams.squares,
        gridParams.colorIdx,
        gridParams.hasText,
        gridParams.dpr
      );
      animationFrameId = requestAnimationFrame(animate);
    };

    const resizeObserver = new ResizeObserver(() => {
      updateCanvasSize();
    });
    resizeObserver.observe(container);

    const intersectionObserver = new IntersectionObserver(
      ([entry]) => {
        setIsInView(entry.isIntersecting);
      },
      { threshold: 0 }
    );
    intersectionObserver.observe(canvas);

    if (isInView) {
      animationFrameId = requestAnimationFrame(animate);
    }

    return () => {
      cancelAnimationFrame(animationFrameId);
      resizeObserver.disconnect();
      intersectionObserver.disconnect();
    };
  }, [setupCanvas, updateSquares, drawGrid, width, height, isInView]);

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ height: "100%", width: "100%" }}
      {...props}
    >
      <canvas
        ref={canvasRef}
        style={{
          pointerEvents: "none",
          width: canvasSize.width,
          height: canvasSize.height,
        }}
      />
    </div>
  );
}
