#!/bin/bash
set -e

if [ ! -z "$UBIQUITY_DB_MAX_CONNECTION" ]
then
       echo "Updating max_connection to $UBIQUITY_DB_MAX_CONNECTION"
       sed -i "s/max_connections = 100/max_connections = ${UBIQUITY_DB_MAX_CONNECTION}/g" /var/lib/postgresql/data/postgresql.conf
fi
