#!/usr/bin/env bun

import { runTargets } from "./build/run";

const target = process.argv[2];

if (target !== "frppc") {
  console.error("Usage: bun run scripts/build-target.ts frppc");
  process.exit(1);
}

await runTargets(import.meta.dir, ["frppc"]);
