#!/bin/bash

SERVICE_NAME=$1

CMD_LOCATION=cmd/$SERVICE_NAME
if [ ! -d "$CMD_LOCATION" ]; then
    go run scripts/service_scaffold/*.go --servicename=$SERVICE_NAME
fi
