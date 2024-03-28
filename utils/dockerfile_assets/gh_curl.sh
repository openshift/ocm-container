#!/bin/bash

## GH Curl wraps CURL to a github API with an 
## authorization header if the GITHUB_TOKEN env
## var is present

GITHUB_AUTH_HEADER=""

if [[ -n $GITHUB_TOKEN ]]
then
  GITHUB_AUTH_HEADER="--header 'Authorization: Bearer $GITHUB_TOKEN'"
fi

curl $GITHUB_AUTH_HEADER $@
