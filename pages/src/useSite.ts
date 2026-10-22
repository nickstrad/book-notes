import { useEffect, useState, useSyncExternalStore } from "react";
import type { MouseEvent } from "react";
import { browserRoutePort, loadArchitectures } from "./sdk.ts";
import { liftDefaultId } from "./catalog.ts";
import { parseSiteRoute } from "./routes.ts";
import { selectArchitecture, settle } from "./site-model.ts";
import type { Loaded } from "./site-model.ts";

const CATALOG_URL = `${import.meta.env.BASE_URL}catalog.json`;
function subscribeHash(listener: () => void): () => void {
  return browserRoutePort().subscribe(listener);
}
function useArchitectures(): Loaded {
  const [loaded, setLoaded] = useState<Loaded>({ state: "loading" });
  useEffect(() => {
    const abort = new AbortController();
    const catalog = liftDefaultId(CATALOG_URL);
    loadArchitectures(CATALOG_URL, abort.signal, catalog.fetch).then(
      (entries) => {
        if (!abort.signal.aborted) setLoaded(settle(entries, catalog.defaultId()));
      },
      (error: unknown) => {
        if (!abort.signal.aborted) setLoaded({
          state: "failed", message: error instanceof Error ? error.message : String(error),
        });
      },
    );
    return () => abort.abort();
  }, []);
  return loaded;
}

export function useSite() {
  const hash = useSyncExternalStore(subscribeHash, () => location.hash, () => "");
  const loaded = useArchitectures();
  const selection = loaded.state === "ready" ? selectArchitecture(loaded, parseSiteRoute(hash)) : undefined;
  const choose = (id: string) => {
    if (loaded.state !== "ready") return;
    const next = loaded.architectures.find((architecture) => architecture.id === id);
    if (next) browserRoutePort().push({
      architectureId: next.id, diagramId: next.validated.bundle.defaultDiagramId,
    });
  };
  const skipToMain = (event: MouseEvent<HTMLAnchorElement>) => {
    event.preventDefault();
    document.getElementById("site-main")?.focus();
  };
  return { loaded, selection, choose, skipToMain };
}
