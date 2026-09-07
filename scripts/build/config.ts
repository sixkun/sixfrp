import { join, resolve } from "node:path";

import { commandOutput } from "./command";
import type { BuildConfig, BuildSpec, BuildTarget, Platform } from "./types";
import { formatBuildTime, normalizeVersion, parsePlatforms } from "./utils";

export function createConfig(scriptDir: string): BuildConfig {
  const projectDir = resolve(scriptDir, "..");
  const rawVersion =
    process.env.VERSION || commandOutput("git", ["describe", "--tags", "--always", "--dirty"], projectDir) || "dev";
  const version = normalizeVersion(rawVersion);
  const releaseTag = process.env.RELEASE_TAG || version;

  return {
    scriptDir,
    projectDir,
    version,
    buildTime: formatBuildTime(new Date()),
    pubDate: new Date().toISOString(),
    summaryVersion: version.replace(/^v/, ""),
    releaseBaseUrl:
      process.env.RELEASE_BASE_URL || `https://cnb.cool/sixkun/sixfrp/-/releases/download/${releaseTag}`,
    requestedPlatforms: parsePlatforms(process.env.PLATFORMS) || getDefaultPlatforms(),
  };
}

export function createSpecs(config: BuildConfig): Record<BuildTarget, BuildSpec> {
  return {
    frppc: {
      target: "frppc",
      title: "frppc Multi-Platform Build",
      sourceDir: join(config.projectDir, "cmd", "frppc"),
      buildDir: join(config.projectDir, "build", "frppc"),
      outputPrefix: "frppc",
    },
  };
}

// frppc: 全平台支持（包括各种架构和操作系统）
function getDefaultPlatforms(): Platform[] {
  return [
    // Darwin (macOS)
    {goos: "darwin", goarch: "amd64"},
    {goos: "darwin", goarch: "arm64"},
    // FreeBSD
    {goos: "freebsd", goarch: "amd64"},
    // Linux (主流架构)
    {goos: "linux", goarch: "amd64"},
    {goos: "linux", goarch: "arm64"},
    {goos: "linux", goarch: "arm", suffix: "arm"}, // ARMv5/v6 软浮点
    {goos: "linux", goarch: "arm", goarm: "7", suffix: "armv7"}, // ARMv7 硬浮点
    {goos: "linux", goarch: "loong64"},
    // Linux (MIPS 系列)
    {goos: "linux", goarch: "mips"},
    {goos: "linux", goarch: "mips64"},
    {goos: "linux", goarch: "mips64le"},
    {goos: "linux", goarch: "mipsle"},
    // Linux (RISC-V)
    {goos: "linux", goarch: "riscv64"},
    // OpenBSD
    {goos: "openbsd", goarch: "amd64"},
    // Windows
    {goos: "windows", goarch: "amd64"},
    {goos: "windows", goarch: "arm64"},
  ];
}
