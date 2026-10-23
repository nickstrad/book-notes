import assert from 'node:assert/strict';
import { readFileSync, readdirSync, statSync } from 'node:fs';

// Publication boundary: only Vite assets, the two real datasets, catalog, HTML and license notices.
const files = readdirSync('dist', { recursive: true }).filter(path => statSync(`dist/${path}`).isFile()).sort();
const licenses = ['licenses/NOTICES.txt', 'licenses/c4-sdk-LICENSE.txt', 'licenses/c4-sdk-THIRD_PARTY_NOTICES.txt'];
const fixed = ['index.html', '404.html', 'catalog.json', 'architectures/litefleet.json', 'architectures/queuelet.json', ...licenses];
for (const path of fixed) assert(files.includes(path), `Missing ${path}`);
for (const path of files) {
  assert(fixed.includes(path) || /^assets\/[^/]+\.(js|css)$/.test(path), `Unexpected published file: ${path}`);
}
const read = path => readFileSync(`dist/${path}`, 'utf8');
assert.equal(read('404.html'), read('index.html'));
const catalog = JSON.parse(read('catalog.json'));
assert.equal(catalog.defaultId, 'litefleet');
assert.deepEqual(catalog.entries.map(entry => entry.id), ['litefleet', 'queuelet']);
for (const [id, count] of [['litefleet', 7], ['queuelet', 5]]) {
  const entry = catalog.entries.find(entry => entry.id === id);
  assert.equal(entry.url, `architectures/${id}.json`);
  const bundle = JSON.parse(read(entry.url));
  assert.equal(bundle.id, id);
  assert.equal(bundle.diagrams.length, count);
}
assert.match(read('licenses/c4-sdk-LICENSE.txt'), /MIT License/);
assert.match(read('licenses/c4-sdk-THIRD_PARTY_NOTICES.txt'), /`elkjs@0\.12\.0` — EPL-2\.0 OR GPL-3\.0-or-later/);
for (const name of ['react', 'react-dom', 'scheduler']) assert.match(read('licenses/NOTICES.txt'), new RegExp(`^## ${name}@`, 'm'));
const references = [...read('index.html').matchAll(/(?:src|href)="([^"]+)"/g)].map(match => match[1]);
assert(references.length > 0);
for (const url of references) {
  assert(url.startsWith('/book-notes/assets/'), `Wrong asset base: ${url}`);
  assert(files.includes(url.slice('/book-notes/'.length)), `Missing asset: ${url}`);
}
const workers = files.filter(path => /worker.*\.js$/.test(path));
assert(workers.length > 0, 'Missing layout worker');
const app = files.filter(path => path.endsWith('.js') && !workers.includes(path)).map(read).join('\n');
for (const path of workers) assert(app.includes(`/book-notes/${path}`), `Wrong worker base: ${path}`);
console.log('PASS: Litefleet 7 views, Queuelet 5 views; HTML/assets/worker under /book-notes/; license notices present; dist-only publication manifest');
console.log(files.join('\n'));
