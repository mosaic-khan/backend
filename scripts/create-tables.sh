#!/bin/sh

MIGRATION_DIR=internal/storage/schema

export PGPASSWORD=$DB_PASS
DB_ARGS="-h $DB_HOST -U $DB_USER -d $DB_NAME -p $DB_PORT"

for migration in "$MIGRATION_DIR"/*.sql; do
  echo "Applying migration: $migration"

  if ! psql $DB_ARGS -f "$migration"
  then
    echo "Migration failed: $migration"
    exit 1
  fi
done

unset PGPASSWORD

printf "All migrations applied successfully!\n"
