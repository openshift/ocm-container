# Hardened image (Hummingbird) variant — ROSAENG-75341

`Containerfile.hardened` is a copy of `Containerfile` with the base image swapped to the Red Hat
Hardened Image `registry.access.redhat.com/hi/core-runtime:2.43-builder` (digest-pinned). It is
built alongside the production `Containerfile`, which is untouched until cutover.

- Build locally: `make build-hardened-local`
- Konflux build: `.tekton/ocm-container-hardened-push.yaml` (build-only pilot, pushes
  `<sha>-hardened`, modelled on openshift/route-monitor-operator#643)
- The base digest is bumped by Renovate (`.github/renovate.json`), so a base-image update
  triggers a build.

## Differences from `Containerfile`
- Base image: `hi/core-runtime:2.43-builder` (a Hummingbird variant that includes `dnf`) instead of `ubi10/ubi`.
- `USER 0` and `HOME=/root` are set, because the base runs as a non-root user.
- `python3.14-pip` is `python3-pip` in the Hummingbird repos.
- `socat`, `tmux`, `wget` and a full `vim` are not in the Hummingbird repos, so the
  `extra-tools` stage builds them from pinned, checksum-verified source tarballs and copies
  them into `/usr/local`. (`tmux` is not installed in the current production image; it is
  added here because `utils/bin/ssm-portforward` needs it.) Their versions are `ARG`s at the
  top of that stage and are not covered by Renovate.
  - `socat` needs a one-line source patch (const `strchr()` results with newer glibc).
  - `tmux` is configured with a stub `yacc` because its tarball ships the generated parser.
- `/root/.bash_profile` is created so login shells load `.bashrc` (UBI provided it).
- `gzip` is installed explicitly (UBI shipped it; Hummingbird does not).
- `rpm -i --nosignature` for the AWS session-manager-plugin: Hummingbird's `rpm` rejects the
  unsigned AWS package.
- CVE scanning is **not yet in place**: this repo's `pull-request-build-image` pipeline has no `clair-scan` task, and
  the hardened build has no registered Konflux Component (so no Enterprise Contract). Trivy does not recognise the
  Hummingbird OS and reports no OS-package CVEs, so it cannot stand in. Before cutover, the hardened image must be
  scanned (Konflux Clair via a registered Component, or boilerplate's `docker-build-oci-ta` pipeline) and the results reviewed.

## Cutover checklist
1. Dedicated Konflux Component registered; builds green (amd64+arm64).
2. Enterprise Contract passes.
3. CVE scan reviewed against production.
4. Replace `Containerfile` with the hardened one; delete the parallel pipeline.
