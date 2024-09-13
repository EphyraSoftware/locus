#!/usr/bin/env bash

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  CREATE DATABASE coldmfa;
  CREATE USER coldmfa WITH ENCRYPTED PASSWORD 'coldmfa';
	GRANT ALL PRIVILEGES ON DATABASE coldmfa TO coldmfa;
	\c coldmfa "$POSTGRES_USER"
	GRANT ALL PRIVILEGES ON SCHEMA public TO coldmfa;
EOSQL
