#!/bin/bash

if [ -d /root/.config/personalizations.d ]
then
    source /root/.config/personalizations.d/*.sh
fi
