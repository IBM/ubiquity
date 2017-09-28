#!/bin/sh
set -e

export SSLDIR="/var/lib/ubiquity/ssl"

# to be used when ubiquity will not run as root
export USER=root
export GROUP=root
export EMAIL="user@test.com"

configure_ssl() {
    if [ ! -d $SSLDIR ]
    then
        # Create SSL directory
        mkdir -p $SSLDIR
        chown ${USER}:${GROUP} $SSLDIR
    fi

    # Create SSL certificates
    cd $SSLDIR

    if [ ! -s "$SSLDIR/server.crt" ]
    then
        echo "Creating default SSL certificates for ubiquity"
        # root CA
        openssl req -new -x509 -nodes -out root.crt -keyout root.key -newkey rsa:4096 -sha512 -subj /CN=TheRootCA
        # chown ${USER}:${GROUP} root.key
        chmod 600 root.key

        # Server certificate
        openssl req -new -out server.req -keyout server.key -nodes -newkey rsa:4096 -subj "/CN=$( hostname )/emailAddress=$EMAIL"
        openssl x509 -req -in server.req -CAkey root.key -CA root.crt -set_serial $RANDOM -sha512 -out server.crt

        chown ${USER}:${GROUP} server.key
        chmod 600 server.key
        echo "Creating default SSL certificates for ubiquity - done!"
    fi

}

if [ "$1" = 'ubiquity' ]
then
    configure_ssl
fi

echo "Calling $@..."

exec "$@"

