#!/bin/bash
set -e

echo "Creating user for ubiquity.."

if [ -z "$UBIQUITY_DB_USERNAME" ]; then
  export UBIQUITY_DB_USERNAME="ubiquity"
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

echo "Creating DB user ${UBIQUITY_DB_USERNAME}, DB name ${UBIQUITY_DB_NAME}, Grant privileges and set password to the DB user."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER ${UBIQUITY_DB_USERNAME};
    CREATE DATABASE ${UBIQUITY_DB_NAME};
    GRANT ALL PRIVILEGES ON DATABASE ${UBIQUITY_DB_NAME} TO ${UBIQUITY_DB_USERNAME};
    ALTER ROLE ${UBIQUITY_DB_USERNAME} password '${UBIQUITY_DB_PASSWORD}';
EOSQL
echo "DB user and DB name was created successfully"
