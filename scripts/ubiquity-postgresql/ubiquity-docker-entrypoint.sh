#!/usr/bin/env bash

#
# This script is called upon first run of a container from this image.
# It will:
# - create SSL certificate and key and configure postgresql to use them
# - use existing certificate and key if these are found (from secrets or mounted volume)
# - configure postgresql to enforce the use of SSL but not to verify the client certificate
#
set -e

if [ "$1" = 'postgres' ] && [ "$(id -u)" = '0' ]; then
    echo "Creating SSL directory $PGSSL and setting ownership to user postgres ..."
    mkdir -p $PGSSL
    chown postgres $PGSSL
    chmod 700 $PGSSL
fi

exec docker-entrypoint.sh "$@"

