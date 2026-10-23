import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";
import { licenseNotices } from "./license-notices.ts";

export default defineConfig({
  plugins: [react(), licenseNotices()],
  // The viewer chunk is about 0.8 MB (c4-sdk bundles React Flow, Markdown and its dialog); the ELK layout
  // worker (about 2.1 MB) is a separate asset, not a chunk.
  build: { chunkSizeWarningLimit: 1600 },
  test: {
    include: ["tests/**/*.test.ts"],
    environment: "node",
  },
});
