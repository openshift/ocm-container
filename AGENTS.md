# AGENTS.md

This file provides guidance to AI agents working with code in this repository.

## What This Repo Does

`ocm-container` is a Go CLI that builds and runs ephemeral, containerized SRE environments for accessing OpenShift v4 clusters. It launches per-cluster containers (via podman or docker) with credentials, cluster context, and tooling, then destroys the container and credentials on exit. Three image variants are built: **micro** (ocm + backplane + oc), **minimal** (micro + backplane-tools), and **full** (minimal + additional SRE tooling and environment configuration).

## Repository Layout

| Area | Location | Description |
|---|---|---|
| **CLI entrypoint** | `main.go`, `cmd/` | Cobra command setup (`cmd/root.go`), flag definitions (`cmd/flags.go`), version subcommand (`cmd/version/`). |
| **Feature plugins** | `pkg/features/` | 13 self-registering feature modules (jira, pagerduty, backplane, gcloud, osdctl, ops-utils, persistent-histories, personalization, legacy-aws-credentials, certificate-authorities, image-cache, additional-cluster-envs, ports). Each implements `Configure`/`Initialize`/`Enabled`/`HandleError`/`ExitOnError` and returns an `OptionSet` (volume mounts, env vars, port mappings, post-start hooks). |
| **Feature registrar** | `pkg/features/registrar/registrar.go` | Imports all features and registers their `--no-<feature>` flags; adding a feature requires an import + entry in `featureFlags` here. |
| **Container runtime** | `pkg/ocmcontainer/ocmcontainer.go` | `Runtime` orchestrator: merges flags + config + feature OptionSets into `engine.ContainerRef`, creates container, runs post-start hooks, attaches or execs. |
| **Engine abstraction** | `pkg/engine/engine.go` | Podman/docker shell-out wrapper (`SupportedEngines`, `SupportedPullImagePolicies`). |
| **OCM client** | `pkg/ocm/ocm.go` | OCM authentication, URL aliasing (`prod`/`stage`/`int`/`prodgov`), cluster lookup. |
| **Utilities** | `pkg/log/`, `pkg/subprocess/`, `pkg/deprecation/`, `pkg/utils/` | Logging, subprocess execution, deprecation messages, version info. **Important:** `pkg/utils.IsRunningInOcmContainer()` is a public API consumed by `osdctl` — do not change its contract. |
| **Container image** | `Containerfile` | Multi-stage build: `ocm-container-micro` → `ocm-container-minimal` → `ocm-container` (full). Builder stages for `backplane-tools`, `omc`, `jira-cli`. |
| **Container assets** | `utils/bin/`, `utils/bashrc.d/`, `utils/dockerfile_assets/` | Scripts, bash configuration, and assets baked into the container images. |
| **CI/CD** | `.tekton/`, `.ci/` | Konflux PipelineRuns for PR and push builds (micro, minimal, full variants); validation scripts. |
| **Boilerplate** | `boilerplate/` | Auto-generated files from `openshift/boilerplate` (subscribed convention: `openshift/golang-lint`). **Never edit by hand** — use `make boilerplate-update`. |
| **Docs** | `docs/` | Feature-specific documentation, migration guide, example config. |

## Build / Test / Lint

```bash
# IMPORTANT: `make build` builds the CONTAINER IMAGE, not the Go binary.
# Use `make build-binary` to compile the CLI executable.

# Container image builds (local use)
make                     # default: build-full-local + tag-full-local (produces ocm-container:latest)
make build               # same as above
make build-micro         # micro image (ocm + backplane + oc)
make build-minimal       # minimal image (micro + backplane-tools)
make build-full          # full image (minimal + full toolset)
make build-full-local    # full image without manifest (for local testing)

# Go binary builds
make build-binary        # compile to build/ocm-container
make build-snapshot      # goreleaser snapshot build
make go-build            # full local gate: mod + fmt + lint + test + build-snapshot

# Testing
make test                # run all Ginkgo v2 test suites (go test ./... -v)
make mod                 # go mod tidy
make fmt                 # gofmt -s -l -w cmd pkg utils
make lint                # golangci-lint via boilerplate (config: boilerplate/openshift/golang-lint/golangci.yml)

# CI checks
make check               # check environment configuration
make check-env           # display resolved config (container engine, Go, image registry, etc.)
make check-github-quota  # check GitHub API quota (requires GITHUB_TOKEN)
make pr-check            # validate-tekton + check-image-build (the PR gate)
make validate-tekton     # run .ci/validate-tekton-pipelines.sh (critical guardrails — see below)

# Boilerplate
make boilerplate-update         # pull latest boilerplate templates
make boilerplate-freeze-check   # verify boilerplate files haven't been hand-edited

# Release
make release-binary      # goreleaser release (requires GITHUB_TOKEN)
```

## Running a Single Test

```bash
# Run one test by name (regex match)
TESTOPTS="-run TestPorts" make test

# Or directly with go test
go test ./pkg/features/ports -v

# Ginkgo focus (Ginkgo v2 syntax)
go test ./pkg/features/ports -v --ginkgo.focus="should register port mappings"
```

## Architecture

**Startup flow:**
1. `main.go` → `cmd.Execute()` → `cmd/root.go` `rootCmd.RunE`
2. `ocmcontainer.New(cmd, args)` — parses flags/config, checks cluster existence
3. `features.Initialize()` — calls `Configure()` + `Initialize()` on all registered features, merges their `OptionSet` results
4. `o.CreateContainer(c)` — builds `engine.ContainerRef` from flags + config + feature options, runs `engine.Create()`
5. `o.Start()` → `o.ExecPostRunBlockingCmds()` — runs post-start hooks (e.g., cluster login)
6. `o.Run()` — either `Attach()` (interactive shell) or `ExecLive(command)` (one-shot command)

**Feature plugin contract** (`pkg/features/feature.go`):
- Each feature is a package in `pkg/features/<name>/` implementing the `Feature` interface:
  - `Configure() error` — reads config from viper (`features.<name>.*`), validates
  - `Initialize() (OptionSet, error)` — returns volume mounts, env vars, port maps, post-start hooks
  - `Enabled() bool` — checks `features.<name>.enabled` config and `--no-<name>` flag
  - `HandleError(error)` — custom error logging (e.g., different log level if user provided config)
  - `ExitOnError() bool` — whether initialization failure should fail the entire launch
- Features self-register in their `init()` via `features.Register("myFeature", &f)`
- Activation requires two steps: (1) implement the interface in a new `pkg/features/<name>/` package; (2) import it in `pkg/features/registrar/registrar.go` and add an entry to `featureFlags` with `FeatureFlagName`/`FlagHelpMessage` constants

See [docs/new_feature.md](docs/new_feature.md) for the full scaffolding template.

## Adding a Feature

Two required steps:

1. **Create the feature package** in `pkg/features/<name>/` implementing the `Feature` interface. Use the scaffolding in [docs/new_feature.md](docs/new_feature.md) as a template. Features self-register in their `init()` via `features.Register("<name>", &f)`.

2. **Register the feature flag** in `pkg/features/registrar/registrar.go`:
   - Import the new package (triggers `init()` registration)
   - Add an entry to `featureFlags` referencing the feature's `FeatureFlagName` and `FlagHelpMessage` constants

Naming conventions:
- Feature flag: `--no-<feature>` (hidden, opt-out style)
- Config key: `features.<feature>.enabled` (boolean) and `features.<feature>.<setting>` (feature-specific settings, using camelCase)
- Feature names use kebab-case in flags (`--no-my-feature`) but map to camelCase in config (`features.myFeature.enabled`)

## Container Images

Three build targets, each layering on the previous:

- **`ocm-container-micro`** — minimal footprint: `ocm`, `ocm-backplane`, `oc`
- **`ocm-container-minimal`** — micro + full `backplane-tools` suite (aws-cli, rosa, osdctl, yq, etc.)
- **`ocm-container`** (full) — minimal + additional SRE tooling: `omc`, `jira-cli`, `oc-nodepp`, `vault`, scripting utilities from `utils/bin/`, and opinionated bash environment from `utils/bashrc.d/`

Makefile targets: `make build-micro`, `make build-minimal`, `make build-full` (or `make build-full-local` for local testing without pushing to a manifest).

## Dependency Updates Are Bot-Managed

This repository uses **Konflux MintMaker** for automated dependency updates:

- **Go modules:** Daily `chore(deps):`/`fix(deps): update gomod dependencies` PRs via `.github/renovate.json` (extends `github>openshift/boilerplate//.github/renovate.json`)
- **Konflux references:** Automated `chore(deps): update konflux references` PRs

**Do not hand-bump `go.mod`/`go.sum`** unless a change genuinely requires a new dependency or version — the bot handles routine updates.

**Security scanning** runs automatically in Tekton on every PR and push:
- Clair (container vulnerability scan)
- Snyk (dependency vulnerabilities)
- ClamAV (malware scan)
- Coverity (static analysis)
- ShellCheck (shell script linting)
- Unicode check (detects bidirectional unicode attacks)
- RPM signature scan

See `.tekton/pipeline-pull-request-build-image.yaml` for the full task list.

## Boilerplate

`boilerplate/` is auto-generated by the [openshift/boilerplate](https://github.com/openshift/boilerplate) framework. This repo subscribes to the `openshift/golang-lint` convention (see `boilerplate/update.cfg`).

**Rules:**
- Files marked `GENERATED BY BOILERPLATE. DO NOT EDIT.` will be overwritten on `make boilerplate-update`
- Never edit boilerplate files by hand — changes should go upstream to `openshift/boilerplate` first
- `.gitattributes` has a freeze-check block that **must remain the last thing in the file** — the `make boilerplate-freeze-check` gate enforces this
- A nightly CronJob (`.tekton/agentic-boilerplate-cronjob.yaml`) already runs Claude Code at 21:36 to execute `make boilerplate-update boilerplate-commit` and open PRs for upstream changes — agents should not duplicate this automation

## Tekton Guardrails

**Critical:** `.ci/validate-tekton-pipelines.sh` enforces two invariants that have real failure history:

1. **Never embed `pipelineSpec:`** — all PipelineRuns must use `pipelineRef: {name: pull-request-build-image}`, not an embedded spec
2. **Always include `linux/arm64` in `build-platforms`** — all builds must target both `linux/x86_64` and `linux/arm64`

**Why:** On 2026-03-12, the Konflux bot (`red-hat-konflux-kflux-prd-rh03`) auto-merged PR #488 in 2 seconds with no human review, replacing `ocm-container-micro`'s `pipelineRef` with a 626-line embedded `pipelineSpec` and silently dropping `linux/arm64` from the build platforms. This broke arm64 support for a multi-arch CLI used by M1 Mac users.

**Enforcement:** `make validate-tekton` (also run as part of `make pr-check`) fails the build if either invariant is violated.

**Reference:** ROSAENG-3945 (full incident investigation).

## Working Rules & Conventions

- **Target branch:** `master` (default branch)
- **Commit messages:**
  - Human commits: short imperative one-liners (`Add --pull flag and consolidate rh-aws-saml-login layer`)
  - Bot commits: `chore(deps):` or `fix(deps):` prefix (Konflux MintMaker style)
- **Testing:** Add Ginkgo v2 specs alongside new functions — every `pkg/` package has a `*_suite_test.go` + test files
- **Formatting:** Run `make fmt` before committing (enforced by golangci-lint in CI)
- **Code review:** Reviewers and approvers are listed in `OWNERS` (AlexSmithGH, chamalabey, clcollins, smarthall, tnierman, tkong-redhat, samanthajayasinghe)

## Gotchas Worth Knowing

1. **`pkg/utils.IsRunningInOcmContainer()` is a public API** — `osdctl` imports it to detect when running inside ocm-container and adjust behavior (e.g., skip browser launch). Do not change its contract or semantics.

2. **Argument splitting happens before Cobra** — `cmd/root.go` `splitArgs` handles `--` separation to extract commands to run inside the container, then resets args for Cobra. This means argument handling is non-standard compared to typical Cobra apps.

3. **`--no-*` feature flags are hidden and inverted** — users disable features with `--no-my-feature`, but the code checks `featureEnabled("my-feature")` which returns `!viper.GetBool("no-my-feature")`. See `pkg/ocmcontainer/ocmcontainer.go` `featureEnabled`/`lookUpNegativeName`.

4. **Container builds need `GITHUB_TOKEN`** — to avoid GitHub API rate limits when downloading binaries during image builds. Set `GITHUB_TOKEN` in your environment or run `make check-github-quota` to verify quota before building. This is automatically available in Konflux but may be missing in local builds.

5. **`.gitignore` ignores `.github/*`, but `.github/renovate.json` is tracked** — the file was force-added (`git add -f`) to override `.gitignore`. New files in `.github/` require the same treatment or won't be committed.

6. **`make build` builds the image, not the binary** — this is the #1 source of confusion. `make build` → container image tagged `ocm-container:latest`. `make build-binary` → `build/ocm-container` Go executable.

7. **Config file location** — `~/.config/ocm-container/ocm-container.yaml` (viper searches `$HOME/ocm-container/` for `ocm-container.yaml`). Can be overridden with `--config` flag.

8. **`imagePullPolicy` vs. `--pull` flag** — the flag `--pull` maps to config key `imagePullPolicy` (not `pull`) via `flagConfigOverrides` in `cmd/flags.go`. Supported values: `always`, `missing`, `never`, `newer`.

## Safety

This container image is used by Red Hat SREs to access production customer clusters. Security considerations:

- **Ephemeral credentials:** Containers are launched with `--rm`, destroying credentials on exit. Do not weaken this.
- **Cluster isolation:** Each cluster gets its own container with isolated `.kube` config and credentials.
- **Sensitive files are gitignored:** `.gitignore` blocks `*crt`, `*cert`, `*key`, `*csr`, `*pem` — never force-add secrets to the repo.
- **Tooling changes require review:** Adding new tools to the full image expands the attack surface — discuss with maintainers before adding binaries or packages.
- **Privileged containers:** The container runs with `--privileged` (needed for podman-in-podman workflows) — be careful with volume mounts and environment variable pass-through.

**When in doubt:** ask for review from a maintainer (see `OWNERS`).
