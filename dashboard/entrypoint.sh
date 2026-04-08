#!/bin/sh
set -e

if [ ! -f /data/dashboard.db ]; then
  litestream restore -if-replica-exists -config /etc/litestream.yml /data/dashboard.db
fi

exec litestream replicate -exec "node server.js" -config /etc/litestream.yml
