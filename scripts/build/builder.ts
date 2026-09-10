import { mkdirSync, rmSync, statSync, writeFileSync } from "node:fs";
import { join } from "node:path";

import { run, runOrThrow } from "./command";
import { fail, info, log, printHeader, success } from "./console";
import type { ArchiveInfo, BuildConfig, BuildSpec } from "./types";
import {
  computeSha256,
  formatFileSize,
  getArchiveFiles,
  getArtifactFiles,
  outputFilename,
  platformFromArchive,
} from "./utils";

export async function build(config: BuildConfig, spec: BuildSpec) {
  printHeader(spec.title);
  info(`Version: ${config.version}`);
  info(`Build Time: ${config.buildTime}`);
  console.log("");

  spec.beforeBuild?.();

  log("Preparing build directory...");
  rmSync(spec.buildDir, {recursive: true, force: true});
  mkdirSync(spec.buildDir, {recursive: true});
  console.log("");

  const failedBuilds: string[] = [];
  let buildCount = 0;

  for (const platform of config.requestedPlatforms) {
    const outputName = outputFilename(spec.outputPrefix, platform);
    const outputPath = join(spec.buildDir, outputName);

    log(`Building for ${platform.goos}/${platform.goarch}${platform.goarm ? ` (ARMv${platform.goarm})` : ""}...`);

    const buildEnv: Record<string, string> = {
      CGO_ENABLED: "0",
      GOEXPERIMENT: "jsonv2",
      GOOS: platform.goos,
      GOARCH: platform.goarch,
    };

    // 添加 GOARM 环境变量（如果指定）
    if (platform.goarm) {
      buildEnv.GOARM = platform.goarm;
    }

    // ReleaseArch must match the arch token stored in the DB
    // (ParseAssetFilename derives it from the filename's arch segment) so the
    // master can match this binary to a release row. For ARM variants that is
    // the suffix (arm vs armv7), which distinguishes ARMv5/v6 from ARMv7;
    // runtime.GOARCH is just "arm" for both and would collide.
    const releaseArch = platform.suffix ?? platform.goarch;

    // The -X package path must be a package that actually exists in this
    // build's graph. It previously read haokun-panel/haokun/version, which is
    // not in cmd/frppc's graph for module cnb.cool/sixkun/sixfrp/v2, so all
    // three -X flags were silently ignored and every released binary reported
    // an empty version — which the master reads to decide upgrade eligibility.
    const versionPkg = "cnb.cool/sixkun/sixfrp/v2/haokun/version";

    const result = run(
      "go",
      [
        "build",
        "-trimpath",
        `-ldflags=-s -w -X ${versionPkg}.Version=${config.version} -X ${versionPkg}.BuildTime=${config.buildTime} -X ${versionPkg}.ReleaseArch=${releaseArch}`,
        "-o",
        outputPath,
        ".",
      ],
      spec.sourceDir,
      buildEnv,
      false,
    );

    if (result.status === 0) {
      buildCount += 1;
      success(`Built: ${outputName} (${formatFileSize(statSync(outputPath).size)})`);

      // Strip a trailing ".exe" from the archive name only: the packed Windows
      // binary keeps its .exe (so it stays runnable), but the archive is named
      // frppc-windows-x64.tar.gz rather than the noisy frppc-windows-x64.exe.tar.gz.
      const archiveName = `${outputName.replace(/\.exe$/, "")}.tar.gz`;
      info("  Creating archive...");
      runOrThrow(
        "tar",
        ["-czf", archiveName, "--exclude=._*", "--exclude=.DS_Store", outputName],
        spec.buildDir,
        {COPYFILE_DISABLE: "1"},
      );
      rmSync(outputPath, {force: true});
      success(`  Archive: ${archiveName} (${formatFileSize(statSync(join(spec.buildDir, archiveName)).size)})`);
    } else {
      fail(`Build failed for ${platform.goos}/${platform.goarch}${platform.goarm ? ` (ARMv${platform.goarm})` : ""}`);
      failedBuilds.push(`${platform.goos}/${platform.goarch}${platform.goarm ? `/ARMv${platform.goarm}` : ""}`);
    }

    console.log("");
  }

  logBuildSummary(spec, config, buildCount, failedBuilds);

  if (failedBuilds.length > 0) {
    return false;
  }

  printHeader("All builds completed successfully!");
  console.log("");

  await spec.afterBuild?.();
  return true;
}

function logBuildSummary(spec: BuildSpec, config: BuildConfig, buildCount: number, failedBuilds: string[]) {
  printHeader("Build Summary");
  success(`Successful builds: ${buildCount}/${config.requestedPlatforms.length}`);

  if (failedBuilds.length > 0) {
    fail("Failed builds:");
    for (const failed of failedBuilds) {
      fail(`  - ${failed}`);
    }
  }

  console.log("");
  info("Build artifacts:");
  for (const file of getArtifactFiles(spec.buildDir)) {
    const filePath = join(spec.buildDir, file);
    console.log(`  ${file} (${formatFileSize(statSync(filePath).size)})`);
  }

  console.log("");
  success(`Output directory: ${spec.buildDir}`);
  console.log("");

  // Per-target SHA256SUMS; scripts/merge-summaries.ts reads these to produce the
  // authoritative build/summary.yml (CNB-aware release URLs, arm vs armv7 keys).
  const archives = getArchiveFiles(spec.buildDir);
  computeChecksums(spec, archives);
}

function computeChecksums(spec: BuildSpec, archives: string[]): ArchiveInfo[] {
  log("Generating checksums...");

  const archiveInfos = archives.map((filename) => {
    const platform = platformFromArchive(filename);
    return {
      filename,
      checksum: computeSha256(join(spec.buildDir, filename)),
      goos: platform.goos,
      goarch: platform.goarch,
    };
  });

  writeFileSync(
    join(spec.buildDir, "SHA256SUMS"),
    archiveInfos.map((archive) => `${archive.checksum}  ${archive.filename}`).join("\n") + "\n",
  );
  success("Checksums saved to SHA256SUMS");
  console.log("");

  return archiveInfos;
}
