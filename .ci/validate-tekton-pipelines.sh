#!/bin/bash
#
# Validates .tekton/ PipelineRun files to prevent regressions from
# Konflux bot auto-updates that silently alter build configuration.
#
# Override in Prow with: /override validate-tekton
#
set -euo pipefail

TEKTON_DIR="$(cd "$(dirname "$0")/../.tekton" && pwd)"
ERRORS=0

PIPELINERUN_FILES=(
  "ocm-container-micro-push.yaml"
  "ocm-container-micro-pull-request.yaml"
  "ocm-container-minimal-push.yaml"
  "ocm-container-minimal-pull-request.yaml"
  "ocm-container-push.yaml"
  "ocm-container-pull-request.yaml"
)

check_no_embedded_pipelinespec() {
  local file="$1"
  local filepath="${TEKTON_DIR}/${file}"

  if [ ! -f "${filepath}" ]; then
    echo "WARNING: ${file} not found, skipping"
    return
  fi

  if grep -q "pipelineSpec:" "${filepath}"; then
    echo ""
    echo "================================================================================"
    echo "FAILURE: ${file} contains an embedded pipelineSpec"
    echo "================================================================================"
    echo ""
    echo "  All PipelineRun files must use 'pipelineRef' (not an embedded 'pipelineSpec')."
    echo ""
    echo "  WHY: On 2026-03-12, the Konflux bot (red-hat-konflux-kflux-prd-rh03) silently"
    echo "  replaced ocm-container-micro's pipelineRef with a 626-line embedded pipelineSpec"
    echo "  via PR #488 (commit a8458a1), stripping arm64 support in the process. The PR was"
    echo "  auto-merged in 2 seconds with no human review."
    echo ""
    echo "  FIX: Revert the file to use pipelineRef to the shared pipeline:"
    echo "    pipelineRef:"
    echo "      name: pull-request-build-image"
    echo ""
    echo "  See ROSAENG-3945 for full investigation details."
    echo "================================================================================"
    echo ""
    ERRORS=$((ERRORS + 1))
  fi
}

check_arm64_in_build_platforms() {
  local file="$1"
  local filepath="${TEKTON_DIR}/${file}"

  if [ ! -f "${filepath}" ]; then
    echo "WARNING: ${file} not found, skipping"
    return
  fi

  if ! grep -q "linux/arm64" "${filepath}"; then
    echo ""
    echo "================================================================================"
    echo "FAILURE: ${file} is missing linux/arm64 in build-platforms"
    echo "================================================================================"
    echo ""
    echo "  All PipelineRun files must include 'linux/arm64' in their build-platforms list."
    echo ""
    echo "  WHY: On 2026-03-12, the Konflux bot silently removed arm64 from"
    echo "  ocm-container-micro's build-platforms via PR #488 (commit a8458a1). The bot's"
    echo "  default template only generates linux/x86_64 and does not preserve custom"
    echo "  build-platforms values."
    echo ""
    echo "  FIX: Add linux/arm64 to the build-platforms parameter:"
    echo "    - name: build-platforms"
    echo "      value:"
    echo "      - linux/x86_64"
    echo "      - linux/arm64"
    echo ""
    echo "  See ROSAENG-3945 for full investigation details."
    echo "================================================================================"
    echo ""
    ERRORS=$((ERRORS + 1))
  fi
}

echo "Validating .tekton/ PipelineRun files..."
echo ""

for file in "${PIPELINERUN_FILES[@]}"; do
  check_no_embedded_pipelinespec "${file}"
  check_arm64_in_build_platforms "${file}"
done

if [ "${ERRORS}" -gt 0 ]; then
  echo "FAILED: ${ERRORS} validation error(s) found."
  echo "Override in Prow with: /override validate-tekton"
  exit 1
fi

echo "All .tekton/ PipelineRun validations passed."
