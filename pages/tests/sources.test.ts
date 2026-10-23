import { describe, expect, it } from "vitest";
import type { ArchitectureBundle } from "c4-sdk/core";
import { bundles, readJson } from "./fixtures.ts";

describe.each(bundles)("catalog provenance: $entry.id", ({ entry, json }) => {
  const bundle = json as ArchitectureBundle;
  const inventory = readJson(`content/${entry.id}/inventory.json`) as { revision: string };
  it("keeps commit-pinned public source links for every reference", () => {
    expect(bundle.revision).toContain(inventory.revision.slice(0, 7));
    expect(bundle.sources.length).toBeGreaterThan(0);
    for (const source of bundle.sources) {
      expect(source.revision).toBe(inventory.revision);
      expect(source.path).toMatch(/^books\//);
      expect(source.url).toMatch(new RegExp(`^https://github.com/nickstrad/book-notes/blob/${inventory.revision}/`));
      expect(source.url).toContain(`${source.path}#L`);
    }
  });
});
