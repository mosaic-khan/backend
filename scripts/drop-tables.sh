#!/bin/sh

export PGPASSWORD=$DB_PASS

DB_ARGS="-h $DB_HOST -U $DB_USER -d $DB_NAME -p $DB_PORT"


TABLES=$(psql $DB_ARGS -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';" -A -t)

DROPPED_ALL=true

for TABLE in $TABLES
do
  DROP_RESULT=$(psql $DB_ARGS -c "DROP TABLE \"$TABLE\" CASCADE;")
  if [ $? -ne 0 ]; then
    echo "Error: Failed on dropping table $TABLE"
    DROPPED_ALL=false
  fi
done

unset PGPASSWORD

if $DROPPED_ALL; then
  echo "All tables have been removed from the $DB_NAME database."
else
  echo "ERROR: Some tables could not be removed from the $DB_NAME database."
fi