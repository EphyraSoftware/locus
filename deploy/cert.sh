#!/usr/bin/env bash

docker compose cp \
    proxy:/data/caddy/pki/authorities/local/root.crt \
    ./root.crt
