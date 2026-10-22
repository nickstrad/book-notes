import { readFileSync } from "node:fs";
import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";

const canvas = "[data-layout-ready]";
const picker = (page: Page) => page.getByRole("combobox", { name: "Architecture" });
const back = (page: Page) => page.getByRole("navigation", { name: "Diagram navigation" }).getByRole("button", { name: "Back" });
const fixture = (name: string) => readFileSync(new URL(`fixtures/${name}`, import.meta.url), "utf8");

// Every smoke slice rejects runtime errors, failed asset requests and wrong-base worker URLs.
test.beforeEach(async ({ page }) => {
  page.on("pageerror", (error) => { throw error; });
  page.on("console", (message) => {
    if (message.type() === "error") throw new Error(message.text());
  });
  page.on("requestfailed", (request) => {
    if (request.failure()?.errorText !== "net::ERR_ABORTED") throw new Error(`Failed request: ${request.url()}`);
  });
  page.on("worker", (worker) => expect(new URL(worker.url()).pathname).toMatch(/^\/book-notes\/assets\/layout-worker.*\.js$/));
});

test("default catalog renders with documentation at the selected base", async ({ page }) => {
  await page.goto("./");
  await expect(page).toHaveURL(/\/book-notes\/#\/architectures\/shell-example\/diagrams\/context$/);
  await expect(page.getByRole("heading", { level: 1, name: "System context: example shell test" })).toBeVisible();
  await expect(page.locator(canvas)).toHaveCount(1);
  await expect(picker(page).getByRole("option")).toHaveText([
    "Shell test: example architecture", "Shell test: alternate architecture",
  ]);
  await expect(page.getByText("Snapshot shell-fixture-1", { exact: false })).toBeVisible();
  await expect(page.getByText(/This invented example verifies/)).toBeVisible();
  await expect(page.locator("body")).not.toContainText("ts-db");
});

test("dropdown switches in both directions and browser history returns", async ({ page }) => {
  await page.goto("./");
  await expect(page.locator(canvas)).toHaveCount(1);
  await picker(page).selectOption("shell-alternate");
  await expect(page.getByRole("heading", { level: 1, name: "System context: alternate shell test" })).toBeVisible();
  await expect(page.locator(canvas)).toHaveCount(1);
  await back(page).click();
  await expect(picker(page)).toHaveValue("shell-example");
  await page.goForward();
  await expect(picker(page)).toHaveValue("shell-alternate");
  await picker(page).selectOption("shell-example");
  await expect(page.locator(canvas)).toHaveCount(1);
});

test("a non-default deep link reloads to the same diagram", async ({ page }) => {
  await page.goto("./#/architectures/shell-alternate/diagrams/context");
  await expect(page.locator(canvas)).toHaveCount(1);
  await page.reload();
  await expect(page.getByRole("heading", { level: 1, name: "System context: alternate shell test" })).toBeVisible();
  await expect(picker(page)).toHaveValue("shell-alternate");
  await expect(page.locator(canvas)).toHaveCount(1);
});

test("unknown and malformed routes offer recovery without a fallback", async ({ page }) => {
  for (const hash of ["#/architectures/nope/diagrams/context", "#/somewhere"]) {
    await page.goto(`./${hash}`);
    await expect(page.getByRole("heading", { name: "Not Found", exact: true })).toBeVisible();
    await expect(picker(page)).toHaveValue("");
    await expect(page.getByRole("link", { name: "Open Shell test: example architecture" })).toBeVisible();
  }
  await page.goto("./#/architectures/shell-example/diagrams/missing");
  await expect(page.getByText(/Not Found: unknown diagram/)).toBeVisible();
});

test("a third catalog entry works with no application change", async ({ page }) => {
  await page.route("**/catalog.json", (route) => route.fulfill({ contentType: "application/json", body: fixture("catalog-three.json") }));
  await page.route("**/architectures/fixture-third.json", (route) => route.fulfill({ contentType: "application/json", body: fixture("fixture-third.json") }));
  await page.goto("./");
  await expect(picker(page)).toHaveValue("fixture-third");
  await expect(page.locator(canvas)).toHaveCount(1);
  await expect(picker(page).getByRole("option")).toHaveCount(3);
  await picker(page).selectOption("shell-example");
  await expect(page.locator(canvas)).toHaveCount(1);
  await back(page).click();
  await expect(picker(page)).toHaveValue("fixture-third");
});

for (const [name, body, expected] of [
  ["empty catalog", '{"entries":[]}', "The diagrams could not be loaded"],
  ["malformed JSON", '{"entries":', "The diagrams could not be loaded"],
  ["missing defaultId", '{"defaultId":"missing","entries":[{"id":"shell-example","title":"Example","url":"architectures/shell-example.json"}]}', "The diagrams could not be loaded"],
] as const) {
  test(`reports ${name}`, async ({ page }) => {
    await page.route("**/catalog.json", (route) => route.fulfill({ contentType: "application/json", body }));
    await page.goto("./");
    await expect(page.getByRole("heading", { name: expected })).toBeVisible();
    await expect(page.locator(canvas)).toHaveCount(0);
  });
}

test("reports a failed catalog request", async ({ page }) => {
  // A deliberate HTTP failure logs a browser resource error; this test checks the user-facing result.
  page.removeAllListeners("console");
  await page.route("**/catalog.json", (route) => route.fulfill({ status: 503, body: "Unavailable" }));
  await page.goto("./");
  await expect(page.getByRole("alert")).toContainText("HTTP 503");
});

test("the skip link and architecture selector work by keyboard", async ({ page }) => {
  await page.goto("./");
  await expect(page.locator(canvas)).toHaveCount(1);
  await page.keyboard.press("Tab");
  await expect(page.getByRole("link", { name: "Skip to the diagrams" })).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(page.getByRole("main")).toBeFocused();
  await picker(page).focus();
  await page.keyboard.press("ArrowDown");
  await expect(picker(page)).toHaveValue("shell-alternate");
  await expect(page.locator(canvas)).toHaveCount(1);
});

for (const [width, height] of [[1600, 1000], [390, 844]]) {
  test(`shell layout and screenshot at ${width}×${height}`, async ({ page }, testInfo) => {
    await page.setViewportSize({ width, height });
    await page.goto("./");
    await expect(page.locator(canvas)).toHaveCount(1);
    await expect(picker(page)).toBeInViewport();
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
    await page.screenshot({ path: testInfo.outputPath(`shell-${width}.png`), fullPage: true });
    if (width < 600) {
      await page.getByRole("button", { name: "Open explanation", exact: true }).click();
      await expect(page.getByRole("region", { name: "Explanation", exact: true }).getByText(/This invented example verifies/)).toBeVisible();
      await page.screenshot({ path: testInfo.outputPath(`shell-${width}-explanation.png`), fullPage: true });
    }
  });
}
