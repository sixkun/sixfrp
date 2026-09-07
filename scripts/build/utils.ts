import { createHash } from "node:crypto";
import { readdirSync, readFileSync } from "node:fs";

import type { Platform } from "./types";

export function outputFilename(prefix: string, platform: Platform) {
  // 如果有自定义 suffix，直接使用
  if (platform.suffix) {
    const suffix = platform.goos === "windows" ? ".exe" : "";
    return `${prefix}-${platform.goos}-${platform.suffix}${suffix}`;
  }

  // 标准命名: Node.js-style arch (amd64 -> x64), GOOS-style os (windows/linux/darwin)
  let arch = platform.goarch;

  // 特殊处理 amd64 -> x64
  if (platform.goarch === "amd64") {
    arch = "x64";
  }

  const exeSuffix = platform.goos === "windows" ? ".exe" : "";

  return `${prefix}-${platform.goos}-${arch}${exeSuffix}`;
}

export function platformFromArchive(filename: string): Platform {
  const name = filename.replace(/\.tar\.gz$/, "").replace(/\.exe$/, "");
  const parts = name.split("-");
  const archPart = parts.at(-1); // x64, arm64, arm, armv7, etc.
  const os = parts.at(-2); // darwin, linux, windows

  if (!os || !archPart) {
    throw new Error(`Cannot derive platform from archive filename: ${filename}`);
  }

  // "win32" is the legacy Node.js-style token; current builds emit GOOS "windows".
  const goos = os === "win32" ? "windows" : os;

  // 处理 armvN 格式
  const armMatch = archPart.match(/^armv(\d+)$/);
  if (armMatch) {
    const goarm = armMatch[1];
    return {goos, goarch: "arm", goarm, suffix: archPart};
  }

  // 处理 arm_hf (兼容旧格式)
  if (archPart === "arm_hf") {
    return {goos, goarch: "arm", goarm: "7", suffix: "armv7"};
  }

  // 标准架构名
  let goarch: string;
  if (archPart === "x64") {
    goarch = "amd64";
  } else {
    goarch = archPart; // arm, arm64, mips, etc.
  }

  return {goos, goarch};
}

export function platformKey(platform: Platform): string {
  // 如果有自定义 suffix，使用它
  if (platform.suffix) {
    return `${platform.goos}-${platform.suffix}`;
  }

  // Node.js-style arch (amd64 -> x64), GOOS-style os (windows/linux/darwin)
  const arch = platform.goarch === "amd64" ? "x64" : platform.goarch;
  return `${platform.goos}-${arch}`;
}

export function parsePlatforms(value?: string): Platform[] | undefined {
  if (!value) {
    return undefined;
  }

  return value.split(",").map((platform) => {
    const parts = platform.trim().split("/");
    const goos = parts[0];
    const goarch = parts[1];
    const extra = parts[2]; // 可能是 "v7", "7", 或自定义 suffix

    if (!goos || !goarch) {
      throw new Error(`Invalid platform "${platform}". Expected format: linux/amd64 or linux/arm/v7`);
    }

    // 如果 extra 存在
    if (extra) {
      // 如果是 "v数字" 格式 (例如: v7)，提取数字作为 GOARM
      const vMatch = extra.match(/^v(\d+)$/);
      if (vMatch) {
        const goarm = vMatch[1];
        const suffix = goarch === "arm" ? `armv${goarm}` : extra;
        return {goos, goarch, goarm, suffix};
      }

      // 如果是纯数字，当作 GOARM
      if (/^\d+$/.test(extra)) {
        const suffix = goarch === "arm" ? `armv${extra}` : extra;
        return {goos, goarch, goarm: extra, suffix};
      }

      // 其他情况当作 suffix
      return {goos, goarch, suffix: extra};
    }

    return {goos, goarch};
  });
}

export function getArchiveFiles(dir: string) {
  return readdirSync(dir)
    .filter((file) => file.endsWith(".tar.gz"))
    .sort();
}

export function getArtifactFiles(dir: string) {
  return readdirSync(dir)
    .filter((file) => file !== ".DS_Store")
    .sort();
}

export function computeSha256(filePath: string) {
  return createHash("sha256").update(readFileSync(filePath)).digest("hex");
}

export function formatFileSize(bytes: number) {
  const units = ["B", "K", "M", "G"];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }

  return `${size >= 10 || unitIndex === 0 ? size.toFixed(0) : size.toFixed(1)}${units[unitIndex]}`;
}

export function formatBuildTime(date: Date) {
  const pad = (value: number) => String(value).padStart(2, "0");
  return [
    date.getUTCFullYear(),
    "-",
    pad(date.getUTCMonth() + 1),
    "-",
    pad(date.getUTCDate()),
    "_",
    pad(date.getUTCHours()),
    ":",
    pad(date.getUTCMinutes()),
    ":",
    pad(date.getUTCSeconds()),
  ].join("");
}

export function yamlQuote(value: string) {
  return `"${value.replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
}

export function normalizeVersion(gitDescribe: string): string {
  // v1.0.1-1-g730c7e4 → v1.0.1-dev+sha-730c7e4
  // v1.0.1 → v1.0.1
  // v1.0.1-dirty → v1.0.1-dirty
  // 730c7e4 → v0.0.0-dev+sha-730c7e4

  const match = gitDescribe.match(/^(v[\d.]+)-(\d+)-(g[a-f0-9]+)(-dirty)?$/);
  if (match) {
    const [, tag, , hash, dirty] = match;
    const cleanHash = hash!.replace(/^g/, "");
    return `${tag}-dev+sha-${cleanHash}${dirty || ""}`;
  }

  // If it's just a commit hash (no tag), use v0.0.0-dev+sha-{hash}
  if (/^[a-f0-9]+(-dirty)?$/.test(gitDescribe)) {
    return `v0.0.0-dev+sha-${gitDescribe}`;
  }

  // Otherwise return as-is (stable tag or dirty tag)
  return gitDescribe;
}
