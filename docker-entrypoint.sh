#!/bin/sh
set -e

export SSL_PRIVATE_DIR=`dirname $UBIQUITY_SERVER_CERT_PRIVATE`
export SSL_PUBLIC_DIR=`dirname $UBIQUITY_SERVER_CERT_PUBLIC`

# to be used when ubiquity will not run as root
export USER=root
export GROUP=root
export EMAIL="user@test.com"

configure_ssl() {
    echo "Creating SSL directory $SSL_PRIVATE_DIR and setting ownership to user postgres ..."
    mkdir -p $SSL_PRIVATE_DIR
    chown ${USER}:${GROUP} $SSL_PRIVATE_DIR
    chmod 700 $SSL_PRIVATE_DIR

    if [ "$SSL_PUBLIC_DIR" != "$SSL_PRIVATE_DIR" ]; then
        echo "Creating SSL directory $SSL_PUBLIC_DIR and setting ownership to user postgres ..."
        mkdir -p $SSL_PUBLIC_DIR
        chown ${USER}:${GROUP} $SSL_PUBLIC_DIR
        chmod 700 $SSL_PUBLIC_DIR
    fi

    if [ ! -s "$UBIQUITY_SERVER_CERT_PUBLIC" ]
    then
        echo "Creating default SSL certificates for ubiquity"
        cd $SSL_PRIVATE_DIR

        # root CA
        openssl req -new -x509 -nodes -out root.crt -keyout root.key -newkey rsa:4096 -sha512 -subj /CN=TheRootCA
        chown ${USER}:${GROUP} root.key
        chmod 600 root.key

        # Server certificate
        openssl req -new -out server.req -keyout $UBIQUITY_SERVER_CERT_PRIVATE -nodes -newkey rsa:4096 -subj "/CN=$( hostname )/emailAddress=$EMAIL"
        openssl x509 -req -in server.req -CAkey root.key -CA root.crt -set_serial $RANDOM -sha512 -out $UBIQUITY_SERVER_CERT_PUBLIC

        chown ${USER}:${GROUP} $UBIQUITY_SERVER_CERT_PRIVATE
        chmod 600 $UBIQUITY_SERVER_CERT_PRIVATE
        chown ${USER}:${GROUP} $UBIQUITY_SERVER_CERT_PUBLIC
        chmod 600 $UBIQUITY_SERVER_CERT_PUBLIC
        echo "Creating default SSL certificates for ubiquity - done!"
    fi

}

if [ "$1" = 'ubiquity' ]
then
    configure_ssl
fi

echo "Calling $@..."

exec "$@"

