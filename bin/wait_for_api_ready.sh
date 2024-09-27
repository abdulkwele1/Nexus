#!/bin/bash
set -x

until curl -f http://localhost:"${NEXUS_API_HOST_PORT}/healthcheck"
do
    echo "waiting for relay receiver service to be healthy"
    sleep 0.5
done

echo "nexus api is healthy"
