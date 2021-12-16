#!/usr/bin/env bash

set -eEuo pipefail

if ! [[ -f ~/.config/gcloud/credentials.db ]]  && [[ -f ~/.config/gcloud/credentials_readonly.db ]]; then
	cp ~/.config/gcloud/credentials_readonly.db ~/.config/gcloud/credentials.db
fi
if ! [[ -f ~/.config/gcloud/configurations/config_default ]]  && [[ -f ~/.config/gcloud/configurations/config_default_readonly ]]; then
	cp ~/.config/gcloud/configurations/config_default_readonly ~/.config/gcloud/configurations/config_default
fi
if ! [[ -f ~/.config/gcloud/active_config ]]  && [[ -f ~/.config/gcloud/active_config_readonly ]]; then
	cp ~/.config/gcloud/active_config_readonly ~/.config/gcloud/active_config
fi
if ! [[ -f ~/.config/gcloud/access_tokens.db ]]  && [[ -f ~/.config/gcloud/access_tokens_readonly.db ]]; then
	cp ~/.config/gcloud/access_tokens_readonly.db ~/.config/gcloud/access_tokens.db
fi

set +eEuo pipefail
