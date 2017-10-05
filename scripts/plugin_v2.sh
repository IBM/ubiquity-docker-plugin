#!/bin/bash

set -xe

PLUGIN_NAME="aswarke/ubiquity-docker-plugin"
TAG="v1"
scripts=$(dirname $0)
PLUGIN_V2_DIR=$scripts/../pluginv2

echo "Creating docker plugin $PLUGIN_NAME:$TAG"

docker build -f Dockerfile -t $PLUGIN_NAME:$TAG $scripts/..

id=$(docker create $PLUGIN_NAME:$TAG true)

rm -rf $PLUGIN_V2_DIR
mkdir -p $PLUGIN_V2_DIR/rootfs

docker export $id | tar -x -C $PLUGIN_V2_DIR/rootfs
docker rm -vf $id
docker rmi $PLUGIN_NAME:$TAG

cp config.json $PLUGIN_V2_DIR

docker plugin create $PLUGIN_NAME:$TAG $PLUGIN_V2_DIR

echo "docker plugin $PLUGIN_NAME:$TAG successfully created"
