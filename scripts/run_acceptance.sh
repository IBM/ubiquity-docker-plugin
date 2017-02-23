#!/usr/bin/env bash
echo "First tests"
echo "Creating volume without  UID:GID"
VolumeName='acceptanceTestVolume'$(date +%s)
docker volume create -d spectrum-scale-nfs --name $VolumeName --opt filesystem=gold

echo "Writing to volume without UID:GID should succeed"
docker run -t -i -d --name acceptanceContainer --volume-driver spectrum-scale-nfs --volume $VolumeName:/data --entrypoint /bin/sh alpine
docker exec acceptanceContainer touch /data/acceptance.txt

echo "writing to volume with UID:GID should error"
docker exec --user 1005:1005 acceptanceContainer  touch /data/unauthorized.txt

echo "Reading from volume without UID:GID should succeed"
docker exec acceptanceContainer ls /data

echo "Reading from volume with UID:GUID should error"
docker exec  --user 1005:1005 acceptanceContainer ls /data

echo "Stoping and removing container and volume"
docker stop acceptanceContainer
docker rm acceptanceContainer
docker volume rm $VolumeName

echo "Second tests"
echo "Creating volume with  UID:GID"
VolumeName='acceptanceTestVolume'$(date +%s)
docker volume create -d spectrum-scale-nfs --name $VolumeName --opt filesystem=gold --opt uid=1010 --opt gid=1010

echo "Writing to volume with UID:GID should succeed"

docker run -t -i -d --name acceptanceContainer --volume-driver spectrum-scale-nfs --volume $VolumeName:/data --user 1010:1010 --entrypoint /bin/sh alpine

docker exec  --user 1010:1010 acceptanceContainer  touch /data/acceptance.txt

echo "Writing to volume with different UID:GID should error"
docker exec --user 1005:1005 acceptanceContainer   touch /data/unauthorized.txt

echo "Reading from volume should succeed"
docker exec acceptanceContainer ls /data

echo "Reading from volume with different UID:GID should error"
docker exec --user 1005:1005 acceptanceContainer  ls /data

echo "Reading from volume without UID:GID should succeed"
docker exec acceptanceContainer ls /data

echo "Reading from volume with UID:GID should succeed"
docker exec --user 1010:1010 acceptanceContainer ls /data

echo "Modifying a file created by another user should error"
docker exec --user 1005:1005 acceptanceContainer sh -c "echo 'unauthorized' >> /data/acceptance.txt"

docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt
docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'unauthorized' >> /data/unauthorized.txt"

docker exec --user 1005:1005 acceptanceContainer more /data/unauthorized.txt

echo "Modifying a file created by same user should succeed"
docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'authorized' >> /data/acceptance.txt"

docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt

echo "Stoping and removing container and volume"
docker stop acceptanceContainer
docker rm acceptanceContainer
docker volume rm $VolumeName


