import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// The app builds into ../static/app, which the Go binary embeds (see api.go).
// index.html is the public website, admin.html the admin console.
const backend = process.env.VITE_BACKEND || 'http://localhost:8080';

export default defineConfig({
  base: '/static/app/',
  plugins: [react()],
  build: {
    outDir: '../static/app',
    emptyOutDir: true,
    rollupOptions: { input: { index: 'index.html', admin: 'admin.html' } },
  },
  // `npm run dev` proxies data and images to a running Go server.
  server: {
    proxy: {
      '/api': backend,
      '/photos': backend,
      '/static/style.css': backend,
      '/static/console.css': backend,
      '/static/logo': backend,
    },
  },
});
