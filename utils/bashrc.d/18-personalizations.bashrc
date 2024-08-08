# shellcheck shell=bash

if [ -d /root/.config/personalizations.d ]
then
    # shellcheck disable=SC1090
    source /root/.config/personalizations.d/*.sh
fi
