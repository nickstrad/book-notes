import { describe, expect, it } from "vitest";
import { decodeArchitecture } from "c4-sdk/core";
import { selectArchitecture, settle } from "../src/site-model.ts";
import { loadArchitectures } from "../src/sdk.ts";
import { bundles } from "./fixtures.ts";

const entries = bundles.map(({ entry, json }) => ({ entry, result: decodeArchitecture(json) }));

describe("catalog settlement and view selection", () => {
  it("rejects an empty catalog and a missing default", () => {
    expect(settle([], undefined)).toEqual({ state: "failed", message: "The catalog lists no architectures." });
    expect(settle(entries, "missing").state).toBe("failed");
  });
  it("rejects the whole catalog when any bundle is invalid", () => {
    const broken = { entry: entries[0].entry, result: decodeArchitecture({}) };
    expect(settle([broken, entries[1]], undefined).state).toBe("failed");
  });
  it("derives snapshot and recovery links, and does not silently replace invalid routes", () => {
    const loaded = settle(entries, "shell-alternate");
    if (loaded.state !== "ready") throw new Error("Fixture failed to load");
    const selected = selectArchitecture(loaded, { kind: "home" });
    expect(selected.shown?.id).toBe("shell-alternate");
    expect(selected.snapshot).toBe("Snapshot shell-fixture-1");
    expect(selected.choices[0].href).toBe("#/architectures/shell-example/diagrams/context");
    expect(selectArchitecture(loaded, { kind: "not-found", message: "Bad hash" }))
      .toMatchObject({ shown: undefined, message: "Bad hash" });
    expect(selectArchitecture(loaded, { kind: "single", architectureId: "missing", diagramId: "context" }))
      .toMatchObject({ shown: undefined, message: 'Unknown architecture "missing".' });
  });
  it("rejects catalog/bundle identity mismatches", async () => {
    const fetchResource: typeof fetch = async (input) => new Response(JSON.stringify(
      String(input).endsWith("catalog.json")
        ? { entries: [{ id: "wrong-id", title: "Wrong", url: "bundle.json" }] }
        : bundles[0].json,
    ));
    const result = await loadArchitectures("https://site.test/book-notes/catalog.json", new AbortController().signal, fetchResource);
    expect(result[0].result).toMatchObject({ ok: false, diagnostics: [{ code: "catalog-identity" }] });
  });
});
