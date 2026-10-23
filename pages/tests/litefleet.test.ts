import { describe, expect, it } from 'vitest';
import { decodeArchitecture } from 'c4-sdk/core';
import type { ArchitectureBundle } from 'c4-sdk/core';
import { execFileSync } from 'node:child_process';
import { readJson } from './fixtures.ts';

const raw = readJson('public/architectures/litefleet.json') as ArchitectureBundle;
const decoded = decodeArchitecture(raw);
const bundle = decoded.ok ? decoded.value.bundle : raw;
const inventory = readJson('content/litefleet/inventory.json') as {
  revision: string;
  records: { id: string; type: string; boundary: string; protocol: string; sourceIds: string[]; revision: string; views: string[] }[];
  sources: { id: string; path: string; symbol: string; line: number; revision: string; evidence: string }[];
  lessons: { lesson: string; views: string[]; note: string; paths: string[]; revision: string }[];
};
const repo = new URL('../../', import.meta.url);
const pinnedText = (path: string) => execFileSync('git', ['show', `${inventory.revision}:${path}`], { cwd: repo, encoding: 'utf8' });
const viewIds = ['context', 'containers', 'gateway-components', 'node-components', 'query-flow', 'usage-flow', 'lab-deployment'];

describe('Litefleet hardened slice', () => {
  it('decodes all seven required views without diagnostics', () => {
    expect(decoded.diagnostics).toEqual([]);
    expect(decoded.ok).toBe(true);
    expect(bundle.id).toBe('litefleet');
    expect(bundle.defaultDiagramId).toBe('context');
    expect(bundle.diagrams.map(d => d.id)).toEqual(viewIds);
    expect(bundle.diagrams.map(d => d.kind)).toEqual(['context', 'container', 'component', 'component', 'dynamic', 'dynamic', 'deployment']);
  });

  it('pins the hardened snapshot independently of the moving tag', () => {
    expect(inventory.revision).toBe('936786399d7bab3ebbfa845b9a07df02b41ed363');
    expect(bundle.revision).toContain(inventory.revision.slice(0, 7));
  });

  it('covers every model ID exactly once in the audit and draws every ID', () => {
    const items = [...bundle.model.elements, ...bundle.model.relationships];
    expect(inventory.records.map(r => r.id).sort()).toEqual(items.map(i => i.id).sort());
    for (const item of items) {
      const record = inventory.records.find(r => r.id === item.id)!;
      expect(record.revision).toBe(inventory.revision);
      expect(record.boundary.length, item.id).toBeGreaterThan(0);
      expect(record.protocol.length, item.id).toBeGreaterThan(0);
      expect(record.sourceIds, item.id).toEqual(item.sourceIds);
      expect(item.sourceIds.length, item.id).toBeGreaterThan(0);
      const drawn = bundle.diagrams.filter(d => d.nodes.some(n => n.elementId === item.id) || d.edges.some(e => e.relationshipId === item.id)).map(d => d.id);
      expect(record.views, item.id).toEqual(drawn);
      expect(drawn.length, item.id).toBeGreaterThan(0);
    }
  });

  it('checks every source selector and exact line against the pinned git object', () => {
    const texts = new Map<string, string>();
    expect(inventory.sources.map(s => s.id).sort()).toEqual(bundle.sources.map(s => s.id).sort());
    for (const proof of inventory.sources) {
      const source = bundle.sources.find(s => s.id === proof.id)!;
      const text = texts.get(proof.path) ?? pinnedText(proof.path);
      texts.set(proof.path, text);
      expect(proof.revision).toBe(inventory.revision);
      expect(source.revision).toBe(inventory.revision);
      expect(source.path).toBe(proof.path);
      expect(source.url).toBe(`https://github.com/nickstrad/book-notes/blob/${inventory.revision}/${proof.path}#L${proof.line}`);
      const line = text.split('\n')[proof.line - 1];
      expect(line, proof.id).toContain(proof.symbol);
      expect(line.trim(), proof.id).toBe(proof.evidence);
    }
  });

  it('maps all twelve lessons to real pinned notes, source paths and views', () => {
    expect(inventory.lessons.map(l => l.lesson)).toEqual(Array.from({ length: 12 }, (_, i) => `L${String(i + 1).padStart(2, '0')}`));
    for (const lesson of inventory.lessons) {
      expect(lesson.revision).toBe(inventory.revision);
      expect(pinnedText(lesson.note)).toContain('#');
      for (const path of lesson.paths) expect(pinnedText(path).length, path).toBeGreaterThan(0);
      for (const view of lesson.views) expect(viewIds).toContain(view);
    }
  });

  it('keeps process boundaries and independent activity paths explicit', () => {
    const components = bundle.model.elements.filter(e => e.kind === 'component');
    expect(components).toHaveLength(12);
    expect(components.every(e => e.kind === 'component' && ['gateway', 'node'].includes(e.parentId))).toBe(true);
    const edges = new Map(bundle.model.relationships.map(r => [r.id, r]));
    expect([edges.get('heartbeat')?.sourceId, edges.get('heartbeat')?.targetId]).toEqual(['node', 'control']);
    expect([edges.get('n-sweep-meter')?.sourceId, edges.get('n-sweep-meter')?.targetId]).toEqual(['nd-outbox', 'meter']);
    expect(bundle.model.relationships.some(r => r.sourceId === 'meter' && r.targetId === 'registry')).toBe(false);
    expect(pinnedText('books/release-it/project/internal/nodeclient/client.go')).toContain('re.Code != "draining"');
    expect(pinnedText('books/release-it/project/internal/registry/registry.go')).toContain('recordUsage(ctx, tx, hb.Databases)');
    expect(bundle.documents.find(d => d.id === 'doc-usage-flow')?.plain).toContain('Abrupt process death can lose queued events');
  });

  it('provides real drill-down, Up links, explanations, keys and dynamic controls', () => {
    const context = bundle.diagrams[0];
    expect(context.nodes.find(n => n.elementId === 'litefleet')?.detail?.diagramId).toBe('containers');
    const containers = bundle.diagrams[1];
    for (const [element, detail] of [['gateway', 'gateway-components'], ['node', 'node-components']]) {
      expect(containers.nodes.find(n => n.elementId === element)?.detail?.diagramId).toBe(detail);
    }
    for (const diagram of bundle.diagrams) {
      if (diagram.id !== 'context') expect(diagram.parent).toBeDefined();
      expect(diagram.legend.entries.length).toBeGreaterThan(0);
      const numbers = diagram.edges.map(e => e.displayNumber);
      expect(new Set(numbers).size).toBe(numbers.length);
      for (const id of diagram.docIds) {
        const doc = bundle.documents.find(d => d.id === id)!;
        expect(doc.plain.length).toBeGreaterThan(100);
        expect(doc.ste.reviewStatus).toBe('draft');
        expect(doc.ste.issue).toBe('ASD-STE100 Issue 9');
        expect(doc.sourceIds.length).toBeGreaterThan(0);
      }
      if (diagram.kind === 'dynamic') {
        expect(diagram.scenario.kind).toBe('sequence');
        if (diagram.scenario.kind === 'sequence') expect(diagram.scenario.children.length).toBeGreaterThan(5);
      }
    }
  });

  it('models real Compose replicas, separate volumes and all profile tools', () => {
    const ids = bundle.diagrams.find(d => d.id === 'lab-deployment')!.nodes.map(n => n.elementId);
    for (const id of ['gateway-a', 'gateway-b', 'gateway-c', 'node-a', 'node-b', 'pgdata', 'node-a-data', 'node-b-data', 'blobs', 'migrate', 'purger', 'seedgen', 'loadgen', 'report', 'harness', 'toxiproxy', 'prometheus', 'exporter']) expect(ids).toContain(id);
    const replicas = bundle.model.elements.filter(e => e.kind === 'instance');
    expect(replicas.filter(e => e.kind === 'instance' && e.instanceOf === 'gateway')).toHaveLength(3);
    expect(replicas.filter(e => e.kind === 'instance' && e.instanceOf === 'node')).toHaveLength(2);
  });
});
