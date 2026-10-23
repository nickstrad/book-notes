import { describe, expect, it } from 'vitest';
import { decodeArchitecture } from 'c4-sdk/core';
import type { ArchitectureBundle } from 'c4-sdk/core';
import { execFileSync } from 'node:child_process';
import { readJson } from './fixtures.ts';

const raw = readJson('public/architectures/queuelet.json') as ArchitectureBundle;
const decoded = decodeArchitecture(raw);
const bundle = decoded.ok ? decoded.value.bundle : raw;
const inventory = readJson('content/queuelet/inventory.json') as {
  revision: string;
  records: { id: string; type: string; boundary: string; protocol: string; sourceIds: string[]; revision: string; views: string[] }[];
  sources: { id: string; path: string; symbol: string; line: number; revision: string; evidence: string }[];
  lessons: { lesson: string; views: string[]; note: string; paths: string[]; revision: string }[];
};
const repo = new URL('../../', import.meta.url);
const pinnedText = (path: string) => execFileSync('git', ['show', `${inventory.revision}:${path}`], { cwd: repo, encoding: 'utf8' });
const viewIds = ['context', 'containers', 'api-components', 'job-flow', 'lab-deployment'];

describe('Queuelet explored slice', () => {
  it('decodes all five required views without diagnostics', () => {
    expect(decoded.diagnostics).toEqual([]);
    expect(decoded.ok).toBe(true);
    expect(bundle.id).toBe('queuelet');
    expect(bundle.defaultDiagramId).toBe('context');
    expect(bundle.diagrams.map(d => d.id)).toEqual(viewIds);
    expect(bundle.diagrams.map(d => d.kind)).toEqual(['context', 'container', 'component', 'dynamic', 'deployment']);
  });

  it('pins the explored snapshot independently of the moving tag', () => {
    expect(inventory.revision).toBe('b6df7db044782dd3517d041a8aad95308a3a80c9');
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

  it('maps all nine experiments to real pinned notes, source paths and views', () => {
    expect(inventory.lessons.map(l => l.lesson)).toEqual(Array.from({ length: 9 }, (_, i) => `E${String(i + 1).padStart(2, '0')}`));
    for (const lesson of inventory.lessons) {
      expect(lesson.revision).toBe(inventory.revision);
      expect(pinnedText(lesson.note)).toContain('#');
      for (const path of lesson.paths) expect(pinnedText(path).length, path).toBeGreaterThan(0);
      for (const view of lesson.views) expect(viewIds).toContain(view);
    }
  });

  it('keeps scheduler, sweeper, shipper, clock and observations inside API', () => {
    const components = bundle.model.elements.filter(e => e.kind === 'component');
    expect(components).toHaveLength(8);
    expect(components.every(e => e.kind === 'component' && e.parentId === 'api')).toBe(true);
    expect(bundle.model.elements.filter(e => e.kind === 'container').map(e => e.id).sort()).toEqual(['api', 'hooksink', 'postgres', 'qctl', 'worker']);
    expect(bundle.model.elements.filter(e => e.kind === 'instance').map(e => e.id).sort()).toEqual(['api-run', 'hook-run', 'pg-run', 'worker-a', 'worker-b']);
    expect(pinnedText('books/explore-it/project/internal/api/api.go')).toContain('go func()');
    expect(pinnedText('books/explore-it/project/internal/jobs/ack.go')).toContain('notify(tx, id, now)');
    expect(pinnedText('books/explore-it/project/internal/jobs/ack.go')).toContain('return tx.Commit()');
    expect(bundle.documents.find(d => d.id === 'doc-job-flow')?.plain).toContain('not exactly-once');
  });

  it('provides real context/container/component navigation and dynamic steps', () => {
    expect(bundle.diagrams[0].nodes.find(n => n.elementId === 'queuelet')?.detail?.diagramId).toBe('containers');
    expect(bundle.diagrams[1].nodes.find(n => n.elementId === 'api')?.detail?.diagramId).toBe('api-components');
    for (const diagram of bundle.diagrams) {
      if (diagram.id !== 'context') expect(diagram.parent).toBeDefined();
      expect(diagram.legend.entries.length).toBeGreaterThan(0);
      const numbers = diagram.edges.map(e => e.displayNumber);
      expect(numbers.every(Number.isInteger)).toBe(true);
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
        if (diagram.scenario.kind === 'sequence') {
          expect(diagram.scenario.children).toHaveLength(16);
          const steps = diagram.scenario.children;
          expect(steps[4]).toMatchObject({ relationshipId: 'f-lease', direction: 'reverse' });
          expect(steps[7]).toMatchObject({ relationshipId: 'f-terminal' });
          expect(steps[8]).toMatchObject({ relationshipId: 'f-nack' });
          expect(steps[12]).toMatchObject({ relationshipId: 'f-outbox', direction: 'reverse' });
        }
      }
    }
  });
});
