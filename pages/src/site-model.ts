import type { Diagnostic, LoadedArchitecture, ValidatedArchitecture } from "./sdk.ts";
import { chooseDefault, shownArchitectureId, singleHash } from "./routes.ts";
import type { SiteRoute } from "./routes.ts";

export interface Architecture {
  readonly id: string;
  readonly title: string;
  readonly validated: ValidatedArchitecture;
}
export type Loaded =
  | { readonly state: "loading" }
  | { readonly state: "failed"; readonly message: string }
  | { readonly state: "ready"; readonly architectures: readonly Architecture[]; readonly defaultId: string };

export function settle(entries: readonly LoadedArchitecture[], defaultId: unknown): Loaded {
  const problems: string[] = [];
  const architectures: Architecture[] = [];
  for (const { entry, result } of entries) {
    if (result.ok) architectures.push({ id: entry.id, title: entry.title, validated: result.value });
    else problems.push(`${entry.id}: ${describe(result.diagnostics)}`);
  }
  const choice = chooseDefault(entries.map(({ entry }) => entry.id), defaultId);
  if (!choice.ok) problems.push(choice.message);
  return problems.length || !choice.ok
    ? { state: "failed", message: problems.join(" ") }
    : { state: "ready", architectures, defaultId: choice.id };
}

function describe(diagnostics: readonly Diagnostic[]): string {
  return diagnostics.filter((d) => d.severity === "error")
    .map((d) => `${d.code} at ${d.path}: ${d.message}`).join("; ");
}

export interface Choice {
  readonly id: string;
  readonly title: string;
  readonly href: string;
}
export interface Selection {
  readonly choices: readonly Choice[];
  readonly shown?: Architecture;
  readonly message?: string;
  readonly snapshot?: string;
}

export function selectArchitecture(loaded: Extract<Loaded, { state: "ready" }>, route: SiteRoute): Selection {
  const shownId = shownArchitectureId(route, loaded.defaultId);
  const shown = loaded.architectures.find((architecture) => architecture.id === shownId);
  return {
    choices: loaded.architectures.map((a) => ({
      id: a.id, title: a.title, href: singleHash(a.id, a.validated.bundle.defaultDiagramId),
    })),
    shown,
    message: shown ? undefined : route.kind === "not-found" ? route.message : `Unknown architecture "${shownId}".`,
    snapshot: shown ? `Snapshot ${shown.validated.bundle.revision}` : undefined,
  };
}
