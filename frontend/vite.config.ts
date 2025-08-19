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
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
  },
})