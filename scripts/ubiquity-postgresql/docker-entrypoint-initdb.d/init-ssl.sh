#!/usr/bin/env bash

#
# This script is called upon first run of a container from this image.
# It will:
# - create SSL certificate and key and configure postgresql to use them
# - use existing certificate and key if these are found (from secrets or mounted volume)
# - configure postgresql to enforce the use of SSL but not to verify the client certificate
#
set -e

echo "Configuring Postgres for SSL!"
echo "Running as id $(id -u)"

if [ -z "$POSTGRES_EMAIL" ]; then
  export POSTGRES_EMAIL="user@test.com"
fi

# Create SSL certificates
cd $PGSSL

if [ ! -s "$PGSSL/server.crt" ]
then
    # root CA
    openssl req -new -x509 -nodes -out root.crt -keyout root.key -newkey rsa:4096 -sha512 -subj /CN=TheRootCA
    chown postgres root.key
    chmod 600 root.key

    # Server certificate
    openssl req -new -out server.req -keyout server.key -nodes -newkey rsa:4096 -subj "/CN=$( hostname )/emailAddress=$POSTGRES_EMAIL"
    openssl x509 -req -in server.req -CAkey root.key -CA root.crt -set_serial $RANDOM -sha512 -out server.crt

    chown postgres server.key
    chmod 600 server.key
fi

# edit the configuration files

# Update HBA to require SSL and Client Cert auth
head -n -1 /var/lib/postgresql/data/pg_hba.conf > /tmp/pg_hba.conf
echo "hostssl all all all password" >> /tmp/pg_hba.conf
mv /tmp/pg_hba.conf /var/lib/postgresql/data/pg_hba.conf

sed -i 's/#ssl/ssl/g' /var/lib/postgresql/data/postgresql.conf
sed -i 's/ssl \= off/ssl \= on/g' /var/lib/postgresql/data/postgresql.conf
sed -i "s/ssl_cert_file = 'server.crt'/ssl_cert_file = '\/var\/lib\/postgresql\/ssl\/server.crt'/g" /var/lib/postgresql/data/postgresql.conf
sed -i "s/ssl_key_file = 'server.key'/ssl_key_file = '\/var\/lib\/postgresql\/ssl\/server.key'/g" /var/lib/postgresql/data/postgresql.conf

echo "SSL Configuration - Done!"

