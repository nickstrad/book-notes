import { defineConfig, devices } from "@playwright/test";

// Smoke test of the production build (run `npm run build` first; `npm run check` does). It serves dist/ with
// `vite preview` at the GitHub Pages base path, so asset and worker URLs are exercised under /book-notes/.
const port = 4327;

export default defineConfig({
  testDir: "tests",
  testMatch: "**/*.spec.ts",
  // Keep generated browser artifacts in the ignored dependency cache.
  outputDir: "node_modules/.cache/playwright-results",
  timeout: 60_000,
  workers: 1,
  reporter: "line",
  use: { baseURL: `http://127.0.0.1:${port}/book-notes/` },
  projects: [{
    name: "chromium",
    use: {
      ...devices["Desktop Chrome"],
      viewport: { width: 1600, height: 1000 },
    },
  }],
  webServer: {
    command: `npx vite preview --base /book-notes/ --host 127.0.0.1 --port ${port} --strictPort`,
    url: `http://127.0.0.1:${port}/book-notes/`,
    reuseExistingServer: false,
  },
});
