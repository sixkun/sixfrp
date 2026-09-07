import { spawnSync } from "node:child_process";

export function commandOutput(command: string, args: string[], cwd: string) {
  const result = spawnSync(command, args, {
    cwd,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "ignore"],
  });

  if (result.status !== 0) {
    return "";
  }

  return result.stdout.trim();
}

export function run(
  command: string,
  args: string[],
  cwd: string,
  env: Record<string, string> = {},
  throwOnSpawnError = true,
) {
  const result = spawnSync(command, args, {
    cwd,
    env: {...process.env, ...env},
    stdio: "inherit",
  });

  if (throwOnSpawnError && result.error) {
    throw result.error;
  }

  return result;
}

export function runOrThrow(command: string, args: string[], cwd: string, env: Record<string, string> = {}) {
  const result = run(command, args, cwd, env);

  if (result.status !== 0) {
    throw new Error(`Command failed: ${[command, ...args].join(" ")}`);
  }
}
