import { describe, expect, it } from "vitest";
import type { ArchitectureBundle } from "c4-sdk/core";
import { bundles } from "./fixtures.ts";

describe.each(bundles)("shell provenance: $entry.id", ({ json }) => {
  const bundle = json as ArchitectureBundle;
  it("identifies invented data and pins its spec rather than asserting book fidelity", () => {
    expect(bundle.revision).toBe("shell-fixture-1");
    expect(bundle.sources.length).toBeGreaterThan(0);
    for (const source of bundle.sources) {
      expect(source.revision).toMatch(/^[0-9a-f]{40}$/);
      expect(source.path).toBe("docs/diagram-site/PLAN.md");
      expect(source.url).toBe(`https://github.com/nickstrad/book-notes/blob/${source.revision}/${source.path}`);
      expect(source.notes).toContain("not architecture evidence");
    }
    expect(bundle.documents[0].plain).toContain("invented example");
  });
});
