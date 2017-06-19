#!/bin/bash -e

############################################
# Acceptance Test for IBM Block Storage via SCBE
# Script prerequisites:
#    1. SCBE server up and running with 1 service delegated to ubiquity interface
#    2. ubiqutiy server up and running with SCBE backend configured
#    3. ubiquity-docker-plugin up and running
#    4. setup connectivity between the docker node to the related storage system of the service.
############################################

S=0
function stepinc() { S=`expr $S + 1`; }

[ -n "$ACCEPTANCE_PROFILE" ] && profile=$ACCEPTANCE_PROFILE || profile=gold  
[ -n "$ACCEPTANCE_WITH_NEGATIVE" ] && withnegative=$ACCEPTANCE_WITH_NEGATIVE || withnegative=""  

vol=myVol1
echo "Start Acceptance Test - Run stateful container on IBM FlashSystem A9000"

echo "####### ---> ${S}. Verify that no volume attached to the docker node"
df | egrep "ubiquity|^Filesystem"     || :  
multipath -ll | grep IBM    || :                    
lsblk | egrep "ubiquity|^NAME" -B 1 || :
docker volume ls


stepinc
echo "####### ---> ${S}. Create volume on SCBE gold service (which is on IBM FlashSystem A9000R)"
docker volume create --driver ubiquity --name $vol --opt size=5 --opt profile=gold

echo "####### ---> ${S}.1. Verify volume info"
docker volume ls
docker volume inspect $vol    

echo "## ---> ${S}.2. Verify storage side : verify the volume was created on the relevant pool\service"
## ssh root@gen4d-67a "xcli.py vol_list vol=u_ubiquity_instance1_$vol"


stepinc
echo "####### ---> ${S}. Run myContainer1 with the new volume"
docker run -t -i -d --name myContainer1 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

echo "## ---> ${S}.1. Verify the volume was attached to the docker node"
df | egrep "ubiquity|^Filesystem" 
multipath -ll | grep IBM
lsblk | egrep "ubiquity|^NAME" -B 1
mount |grep ubiquity                            # you should see its ext4

echo "## ---> ${S}.2. Verify volume exist inside the container"
docker exec myContainer1 df | egrep "/data|^Filesystem"

echo "## ---> ${S}.3. Verify container up and with the mount point"
docker ps | grep myContainer1
docker inspect --format '{{ index .HostConfig.Binds }}' myContainer1
docker inspect --format '{{ index .Mounts }}' myContainer1

echo "## ---> ${S}.3. Verity the storage side : check volume has mapping to the host"
## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


stepinc
echo "####### ---> ${S}. Write DATA on the volume by create a file in /data inside the container"
docker exec myContainer1 touch /data/file_on_A9000_volume
docker exec myContainer1 ls -l /data/file_on_A9000_volume

stepinc
echo "####### ---> ${S}. Stop the container"
docker stop myContainer1

echo "## ---> ${S}.1. Verify the volume was detached from the docker node"
df | egrep "ubiquity|^Filesystem" 
multipath -ll | grep IBM || :
lsblk | egrep "ubiquity|^NAME" -B 1
mount |grep ubiquity  || :       

echo "## ---> ${S}.2. Verify container stoped but volume still exist"
docker ps | grep myContainer1 || :
docker volume ls  

echo "## ---> ${S}.3. Verity the storage side : check volume is no longer mapped to the hos"
## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


stepinc
echo "####### ---> ${S}. Run another container(myContainer2) with the same volume and check the if the data remains"
docker run -t -i -d --name myContainer2 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

echo "## ---> ${S}.1. Verify that the data remains (file exist on the /data inside the container)"
docker exec myContainer2 ls -l /data/file_on_A9000_volume


stepinc
echo "####### ---> ${S}. Stop the container (of cause will detach the volume from the host)"
docker stop myContainer2
docker rm myContainer1 myContainer2  # so we can delete the docker volume

stepinc
echo "####### ---> ${S}. Remove the volume"
docker volume rm $vol
docker volume ls

echo "## ---> ${S}.1. Verity the storage side : check volume is no longer exist"
##  ssh root@[A9000] "xcli.py vol_list vol=u_ubiquity_instance1_$vol"

stepinc
echo "####### ---> ${S}. Run container without creating vol in advance"
docker run -t -i -d --name myContainer3 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

echo "## ---> ${S}.1. Verify volume was created for this container and you can touch a file inside the container"
docker volume ls | grep $vol
docker exec myContainer3 touch /data/file_on_A9000_volume
docker exec myContainer3 ls -l /data/file_on_A9000_volume

echo "## ---> ${S}.2. Verify that you stop the container and start the same container so the file still exist"
docker stop myContainer3
docker start myContainer3
docker exec myContainer3 ls -l /data/file_on_A9000_volume

echo "## ---> ${S}.3 Stop the container and remove the volume"
docker stop myContainer3
docker rm myContainer3 
docker volume rm $vol
docker volume ls | grep -v $vol


if [ -n "$withnegative" ]; then 
stepinc
echo "####### ---> ${S}. some negative"
echo "## ---> ${S}.1. Should fail to create volume with long name"
long_vol_name=""; for i in `seq 1 63`; do long_vol_name="$long_vol_name${i}"; done
docker volume create --driver ubiquity --name $long_vol_name --opt size=5 --opt profile=gold && exit 81 || :   

echo "## ---> ${S}.2. Should fail to create volume with wrong size"
docker volume create --driver ubiquity --name $vol --opt size=10XX --opt profile=gold && exit 82 || :   

echo "## ---> ${S}.3. Should fail to create volume on wrong service"
docker volume create --driver ubiquity --name $vol --opt size=10 --opt profile=goldXX && exit 83 || : 
fi
echo ""
echo "======================================================"
echo "Successfully Finish The Acceptance test ([$S] steps). Running stateful container on IBM Block Storage."

