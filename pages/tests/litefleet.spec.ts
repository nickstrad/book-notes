import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';
import { readFileSync } from 'node:fs';

const bundle = JSON.parse(readFileSync(new URL('../public/architectures/litefleet.json', import.meta.url), 'utf8')) as {
  title: string; diagrams: { id: string; title: string; kind: string }[];
};
const ready = (page: Page) => expect(page.locator('[data-layout-ready]')).toHaveCount(1);
const nav = (page: Page) => page.getByRole('navigation', { name: 'Diagram navigation' });

// Item 2 owns a slice, not the production catalog (integration is item 4).
// Only catalog loading is replaced; the bundle/assets/worker come from the production build.
test.beforeEach(async ({ page }) => {
  await page.route('**/catalog.json', route => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ defaultId: 'litefleet', entries: [{ id: 'litefleet', title: bundle.title, url: 'architectures/litefleet.json' }] }) }));
  page.on('pageerror', error => { throw error; });
  page.on('console', message => { if (message.type() === 'error') throw new Error(message.text()); });
  page.on('requestfailed', request => { if (request.failure()?.errorText !== 'net::ERR_ABORTED') throw new Error(request.url()); });
  page.on('worker', worker => expect(new URL(worker.url()).pathname).toMatch(/^\/book-notes\/assets\/layout-worker.*\.js$/));
});

for (const diagram of bundle.diagrams) {
  test(`litefleet route ${diagram.id} renders and reloads`, async ({ page }) => {
    await page.goto(`./#/architectures/litefleet/diagrams/${diagram.id}`);
    await ready(page);
    await expect(page.getByRole('heading', { level: 1, name: diagram.title, exact: true })).toBeVisible();
    await expect(page.getByRole('heading', { name: diagram.title + ' explained', exact: true })).toBeVisible();
    await page.reload();
    await ready(page);
    await expect(page.getByRole('heading', { level: 1, name: diagram.title, exact: true })).toBeVisible();
    if (diagram.kind === 'dynamic') {
      const select = page.getByRole('combobox', { name: 'Scenario step' });
      await page.getByRole('button', { name: 'Next selection', exact: true }).click();
      await expect(select).not.toHaveValue('');
      const first = await select.inputValue();
      await page.getByRole('button', { name: 'Next selection', exact: true }).click();
      await expect(select).not.toHaveValue(first);
      await page.getByRole('button', { name: 'Previous selection', exact: true }).click();
      await expect(select).toHaveValue(first);
    }
  });
}

test('litefleet actual element drill-down, Up and Back', async ({ page }) => {
  await page.goto('./');
  await ready(page);
  await page.getByRole('button', { name: 'STE100 text', exact: true }).click();
  await expect(page.getByText('ASD-STE100 Issue 9 · draft', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Plain explanation', exact: true }).click();
  await page.getByRole('link', { name: 'Open detail for Litefleet', exact: true }).focus();
  await page.getByRole('link', { name: 'Open detail for Litefleet', exact: true }).click();
  await expect(page).toHaveURL(/diagrams\/containers$/);
  await ready(page);
  await page.getByRole('link', { name: 'Open detail for Gateway', exact: true }).focus();
  await page.getByRole('link', { name: 'Open detail for Gateway', exact: true }).click();
  await expect(page).toHaveURL(/diagrams\/gateway-components$/);
  await ready(page);
  await nav(page).getByRole('button', { name: 'Back', exact: true }).click();
  await expect(page).toHaveURL(/diagrams\/containers$/);
  await ready(page);
  await page.getByRole('link', { name: 'Open detail for Node agent', exact: true }).focus();
  await page.getByRole('link', { name: 'Open detail for Node agent', exact: true }).click();
  await expect(page).toHaveURL(/diagrams\/node-components$/);
  await ready(page);
  await nav(page).getByRole('link', { name: 'Up', exact: true }).click();
  await expect(page).toHaveURL(/diagrams\/containers$/);
});

for (const [width, height] of [[1600, 1000], [390, 844]]) {
  test(`litefleet all view screenshots at ${width}x${height}`, async ({ page }, testInfo) => {
    test.setTimeout(180_000);
    await page.setViewportSize({ width, height });
    for (const diagram of bundle.diagrams) {
      await page.goto(`./#/architectures/litefleet/diagrams/${diagram.id}`);
      await ready(page);
      await expect(page.getByRole('combobox', { name: 'Architecture' })).toBeInViewport();
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
      await page.evaluate(() => window.scrollTo(0, 0));
      await page.screenshot({ path: testInfo.outputPath(`litefleet-${diagram.id}-${width}.png`) });
      await page.getByRole('button', { name: 'Fit View', exact: true }).click();
      await page.locator('[data-layout-ready]').screenshot({ path: testInfo.outputPath(`litefleet-${diagram.id}-${width}-canvas.png`) });
      await page.evaluate(() => window.scrollTo(0, 0));
      await page.screenshot({ path: testInfo.outputPath(`litefleet-${diagram.id}-${width}-overview.png`), fullPage: true });
      if (width < 600) {
        await page.getByRole('button', { name: 'Open explanation', exact: true }).click();
        await expect(page.getByRole('region', { name: 'Explanation', exact: true }).getByRole('heading', { name: diagram.title + ' explained', exact: true })).toBeVisible();
        await page.screenshot({ path: testInfo.outputPath(`litefleet-${diagram.id}-${width}-docs.png`), fullPage: false });
        await page.getByRole('button', { name: 'Close explanation', exact: true }).click();
      }
    }
  });
}
