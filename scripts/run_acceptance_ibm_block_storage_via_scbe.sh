#!/bin/bash -e

############################################
# Acceptance Test for IBM Block Storage via SCBE
# Script prerequisites:
#    1. SCBE server up and running with 1 service delegated to ubiquity interface
#    2. ubiqutiy server up and running with SCBE backend configured
#    3. ubiquity-docker-plugin up and running
#    4. setup connectivity between the docker node to the related storage system of the service.
############################################

[ $# -eq 1 ] && profile=$1 || profile=gold  
vol=myVol1
echo "Start Acceptance Test - Run stateful container on IBM FlashSystem A9000"

echo "####### ---> 0. Verify that no volume attached to the docker node"
df | egrep "ubiquity|^Filesystem"     || :  
multipath -ll | grep IBM    || :                    
lsblk | egrep "ubiquity|^NAME" -B 1 || :
docker volume ls



echo "####### ---> 1. Create volume on SCBE gold service (which is on IBM FlashSystem A9000R)"
docker volume create --driver ubiquity --name $vol --opt size=5 --opt profile=gold

echo "####### ---> 1.1. Verify volume info"
docker volume ls
docker volume inspect $vol    

echo "## ---> 1.2. Verify storage side : verify the volume was created on the relevant pool\service"
## ssh root@gen4d-67a "xcli.py vol_list vol=u_ubiquity_instance1_$vol"



echo "####### ---> 2. Run myContainer1 with the new volume"
docker run -t -i -d --name myContainer1 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

echo "## ---> 2.1. Verify the volume was attached to the docker node"
df | egrep "ubiquity|^Filesystem" 
multipath -ll | grep IBM
lsblk | egrep "ubiquity|^NAME" -B 1
mount |grep ubiquity                            # you should see its ext4

echo "## ---> 2.2. Verify volume exist inside the container"
docker exec myContainer1 df | egrep "/data|^Filesystem"

echo "## ---> 2.3. Verify container up and with the mount point"
docker ps | grep myContainer1
docker inspect --format '{{ index .HostConfig.Binds }}' myContainer1
docker inspect --format '{{ index .Mounts }}' myContainer1

echo "## ---> 2.3. Verity the storage side : check volume has mapping to the host"
## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"



echo "####### ---> 3. Write DATA on the volume by create a file in /data inside the container"
docker exec myContainer1 touch /data/file_on_A9000_volume
docker exec myContainer1 ls -l /data/file_on_A9000_volume

echo "####### ---> 4. Stop the container"
docker stop myContainer1

echo "## ---> 4.1. Verify the volume was detached from the docker node"
df | egrep "ubiquity|^Filesystem" 
multipath -ll | grep IBM || :
lsblk | egrep "ubiquity|^NAME" -B 1
mount |grep ubiquity  || :       

echo "## ---> 4.2. Verify container stoped but volume still exist"
docker ps | grep myContainer1 || :
docker volume ls  

echo "## ---> 4.3. Verity the storage side : check volume is no longer mapped to the hos"
## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"



echo "####### ---> 5. Run another container(myContainer2) with the same volume and check the if the data remains"
docker run -t -i -d --name myContainer2 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

echo "## ---> 5.1. Verify that the data remains (file exist on the /data inside the container)"
docker exec myContainer2 ls -l /data/file_on_A9000_volume



echo "####### ---> 6. Stop the container (of cause will detach the volume from the host)"
docker stop myContainer2
docker rm myContainer1 myContainer2  # so we can delete the docker volume

echo "####### ---> 7. Remove the volume"
docker volume rm $vol
docker volume ls

echo "####### ---> 7.1. Verity the storage side : check volume is no longer exist"
##  ssh root@[A9000] "xcli.py vol_list vol=u_ubiquity_instance1_$vol"

echo "Successfully Finish The Acceptance test. Running stateful container on IBM Block Storage."

