#!/usr/bin/env bash
echo Generating Protos
protoc --go_out=. overlay.proto