#!/bin/bash
docker compose \
    -f docker-compose.main.yml \
    -f ./docker-compose.tempo.yml \
    -f ./docker-compose.loki.yml \
    -f ./docker-compose.exemplars.yml up -d
