#!/usr/bin/env bash
docker run --mount type=bind,source="$PWD",target=/home/proto -it hub.docker.com/saichler/protoc

mkdir -p ../go/model
mv ./model/overlay.pb.go ../go/model/.
rm -rf ./model