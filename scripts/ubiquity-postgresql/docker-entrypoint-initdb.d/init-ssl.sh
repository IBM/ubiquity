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

# edit the configuration files - this should be done only after postgresql image initialization!

# Update HBA to require SSL and Client Cert auth
PG_HBA=/var/lib/postgresql/data/pg_hba.conf
head -n -1 $PG_HBA > /tmp/pg_hba.conf
echo "hostssl all all all password" >> /tmp/pg_hba.conf
mv /tmp/pg_hba.conf $PG_HBA

PG_CONF=/var/lib/postgresql/data/postgresql.conf
sed -i 's/#ssl/ssl/g' $PG_CONF
sed -i 's/ssl \= off/ssl \= on/g' $PG_CONF
sed -i "s~ssl_cert_file = 'server.crt'~ssl_cert_file = '$UBIQUITY_DB_CERT_PUBLIC'~g" $PG_CONF
sed -i "s~ssl_key_file = 'server.key'~ssl_key_file = '$UBIQUITY_DB_CERT_PRIVATE'~g" $PG_CONF

echo "SSL Configuration - Done!"

