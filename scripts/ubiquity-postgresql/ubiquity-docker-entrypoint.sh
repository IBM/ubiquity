#!/usr/bin/env bash

#
# This script is called upon first run of a container from this image.
# It will:
# - create SSL certificate and key and configure postgresql to use them
# - use existing certificate and key if these are found (from secrets or mounted volume)
# - configure postgresql to enforce the use of SSL but not to verify the client certificate
#
set -e

export PGSSL_PRIVATE_DIR="`dirname $UBIQUITY_DB_CERT_PRIVATE`"
export PGSSL_PUBLIC_DIR="`dirname $UBIQUITY_DB_CERT_PUBLIC`"

if [ "$1" = 'postgres' ] && [ "$(id -u)" = '0' ]; then

    echo "Creating SSL directory $PGSSL_PRIVATE_DIR and setting ownership to user postgres ..."
    mkdir -p $PGSSL_PRIVATE_DIR
    chown postgres $PGSSL_PRIVATE_DIR
    chmod 700 $PGSSL_PRIVATE_DIR

    if [ "$PGSSL_PUBLIC_DIR" != "$PGSSL_PRIVATE_DIR" ]; then
        echo "Creating SSL directory $PGSSL_PUBLIC_DIR and setting ownership to user postgres ..."
        mkdir -p $PGSSL_PUBLIC_DIR
        chown postgres $PGSSL_PUBLIC_DIR
        chmod 700 $PGSSL_PUBLIC_DIR
    fi

    if [ ! -s "$UBIQUITY_DB_CERT_PUBLIC" ]; then
        echo "Generateing SSL private and public keys..."

        if [ -z "$POSTGRES_EMAIL" ]; then
          export POSTGRES_EMAIL="user@test.com"
        fi

        # Create SSL certificates
        cd $PGSSL_PRIVATE_DIR

        # root CA
        openssl req -new -x509 -nodes -out root.crt -keyout root.key -newkey rsa:4096 -sha512 -subj /CN=TheRootCA
        chown postgres root.key
        chmod 600 root.key

        # Server certificate
        openssl req -new -out server.req -keyout $UBIQUITY_DB_CERT_PRIVATE -nodes -newkey rsa:4096 -subj "/CN=$( hostname )/emailAddress=$POSTGRES_EMAIL"
        openssl x509 -req -in server.req -CAkey root.key -CA root.crt -set_serial $RANDOM -sha512 -out $UBIQUITY_DB_CERT_PUBLIC

        chown postgres $UBIQUITY_DB_CERT_PRIVATE
        chmod 600 $UBIQUITY_DB_CERT_PRIVATE
        chown postgres $UBIQUITY_DB_CERT_PUBLIC
        chmod 600 $UBIQUITY_DB_CERT_PUBLIC
    fi
fi

exec docker-entrypoint.sh "$@"

