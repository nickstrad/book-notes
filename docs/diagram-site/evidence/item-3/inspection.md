# Item 3 visual inspection

Queuelet's five views were inspected at 1600×1000 and 390×844 using the slice's
Playwright captures. The final full check captures the same diagrams with additional
supporting model citations in the source lists. Retained screenshots are copied
from that full check, not from a preview server or an invented fixture bundle.

- Context: producer, worker role, operator and external receiver are distinct.
  The Queuelet detail link works. At native scale the initial canvas shows part
  of the topology; Fit shows the whole graph. The document explains the pending
  end review and rejects complete-exploration / exactly-once claims.
- Containers: API and Postgres share the service group. Workers and qctl are
  separate clients. Receiver contract and hooksink fake have distinct labels;
  the explanation states these are alternative abstraction levels, not two sends.
  API detail opens the actual component view.
- Components: the API process boundary contains all eight responsibilities.
  Scheduler and shipper are not services. At Fit scale the structure is clear;
  pan/zoom is required for comfortable reading of the 17 edge labels.
- Dynamic: the canvas shows operation edges, while scenario controls select the
  causal steps. Success and alternative nack/expiry/cancel/retry paths are
  explicitly labelled and explained. The step sequence includes lease response,
  worker execution, terminal transaction, outbox read, POST and delivery result.
- Deployment: host and child Compose boundaries distinguish external qctl/probes
  from five services. Both worker queues, persistent volume and API clock config
  are present. Physical-tree text is available below the canvas.
- Narrow screens: architecture selector and navigation remain visible and the
  document does not overflow horizontally. Native-scale diagrams require pan;
  Fit is an overview, not a way to read every tiny label. Explanation drawers
  have readable paragraphs, source links and close controls. Long documents
  scroll vertically. The snapshot uses tag + short SHA; provenance uses full SHA.

No new layout or rendering exception was required. This is a visual review of
rendering/navigation, not independent code-fidelity approval (item 6).
