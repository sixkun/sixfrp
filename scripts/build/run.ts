import { build } from "./builder";
import { createConfig, createSpecs } from "./config";
import { fail } from "./console";
import type { BuildTarget } from "./types";

export async function runTargets(scriptDir: string, targets: BuildTarget[]) {
  try {
    for (const target of targets) {
      const config = createConfig(scriptDir);
      const specs = createSpecs(config);

      const succeeded = await build(config, specs[target]);
      if (!succeeded) {
        process.exit(1);
      }
    }
  } catch (error) {
    fail(error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
}
