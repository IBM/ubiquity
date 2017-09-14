#!/bin/bash
set -e

echo "Creating user for ubiquity.."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE USER ubiquity;
    CREATE DATABASE ubiquity;
    GRANT ALL PRIVILEGES ON DATABASE ubiquity TO ubiquity;
    ALTER ROLE ubiquity password 'ubiquity';
EOSQL

