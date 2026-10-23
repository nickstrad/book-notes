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
  readonly snapshot?: Snapshot;
}
/** The dataset's revision label, plus a link to the pinned commit when its sources name one on GitHub. */
export interface Snapshot {
  readonly label: string;
  readonly commitUrl?: string;
}

export function snapshotOf({ bundle }: ValidatedArchitecture): Snapshot {
  const short = bundle.revision.match(/\b[0-9a-f]{7,40}\b/)?.[0];
  for (const source of bundle.sources) {
    const pinned = source.url.match(/^(https:\/\/github\.com\/[^/]+\/[^/]+)\/blob\/([0-9a-f]{40})\//);
    if (short && pinned?.[2].startsWith(short)) return { label: bundle.revision, commitUrl: `${pinned[1]}/commit/${pinned[2]}` };
  }
  return { label: bundle.revision };
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
    snapshot: shown ? snapshotOf(shown.validated) : undefined,
  };
}
