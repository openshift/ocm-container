# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OCM Container is a containerized environment for accessing OpenShift v4 clusters, built with Go and using containers. The project migrated from a bash-based system to a Go binary with a plugin-based feature system. The binary manages container lifecycle, configuration, and feature enablement through a CLI interface built with Cobra.

## Development Commands

### Build Commands
```bash
# Build the full container image (default target for local development)
make build

# Build all container variants (micro, minimal, full)
make build-all

# Build specific variants
make build-micro    # OCM, backplane, oc only
make build-minimal  # Micro + SRE backplane tools
make build-full     # Minimal + additional packages and tooling

# Cross-platform builds
make build-image-amd64
make build-image-arm64
```

### Go Development Commands
```bash
# Full Go build pipeline
make go-build

# Build binary only
make build-binary

# Run tests
make test
go test ./... -v

# Format code
make fmt

# Lint code
make lint

# Tidy modules
make mod

# Release commands
make build-snapshot      # Build snapshot without releasing
make release-binary      # Full release (requires GITHUB_TOKEN)
```

### Container Registry Operations
```bash
# Tag images for registry
make tag-all

# Push to registry (requires REGISTRY_USER/REGISTRY_TOKEN)
make push-all
make registry-login

# Push multi-arch manifests
make push-manifests
```

### Environment Check
```bash
# Display current build configuration
make check-env
```

## Code Architecture

### Core Structure
- **main.go**: Entry point that calls `cmd.Execute()`
- **cmd/**: CLI command definitions using Cobra framework
  - `root.go`: Main command and container execution logic
  - `configure.go`: Configuration management subcommands
  - `build.go`: Container build functionality
  - `flags.go`: Common flag definitions
- **pkg/**: Core business logic packages

### Key Packages
- **pkg/ocmcontainer/**: Main container orchestration logic and environment setup
- **pkg/engine/**: Container engine abstraction (Docker/Podman)
- **pkg/ocm/**: OCM (OpenShift Cluster Manager) integration
- **pkg/backplane/**: Backplane authentication and cluster access
- **pkg/featureSet/**: Modular feature system for mounting host resources:
  - `aws/`: AWS credentials and config mounting
  - `gcloud/`: Google Cloud config mounting
  - `jira/`: JIRA CLI integration
  - `pagerduty/`: PagerDuty CLI integration
  - `opsutils/`: Red Hat SRE utilities mounting
  - `osdctl/`: OSDCTL configuration mounting
  - `persistentHistories/`: Cluster terminal history persistence
  - `personalization/`: User bashrc/dotfile mounting
  - `scratch/`: Arbitrary directory mounting
  - `certificateAuthorities/`: CA trust bundle mounting
- **pkg/subprocess/**: Container process management
- **pkg/utils/**: Version and utility functions

### Configuration System
Uses Viper for configuration management with precedence:
1. CLI flags
2. Environment variables (prefixed with `OCMC_`)
3. Configuration file (`~/.config/ocm-container/ocm-container.yaml`)

### Feature System Architecture
Each feature in `pkg/featureSet/` implements a common interface for:
- Checking if feature should be enabled
- Generating container mount arguments
- Handling feature-specific environment variables

## Container Targets

The Containerfile defines three build targets:
- **micro**: Basic OCM tools (ocm, backplane, oc)
- **minimal**: Micro + SRE backplane tools
- **full**: Minimal + comprehensive tooling and environment setup

## Testing

Run tests with:
```bash
go test ./...
```

Test files are located alongside source files with `_test.go` suffix.

## Key Dependencies

- **Cobra**: CLI framework
- **Viper**: Configuration management
- **OCM SDK**: OpenShift Cluster Manager integration
- **Logrus**: Logging
- **Bubble Tea/Huh**: Interactive CLI components

## Build Configuration

- Go version: 1.24+
- Supports: Linux/Darwin on amd64/arm64
- Container engines: Podman or Docker (auto-detected)
- Registry: quay.io/app-sre/ocm-container