# Add CNB build to sixfrp (frppc-only), ported from haokun-panel

## Context

`sixfrp` (repo `cnb.cool/sixkun/sixfrp`) is a slimmed, **frppc-only** slice of
`haokun-panel` (Go module name is still `haokun-panel`). It ships only the frpp
client: `cmd/frppc` + shared `haokun/…` code + `Dockerfile.frppc`. It is already
bun/TS-ready (`bun.lock`, `tsconfig.json`) but has **no `scripts/`** and **no
`.cnb/` directory**, even though `.cnb.yml` already `include`s
`.cnb/tag-push.yml` and `.cnb/main-push.yml`.

Goal: give sixfrp its own CI, **full-mirroring haokun-panel's frppc pipeline**
(nightly + tagged release + Docker image), by **porting haokun-panel's
`scripts/build/` TS infra trimmed to frppc-only**. Master/frpps/ztna/haokun-email
concerns are dropped — they don't exist in sixfrp.

Decisions confirmed with user: **Full mirror** scope; **Port haokun-panel scripts** approach.

## Source of truth (reference files under `/Users/dog/Gradii/haokun-panel/`)

- `scripts/build/{types,console,command,utils,config,builder,run,generate-version}.ts`
- `scripts/build-target.ts`, `scripts/merge-summaries.ts`
- `.cnb/main-push.yml`, `.cnb/tag-push.yml`
- `Dockerfile.frppc` (already identical in sixfrp)

## Files to create

### 1. `scripts/build/` — ported infra, frppc-only

| File | Port notes |
|---|---|
| `types.ts` | Copy; narrow `BuildTarget` to `"frppc"`. Keep `Platform`, `BuildSpec`, `ArchiveInfo`, `BuildConfig`. |
| `console.ts` | Copy verbatim. |
| `command.ts` | Copy verbatim (`commandOutput`, `run`, `runOrThrow`). |
| `utils.ts` | Copy verbatim — target-agnostic (`outputFilename`, `platformFromArchive`, `platformKey`, `parsePlatforms`, `getArchiveFiles`, `getArtifactFiles`, `computeSha256`, `formatFileSize`, `formatBuildTime`, `yamlQuote`, `normalizeVersion`). Handles `amd64→x64`, `arm`/`armv7` suffixes. |
| `config.ts` | Keep `createConfig`; trim `createSpecs` to return **only the frppc spec** with **no `beforeBuild`/`afterBuild` hooks** (drops ztna + interactive GitHub release prompt). Keep the frppc 16-platform default matrix. Default `releaseBaseUrl` → sixfrp cnb (secondary; see below). |
| `builder.ts` | Copy the `build()` loop and `computeChecksums()` (writes `build/frppc/SHA256SUMS`). **Drop** the per-target `sumary.yml` write + `renderSummary` call (buggy in reference — `platformKey` called with wrong args — and unused because `merge-summaries` produces the authoritative summary). Consequently **do not port `summary.ts`**. ldflags stay `-s -w -X haokun-panel/haokun/version.Version=… -X …BuildTime=… -X …ReleaseArch=…` (module name matches). |
| `run.ts` | Copy `runTargets`, drop the `buildZtnaFrontend`/`promptRelease` hook wiring. |
| `generate-version.ts` | Copy verbatim (used by nightly `main-push`). |

### 2. `scripts/build-target.ts`

Entry point. Keep invocation shape `bun run scripts/build-target.ts frppc`, but
validate only `frppc` (usage error otherwise), then `runTargets(import.meta.dir, ["frppc"])`.

### 3. `scripts/merge-summaries.ts`

Port; set `const targets = ["frppc"]` and default-repoSlug fallback to
`sixkun/sixfrp`. **Already CNB-aware**: emits
`https://cnb.cool/${CNB_REPO_SLUG}/-/releases/download/${tag}` and derives platform
keys (`frppc-linux-armv7`, …) from filenames, so `arm`/`armv7` don't collide.
Writes `build/summary.yml`.

### 4. `.cnb/main-push.yml` — nightly (frppc-only)

Adapt reference `build-nightly`. `runner.cpus: 4`, image `golang:1.26.1-bookworm`.
- `imports:` only `https://cnb.cool/gradii/secret/-/blob/main/gradii/go-fetch-private.yml` (provides `CNB_PRIVATE_REPO_TOKEN`). **Drop** the sixkun `release-envs.yml` import, `SIXKUN_PROJECT`, `SKIP_ZTNA_FRONTEND`.
- `env:` `GOEXPERIMENT: jsonv2`, `GOPRIVATE: cnb.cool/gradii/goravel,cnb.cool/gradii/frp,cnb.cool/sixkun`.
- Stages: `install-dependencies` (git/curl/jq/nodejs/npm + `npm i -g bun`) → `setup-go-private-auth` (write `~/.netrc`) → `fetch-tags` (fetch `v*`, drop local `nightly`) → `generate-dev-version` (`bun run scripts/build/generate-version.ts`, export `devVersion`) → `build-frppc` (`VERSION=$devVersion`, `PLATFORMS=<full frppc list>`, `bun run scripts/build-target.ts frppc`) → `merge-summaries` → `create-nightly-release` (`type: git:release`, tag `nightly`, `preRelease: true`, `latest: false`) → `upload-artifacts-to-release` (`cnbcool/attachments`, `build/frppc/*.tar.gz` + `build/summary.yml`) → `cleanup` (rm netrc).
- **Drop** all ztna/haokun-email download + `cnb:resolve` signal stages.

### 5. `.cnb/tag-push.yml` — tagged release + Docker (frppc-only)

Trigger `"v*.*.*": tag_push:`. Two pipelines mirroring the reference:

**a) `release-pipeline`** (image `golang:1.26.1-bookworm`, same imports/env as above):
`echo build info` → `install-dependencies` → `setup-go-private-auth` →
`build-frppc` (`VERSION=${CNB_BRANCH}`, full `PLATFORMS`) → `merge-summaries` →
`resolve-image-version` (`VERSION_NO_V=${CNB_BRANCH#v}`, export) →
`create-release` (`type: git:release`, `latest: true`, `preRelease: false`) →
`cleanup` → `upload-artifacts-to-release` (`build/frppc/*.tar.gz` + `build/summary.yml`) →
`notify-frppc-docker-pipeline` (`type: cnb:resolve`, key `release-ready`, data `version: $VERSION_NO_V`).

**b) `frppc-docker-pipeline`** (needs docker/buildx):
- `services: - name: docker` (rootless buildkitd), `imports:` `go-fetch-private.yml`.
- `env:` `FRPPC_IMAGE: ${CNB_DOCKER_REGISTRY}/${CNB_REPO_SLUG_LOWERCASE}/frppc`, `GOEXPERIMENT`, `GOPRIVATE` as above.
- Stages: `resolve-image-version` → `await-release` (`type: cnb:await`, key `release-ready`, export `version→IMAGE_VERSION`) → `install-dependencies` (git/curl/jq) → `build-and-push-frppc-docker`.
- `build-and-push-frppc-docker`: `docker login docker.cnb.cool` with `$CNB_TOKEN`/`$CNB_TOKEN_USER_NAME`; write `/tmp/frppc-netrc` from `CNB_PRIVATE_REPO_TOKEN`; `docker buildx build --platform linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64 --build-arg VERSION=v${IMAGE_VERSION} --secret id=netrc,src=/tmp/frppc-netrc -f Dockerfile.frppc --push -t ${FRPPC_IMAGE}:${IMAGE_VERSION} -t ${FRPPC_IMAGE}:latest .`; rm netrc.
- **Drop** the reference's cross-registry `skopeo copy` into `sixfrp` (`SIXFRP_IMAGE`, `sixfrp-registry-release.yml`) — sixfrp now *is* the target repo, so the image is pushed straight to its own registry.

Full frppc `PLATFORMS` string (both nightly & tag):
`darwin/amd64,darwin/arm64,freebsd/amd64,linux/amd64,linux/arm,linux/arm/v7,linux/arm64,linux/loong64,linux/mips,linux/mips64,linux/mips64le,linux/mipsle,linux/riscv64,openbsd/amd64,windows/amd64,windows/arm64`

## Not doing
- No `.cnb.yml` change — it already includes exactly these two files.
- Not porting `summary.ts`, `release.ts`, `release-frppc.ts`, `ztna.ts`, `build-frpps.ts`, `build-master.ts`, `build-all.ts` (frpps/master/ztna/interactive-GitHub-release — irrelevant here).
- Not adding security-scan / web-nilaway / custom-domain-checks (not in sixfrp's `.cnb.yml`).

## Points to verify after implementation
- `bun run scripts/build-target.ts frppc` builds all 16 platforms locally and writes `build/frppc/*.tar.gz` + `build/frppc/SHA256SUMS`; then `bun run scripts/merge-summaries.ts` writes a well-formed `build/summary.yml` (spot-check `frppc-linux-arm` vs `frppc-linux-armv7` are distinct keys). Requires netrc/GOPRIVATE access to `cnb.cool/gradii/*` + `cnb.cool/sixkun/*` private modules — a bare `go build ./cmd/frppc` for the host platform is the minimum offline smoke test.
- Confirm the `.env`-referenced secret files exist for `sixkun/sixfrp` on cnb (`gradii/go-fetch-private.yml` → `CNB_PRIVATE_REPO_TOKEN`); the nightly/tag pipelines fail at `setup-go-private-auth` otherwise.
- YAML is only exercised on cnb.cool push events; validate structurally (indentation / stage keys) before pushing a real tag.
