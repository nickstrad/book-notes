import type { MouseEvent, ReactNode } from "react";
import {
  ArchitectureProvider, browserRoutePort, DiagramDocumentation,
  DiagramNavigation, DiagramView, layoutPort,
} from "./sdk.ts";
import type { Choice, Loaded, Selection } from "./site-model.ts";

const REPO = "https://github.com/nickstrad/book-notes";
interface SiteProps {
  readonly loaded: Loaded;
  readonly selection?: Selection;
  readonly choose: (id: string) => void;
  readonly skipToMain: (event: MouseEvent<HTMLAnchorElement>) => void;
}

/** Props-only shell. Catalog loading, routes, selection and status text live outside this view. */
export function Site({ loaded, selection, choose, skipToMain }: SiteProps) {
  return <>
    <a className="skip-link" href="#site-main" onClick={skipToMain}>Skip to the diagrams</a>
    <header className="site-header">
      <div className="site-heading">
        <p className="site-eyebrow">C4 model</p>
        <p className="site-title">Book-notes architecture</p>
        <p className="site-purpose">
          Choose an architecture, then explore its diagrams and explanations.
          Each dataset identifies the snapshot it describes.
        </p>
        <p className="site-links">
          <a href={REPO}>Source and lesson notes</a>
          <span aria-hidden="true">·</span>
          <a href={`${REPO}/blob/main/docs/diagram-site/PLAN.md`}>Diagram site plan</a>
        </p>
      </div>
    </header>
    {loaded.state === "loading" && <SiteMain><p role="status">Loading the architecture catalog…</p></SiteMain>}
    {loaded.state === "failed" && <SiteMain>
      <section role="alert" className="site-panel">
        <h2>The diagrams could not be loaded</h2><p>{loaded.message}</p>
      </section>
    </SiteMain>}
    {selection && (selection.shown
      ? <ArchitectureProvider key={selection.shown.id}
        validatedBundle={selection.shown.validated} routePort={browserRoutePort()} layoutPort={layoutPort}>
        <SiteBar choices={selection.choices} selectedId={selection.shown.id} choose={choose}>
          <DiagramNavigation />
        </SiteBar>
        <SiteMain>
          <section aria-label={selection.shown.title}>
            <p className="dataset-note"><strong>{selection.shown.title}.</strong> {selection.snapshot}</p>
            <div className="c4-viewer-grid"><div><DiagramView /></div><DiagramDocumentation /></div>
          </section>
        </SiteMain>
      </ArchitectureProvider>
      : <>
        <SiteBar choices={selection.choices} selectedId="" choose={choose} />
        <SiteMain><section role="alert" className="site-panel">
          <h2>Not Found</h2><p>{selection.message}</p>
          <ul>{selection.choices.map((choice) => <li key={choice.id}>
            <a href={choice.href}>Open {choice.title}</a>
          </li>)}</ul>
        </section></SiteMain>
      </>)}
  </>;
}

function SiteMain({ children }: { readonly children: ReactNode }) {
  return <main id="site-main" tabIndex={-1} className="site-main">{children}</main>;
}
function SiteBar({ choices, selectedId, choose, children }: {
  readonly choices: readonly Choice[];
  readonly selectedId: string;
  readonly choose: (id: string) => void;
  readonly children?: ReactNode;
}) {
  return <div className="site-bar"><div className="site-bar-inner">
    <div className="architecture-picker">
      <label htmlFor="architecture-select">Architecture</label>
      <select id="architecture-select" value={selectedId} onChange={(event) => choose(event.target.value)}>
        {!selectedId && <option value="" disabled>Choose an architecture</option>}
        {choices.map((choice) => <option key={choice.id} value={choice.id}>{choice.title}</option>)}
      </select>
    </div>
    {children}
  </div></div>;
}
