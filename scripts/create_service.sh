#!/bin/bash

SERVICE_NAME=$1

mkdir api/graphql/$SERVICE_NAME
mkdir api/jsonschema/$SERVICE_NAME
mkdir api/proto/$SERVICE_NAME

SERVICE_LOCATION=internal/services/$SERVICE_NAME
if [ ! -d "$SERVICE_LOCATION" ]; then
    # copy template to service
    cp -a scripts/service_scaffold/service_template $SERVICE_LOCATION
fi

CMD_LOCATION=cmd/$SERVICE_NAME
if [ ! -d "$CMD_LOCATION" ]; then
    mkdir cmd/$SERVICE_NAME
    cp scripts/service_scaffold/cmd/main.go cmd/$SERVICE_NAME
fi
