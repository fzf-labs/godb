#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$ROOT_DIR"

PGHOST=${PGHOST:-127.0.0.1}
PGPORT=${PGPORT:-5432}
PGUSER=${PGUSER:-postgres}
PGPASSWORD=${PGPASSWORD:-123456}
export PGPASSWORD

psql_base=(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -v ON_ERROR_STOP=1)

create_db() {
	local db="$1"
	"${psql_base[@]}" -d postgres -c "DROP DATABASE IF EXISTS \"$db\";"
	"${psql_base[@]}" -d postgres -c "CREATE DATABASE \"$db\";"
}

create_db gorm_gen
create_db user
create_db fkratos_sys

"${psql_base[@]}" -d gorm_gen -f scripts/ci/bootstrap-postgres.sql

"${psql_base[@]}" -d fkratos_sys -c 'CREATE TABLE IF NOT EXISTS sys_admin_202301 (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);'
