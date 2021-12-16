#!/usr/bin/env bash

set -eEuo pipefail

complete -C '/usr/local/aws-cli/v2/current/dist/aws_completer' aws

set +eEuo pipefail
