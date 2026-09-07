#!/usr/bin/env bun

import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

interface Platform {
  sha256: string;
  url: string;
}

interface ArchiveEntry {
  target: string;
  filename: string;
  checksum: string;
  goos: string;
  goarch: string;
}

function parseFilename(filename: string): { target: string; goos: string; goarch: string } | null {
  // 匹配格式: frppc-windows-amd64.tar.gz, frppc-linux-armv7.tar.gz
  const match = filename.match(/^(frppc)-([^-]+)-([^.]+)\.tar\.gz$/);
  if (!match) return null;

  return {
    target: match[1]!,
    goos: match[2]!,
    goarch: match[3]!,
  };
}

function readSHA256SUMS(buildDir: string, target: string): ArchiveEntry[] {
  const sha256Path = join(buildDir, target, "SHA256SUMS");

  if (!existsSync(sha256Path)) {
    console.warn(`Warning: ${sha256Path} not found, skipping ${target}`);
    return [];
  }

  const content = readFileSync(sha256Path, "utf-8");
  const entries: ArchiveEntry[] = [];

  for (const line of content.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;

    // 格式: checksum  filename
    const parts = trimmed.split(/\s+/);
    if (parts.length !== 2) continue;

    const [checksum, filename] = parts;
    const parsed = parseFilename(filename!);

    if (parsed) {
      entries.push({
        target: parsed.target,
        filename: filename!,
        checksum: checksum!,
        goos: parsed.goos,
        goarch: parsed.goarch,
      });
    }
  }

  return entries;
}

function mergeSummaries(projectDir: string, outputPath: string) {
  const targets = ["frppc"];
  const buildDir = join(projectDir, "build");

  const allEntries: ArchiveEntry[] = [];

  // 读取每个 target 的 SHA256SUMS
  for (const target of targets) {
    const entries = readSHA256SUMS(buildDir, target);
    allEntries.push(...entries);
  }

  if (allEntries.length === 0) {
    console.error("Error: No build artifacts found");
    process.exit(1);
  }

  // 确定版本号和发布日期
  const version = process.env.VERSION || "dev";
  const pubDate = new Date().toISOString();

  // 构建平台信息
  const isCNB = process.env.CNB_COMMIT !== undefined;

  let baseUrl: string;
  if (isCNB) {
    const repoSlug = process.env.CNB_REPO_SLUG || "sixkun/sixfrp";
    // CNB_BRANCH 在 tag push 时是 tag 名称（如 v1.0.0），在普通 push 时是分支名（如 main）
    const ref = process.env.CNB_BRANCH || "nightly";
    const isTag = ref.startsWith("v");
    const tag = isTag ? ref : "nightly";
    baseUrl = `https://cnb.cool/${repoSlug}/-/releases/download/${tag}`;
  } else {
    baseUrl = `https://cnb.cool/sixkun/sixfrp/-/releases/download/${version}`;
  }

  const platforms: Record<string, Platform> = {};

  for (const entry of allEntries) {
    const key = `${entry.target}-${entry.goos}-${entry.goarch}`;
    platforms[key] = {
      sha256: entry.checksum,
      url: `${baseUrl}/${entry.filename}`,
    };
  }

  // 生成 YAML
  const lines = [
    `version: "${version.replace(/^v/, "")}"`,
    `pub_date: "${pubDate}"`,
    "platforms:",
  ];

  for (const [key, platform] of Object.entries(platforms)) {
    lines.push(`  ${key}:`);
    lines.push(`    sha256: "${platform.sha256}"`);
    lines.push(`    url: "${platform.url}"`);
  }

  const output = lines.join("\n") + "\n";
  writeFileSync(outputPath, output, "utf-8");

  console.log(`✓ Merged summary written to: ${outputPath}`);
  console.log(`  Version: ${version}`);
  console.log(`  Platforms: ${Object.keys(platforms).length}`);
  console.log(`  Base URL: ${baseUrl}`);

  if (isCNB) {
    const ref = process.env.CNB_BRANCH || "nightly";
    const isTag = ref.startsWith("v");
    console.log(`  CNB mode: ${isTag ? `tag (${ref})` : "nightly"}`);
  }
}

// 运行
const projectDir = join(import.meta.dir, "..");
const outputPath = join(projectDir, "build", "summary.yml");

mergeSummaries(projectDir, outputPath);
