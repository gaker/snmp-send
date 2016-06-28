#!/bin/bash

if [ $1 ]; then
   VERSION=$1
else
   echo "A new version is required!"
   exit 1
fi

docker build -t snmp-send-builder -f ./docker-files/Dockerfile.build .
docker run --rm snmp-send-builder | docker build -t gaker/snmp-send:latest -t gaker/snmp-send:$1 -f Dockerfile.run -


