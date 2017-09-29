#!/bin/bash

set -e

PLUGIN_NAME="aswarke/ubiquity-docker-plugin"
scripts=$(dirname $0)
PLUGIN_V2_DIR=$scripts/../pluginv2

echo "Creating docker plugin $PLUGIN_NAME"

docker build -f Dockerfile -t $PLUGIN_NAME $scripts/..

id=$(docker create $PLUGIN_NAME true)

rm -rf $PLUGIN_V2_DIR
mkdir -p $PLUGIN_V2_DIR/rootfs

docker export $id | tar -x -C $PLUGIN_V2_DIR/rootfs
docker rm -vf $id
docker rmi $PLUGIN_NAME

cp config.json $PLUGIN_V2_DIR

docker plugin create $PLUGIN_NAME $PLUGIN_V2_DIR

echo "docker plugin $PLUGIN_NAME successfully created"