#!/usr/bin/env bash
# You don't have to use this script to build and push the image.
# This is just for me to build the image once and push it

# make sure an error stops the script
set -e

# We build the amd64 image, giving the docker desktop can run this with Rosetta
docker build --platform=linux/amd64 -t saichler/protoc:latest .

# Push the image to the repository for others to use
docker push saichler/protoc:latest