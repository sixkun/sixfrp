export type BuildTarget = "frppc";

export type Platform = {
  goos: string;
  goarch: string;
  goarm?: string; // 用于 ARM 变体 (例如: "7" 对应 arm_hf)
  suffix?: string; // 用于文件名后缀 (例如: "arm_hf", 而不是 "armv7")
};

export type BuildSpec = {
  target: BuildTarget;
  title: string;
  sourceDir: string;
  buildDir: string;
  outputPrefix: string;
  beforeBuild?: () => void;
  afterBuild?: () => Promise<void>;
};

export type ArchiveInfo = {
  filename: string;
  checksum: string;
  goos: string;
  goarch: string;
};

export type BuildConfig = {
  scriptDir: string;
  projectDir: string;
  version: string;
  buildTime: string;
  pubDate: string;
  summaryVersion: string;
  releaseBaseUrl: string;
  requestedPlatforms: Platform[];
};
