#!/usr/bin/env bash

if [ -z "$1" ]
    then
        DRIVER="ubiquity"
else
    DRIVER=$1
fi


echo "Starting tests with: "$DRIVER
echo "********************************************************************"
echo "Creating volume without UID:GID"
VolumeName='acceptanceTestVolume'$(date +%s)
docker volume create -d $DRIVER --name $VolumeName --opt filesystem=gold

echo "1. Writing to volume without UID:GID should succeed"
docker run -t -i -d --name acceptanceContainer --volume-driver $DRIVER --volume $VolumeName:/data --entrypoint /bin/sh alpine
docker exec acceptanceContainer touch /data/acceptance.txt

echo
echo "2. Writing to volume with UID:GID should succeed"
docker exec --user 1005:1005 acceptanceContainer  touch /data/unauthorized.txt

echo
echo "3. Reading from volume without UID:GID should succeed"
docker exec acceptanceContainer ls /data

echo
echo "4. Reading from volume with UID:GUID should succeed"
docker exec  --user 1005:1005 acceptanceContainer ls /data

echo
echo "Stoping and removing container and volume"
docker stop acceptanceContainer
docker rm acceptanceContainer
docker volume rm $VolumeName
echo "********************************************************************"
echo
echo "********************************************************************"
echo "Second tests"
echo "Creating volume with  UID:GID"
VolumeName='acceptanceTestVolume'$(date +%s)
docker volume create -d $DRIVER --name $VolumeName --opt filesystem=gold --opt uid=1010 --opt gid=1010

echo "1. Writing to volume with UID:GID should succeed"

docker run -t -i -d --name acceptanceContainer --volume-driver $DRIVER --volume $VolumeName:/data --user 1010:1010 --entrypoint /bin/sh alpine

docker exec  --user 1010:1010 acceptanceContainer  touch /data/acceptance.txt

echo
echo "2. Writing to volume with different UID:GID should error"
docker exec --user 1005:1005 acceptanceContainer   touch /data/unauthorized.txt

echo
echo "3. Reading from volume should succeed"
docker exec acceptanceContainer ls /data

echo
echo "4. Reading from volume with different UID:GID should error"
docker exec --user 1005:1005 acceptanceContainer  ls /data

echo
echo "5. Reading from volume without UID:GID should succeed"
docker exec acceptanceContainer ls /data

echo
echo "6. Reading from volume with UID:GID should succeed"
docker exec --user 1010:1010 acceptanceContainer ls /data

echo
echo "7. Modifying a file created by another user should error"
docker exec --user 1005:1005 acceptanceContainer sh -c "echo 'unauthorized' >> /data/acceptance.txt"
docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt
docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'unauthorized' >> /data/unauthorized.txt"

docker exec --user 1005:1005 acceptanceContainer more /data/unauthorized.txt

echo
echo "8. Modifying a file created by same user should succeed"
docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'authorized' >> /data/acceptance.txt"

docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt

echo
echo "Stoping and removing container and volume"
docker stop acceptanceContainer
docker rm acceptanceContainer
docker volume rm $VolumeName
echo "********************************************************************"

