#!/usr/bin/env bash
docker run --mount type=bind,source="$PWD",target=/home/proto -it saichler/protoc:latest

mkdir -p ../go/model
mv ./model/overlay.pb.go ../go/model/.
rm -rf ./model