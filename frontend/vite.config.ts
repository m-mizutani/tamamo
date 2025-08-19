import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import fs from 'fs'

// Custom plugin to preserve .gitkeep file
const preserveGitkeep = () => {
  return {
    name: 'preserve-gitkeep',
    writeBundle() {
      const gitkeepPath = path.resolve(__dirname, 'dist/.gitkeep')
      // Create .gitkeep after all files are written
      fs.writeFileSync(gitkeepPath, '')
      console.log('✓ Created dist/.gitkeep')
    }
  }
}

export default defineConfig({
  plugins: [react(), preserveGitkeep()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    allowedHosts: [
      ".ngrok-free.app",
      ".ts.net"
    ],
    proxy: {
      "/graphql": "http://localhost:8080",
      "/api": "http://localhost:8080",
      "/hooks": "http://localhost:8080",
      "/health": "http://localhost:8080",
      "/ws": {
        target: "http://localhost:8080",
        ws: true,
        changeOrigin: true,
        rewrite: (path) => path,
      },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    // Ensure proper handling of SPA routing
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ["react", "react-dom"],
          ui: [
            "@radix-ui/react-alert-dialog",
            "@radix-ui/react-avatar", 
            "@radix-ui/react-dialog",
            "@radix-ui/react-select",
            "@radix-ui/react-separator",
            "@radix-ui/react-label",
          ],
        },
      },
    },
  },
})