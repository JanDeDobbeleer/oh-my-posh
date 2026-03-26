import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'node:path';

export default defineConfig({
  plugins: [react()],
  resolve: { alias: { '@': resolve(__dirname, './src') } },
  server: {
    port: 5173,
    proxy: { '/api': { target: 'http://localhost:4000', changeOrigin: true, secure: false } },
  },
  preview: { port: 4173 },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: { output: { manualChunks: { vendor: ['react','react-dom'], router: ['react-router-dom'], icons: ['lucide-react'], http: ['axios'] } } },
  },
});
