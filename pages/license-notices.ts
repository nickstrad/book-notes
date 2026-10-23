import { readFileSync } from "node:fs";
import { join } from "node:path";
import type { Plugin } from "vite";

/**
 * Publishes license notices for every npm package bundled into the site, under `licenses/`. c4-sdk ships its
 * own MIT license and the notices for the 76 packages inside its bundles and ELK worker; both are copied
 * verbatim. The site itself bundles React, React DOM and scheduler, whose license texts go into NOTICES.txt.
 * The build fails if a bundled package is not covered, so a new dependency cannot ship without its notice.
 */
const read = (name: string, file: string) => readFileSync(join(import.meta.dirname, "node_modules", name, file), "utf8");
const version = (name: string) => (JSON.parse(read(name, "package.json")) as { version: string }).version;

/** Packages the site bundles directly; their license texts are reproduced in NOTICES.txt. */
const DIRECT = ["react", "react-dom", "scheduler"];
/** c4-sdk carries its own first-party license and third-party notices file. */
const SDK = "c4-sdk";

function notices(): string {
  const sections = DIRECT.map((name) => `## ${name}@${version(name)}\n\n${read(name, "LICENSE").trim()}\n`);
  return [
    "Book-notes architecture site — license notices",
    "",
    "The site's JavaScript bundles include the following npm packages.",
    "",
    `- ${SDK}@${version(SDK)} (MIT): see c4-sdk-LICENSE.txt. The packages bundled inside c4-sdk and its`,
    "  ELK layout worker (including elkjs, EPL-2.0 OR GPL-3.0-or-later) are listed with their license texts",
    "  in c4-sdk-THIRD_PARTY_NOTICES.txt.",
    ...DIRECT.map((name) => `- ${name}@${version(name)} (MIT): license text below.`),
    "",
    ...sections,
  ].join("\n");
}

export function licenseNotices(): Plugin {
  return {
    name: "book-notes-license-notices",
    apply: "build",
    generateBundle(_options, bundle) {
      const covered = new Set([SDK, ...DIRECT]);
      for (const output of Object.values(bundle)) {
        if (output.type !== "chunk") continue;
        for (const id of output.moduleIds) {
          const name = id.match(/node_modules\/((?:@[^/]+\/)?[^/]+)\//)?.[1];
          if (name && !covered.has(name)) this.error(`Bundled package ${name} has no published license notice`);
        }
      }
      const emit = (fileName: string, source: string) => this.emitFile({ type: "asset", fileName, source });
      emit("licenses/NOTICES.txt", notices());
      emit("licenses/c4-sdk-LICENSE.txt", read(SDK, "LICENSE"));
      emit("licenses/c4-sdk-THIRD_PARTY_NOTICES.txt", read(SDK, "THIRD_PARTY_NOTICES.md"));
    },
  };
}
