#!/bin/bash
set -e

echo "Creating user for ubiquity.."

if [ -z "$UBIQUITY_DB_USER" ]; then
  export UBIQUITY_DB_USER="ubiquity"
fi

if [ -z "$UBIQUITY_DB_PASSWORD" ]; then
  export UBIQUITY_DB_PASSWORD="ubiquity"
fi

if [ -z "$UBIQUITY_DB_NAME" ]; then
  export UBIQUITY_DB_NAME="ubiquity"
fi

if [ -z "$POSTGRES_USER" ]; then
  export POSTGRES_USER="postgres"
fi

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER ${UBIQUITY_DB_USER};
    CREATE DATABASE ${UBIQUITY_DB_NAME};
    GRANT ALL PRIVILEGES ON DATABASE ${UBIQUITY_DB_NAME} TO ${UBIQUITY_DB_USER};
    ALTER ROLE ${UBIQUITY_DB_USER} password '${UBIQUITY_DB_PASSWORD}';
EOSQL

