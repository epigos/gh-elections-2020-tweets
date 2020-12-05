#!/usr/bin/env sh
set -eu

envsubst '${KIBANA_HOST} ${AUTH_TOKEN}' < /etc/nginx/conf.d/default.conf.template > /etc/nginx/conf.d/default.conf

exec "$@"