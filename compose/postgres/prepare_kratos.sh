#!/usr/bin/env bash

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE DATABASE kratos;
  CREATE USER kratos WITH ENCRYPTED PASSWORD 'kratos';
	GRANT ALL PRIVILEGES ON DATABASE kratos TO kratos;
	\c kratos "$POSTGRES_USER"
	GRANT ALL PRIVILEGES ON SCHEMA public TO kratos;
EOSQL
