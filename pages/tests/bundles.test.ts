import { describe, expect, it } from "vitest";
import { decodeArchitecture } from "c4-sdk/core";
import type { ArchitectureBundle } from "c4-sdk/core";
import { bundles, catalog } from "./fixtures.ts";

describe("catalog", () => {
  it("contains only explicitly labeled shell test data", () => {
    expect(catalog.entries.map((entry) => entry.id)).toEqual(["shell-example", "shell-alternate"]);
    expect(catalog.defaultId).toBe("shell-example");
    expect(catalog.entries.every((entry) => entry.title.startsWith("Shell test:"))).toBe(true);
  });
});

describe.each(bundles)("bundle $entry.id", ({ entry, path, json }) => {
  const decoded = decodeArchitecture(json);
  const bundle = decoded.ok ? decoded.value.bundle : (json as ArchitectureBundle);

  it("decodes with decodeArchitecture and no diagnostics at all", () => {
    // Errors reject the bundle; warnings (missing protocol, incomplete key, large view) are treated as failures
    // too, as c4-sdk's Milestone B gate does.
    expect(decoded.diagnostics, JSON.stringify(decoded.diagnostics, null, 2))
      .toEqual([]);
    expect(decoded.ok).toBe(true);
  });

  it("has the id of its catalog entry", () => {
    expect(bundle.id).toBe(entry.id);
  });

  it("draws every model element and relationship in at least one diagram", () => {
    const drawnElements = new Set(
      bundle.diagrams.flatMap((diagram) => diagram.nodes.map((node) => node.elementId)),
    );
    const drawnRelationships = new Set(
      bundle.diagrams.flatMap((diagram) => diagram.edges.map((edge) => edge.relationshipId)),
    );
    expect(
      bundle.model.elements.filter((element) => !drawnElements.has(element.id))
        .map((e) => e.id),
    ).toEqual([]);
    expect(
      bundle.model.relationships.filter((r) => !drawnRelationships.has(r.id))
        .map((r) => r.id),
    ).toEqual([]);
  });

  it("numbers every edge with a unique source edge number per diagram", () => {
    for (const diagram of bundle.diagrams) {
      const numbers = diagram.edges.map((edge) => edge.displayNumber);
      expect(numbers.every((n) => Number.isInteger(n)), diagram.id).toBe(true);
      expect(new Set(numbers).size, diagram.id).toBe(numbers.length);
    }
  });

  it("gives every diagram a plain and an STE draft document", () => {
    const documents = new Map(
      bundle.documents.map((document) => [document.id, document]),
    );
    for (const diagram of bundle.diagrams) {
      expect(diagram.docIds?.length, diagram.id).toBeGreaterThan(0);
      for (const id of diagram.docIds ?? []) {
        const document = documents.get(id)!;
        expect(document.plain.length, id).toBeGreaterThan(0);
        expect(document.ste.text.length, id).toBeGreaterThan(0);
        expect(document.ste.issue, id).toBe("ASD-STE100 Issue 9");
      }
    }
    expect(
      bundle.documents.every((document) => document.ste.reviewStatus === "draft"),
    ).toBe(true);
  });

  it("uses the catalog-relative architecture path", () => {
    expect(path).toBe(`public/architectures/${entry.id}.json`);
  });
});
