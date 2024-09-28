#!/bin/bash
set -x

until curl -f http://localhost:"${NEXUS_API_HOST_PORT}/healthcheck"
do
    echo "waiting for nexus api to be healthy"
    sleep 0.5
done

echo "nexus api is healthy"
