#!/usr/bin/env bash

if ! [[ -f ~/.config/gcloud/credentials.db ]]  && [[ -f ~/.config/gcloud/credentials_readonly.db ]]; then
	cp ~/.config/gcloud/credentials_readonly.db ~/.config/gcloud/credentials.db
fi
