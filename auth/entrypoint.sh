#!/bin/sh
set -e

# Restore from S3 only if the local DB doesn't already exist.
# On first deploy with no replica, this is a no-op (-if-replica-exists).
# On subsequent deploys, the volume already has the DB — skip restore.
if [ ! -f /data/auth.db ]; then
  litestream restore -if-replica-exists -config /etc/litestream.yml /data/auth.db
fi

# Start Litestream as supervisor: runs auth as subprocess,
# continuously replicates WAL changes to S3.
exec litestream replicate -exec "auth" -config /etc/litestream.yml
