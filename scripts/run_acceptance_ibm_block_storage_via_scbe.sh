#!/bin/bash -ex

############################################
# Acceptance Test for IBM Block Storage via SCBE
# Script prerequisites:
#    1. SCBE server up and running with 1 service delegated to ubiquity interface
#    2. ubiqutiy server up and running with SCBE backend configured
#    3. ubiquity-docker-plugin up and running
#    4. setup connectivity between the docker node to the related storage system of the service.
#
#   Two nodes tests : 
#      In case second node provided a migration tests will take place
#      prerequisites for that is : the second node should apply to #3, #4 and has ssh keys from current node to the second node.
############################################


function basic_tests_on_one_node()
{
	echo "####### ---> ${S}. Verify that no volume attached to the docker node"
	df | egrep "ubiquity|^Filesystem"     || :  
	multipath -ll | grep IBM    || :                    
	lsblk | egrep "ubiquity|^NAME" -B 1 || :
	docker volume ls


	stepinc
	echo "####### ---> ${S}. Create volume on SCBE ${profile} service (which is on IBM FlashSystem A9000R)"
	docker volume create --driver ubiquity --name $vol --opt size=5 --opt profile=${profile}

	echo "####### ---> ${S}.1. Verify volume info"
	docker volume ls
	docker volume inspect $vol    

	echo "## ---> ${S}.2. Verify storage side : verify the volume was created on the relevant pool\service"
	## ssh root@gen4d-67a "xcli.py vol_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run ${CName}1 with the new volume"
	docker run -t -i -d --name ${CName}1 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify the volume was attached to the docker node"
	df | egrep "ubiquity|^Filesystem" 
	multipath -ll | grep IBM
	lsblk | egrep "ubiquity|^NAME" -B 1
	mount |grep ubiquity                            # you should see its ext4

	echo "## ---> ${S}.2. Verify volume exist inside the container"
	docker exec ${CName}1 df | egrep "/data|^Filesystem"

	echo "## ---> ${S}.3. Verify container up and with the mount point"
	docker ps | grep ${CName}1
	docker inspect --format '{{ index .HostConfig.Binds }}' ${CName}1
	docker inspect --format '{{ index .Mounts }}' ${CName}1

	echo "## ---> ${S}.3. Verity the storage side : check volume has mapping to the host"
	## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Write DATA on the volume by create a file in /data inside the container"
	docker exec ${CName}1 touch /data/file_on_A9000_volume
	docker exec ${CName}1 ls -l /data/file_on_A9000_volume

	stepinc
	echo "####### ---> ${S}. Stop the container"
	docker stop ${CName}1

	echo "## ---> ${S}.1. Verify the volume was detached from the docker node"
	df | egrep "ubiquity|^Filesystem" 
	multipath -ll | grep IBM || :
	lsblk | egrep "ubiquity|^NAME" -B 1
	mount |grep ubiquity  || :       

	echo "## ---> ${S}.2. Verify container stoped but volume still exist"
	docker ps | grep ${CName}1 || :
	docker volume ls  

	echo "## ---> ${S}.3. Verity the storage side : check volume is no longer mapped to the hos"
	## ssh root@gen4d-67a "xcli.py vol_mapping_list vol=u_ubiquity_instance1_$vol"


	stepinc
	echo "####### ---> ${S}. Run another container(${CName}2) with the same volume and check the if the data remains"
	docker run -t -i -d --name ${CName}2 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify that the data remains (file exist on the /data inside the container)"
	docker exec ${CName}2 ls -l /data/file_on_A9000_volume


	stepinc
	echo "####### ---> ${S}. Stop the container (of cause will detach the volume from the host)"
	docker stop ${CName}2
	docker rm ${CName}1 ${CName}2  # so we can delete the docker volume

	stepinc
	echo "####### ---> ${S}. Remove the volume"
	docker volume rm $vol
	docker volume ls

	echo "## ---> ${S}.1. Verity the storage side : check volume is no longer exist"
	##  ssh root@[A9000] "xcli.py vol_list vol=u_ubiquity_instance1_$vol"

	stepinc
	echo "####### ---> ${S}. Run container without creating vol in advance"
	docker run -t -i -d --name ${CName}3 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify volume was created for this container and you can touch a file inside the container"
	docker volume ls | grep $vol
	docker exec ${CName}3 touch /data/file_on_A9000_volume
	docker exec ${CName}3 ls -l /data/file_on_A9000_volume

	echo "## ---> ${S}.2. Verify that you stop the container and start the same container so the file still exist"
	docker stop ${CName}3
	docker start ${CName}3
	docker exec ${CName}3 ls -l /data/file_on_A9000_volume

	echo "## ---> ${S}.3 Stop the container and remove the volume"
	docker stop ${CName}3
	docker rm ${CName}3 
	docker volume rm $vol
	docker volume ls | grep -v $vol


	stepinc
	echo "####### ---> ${S}. Run container with 2 volumes"
	docker run -t -i -d --name ${CName}4 --volume-driver ubiquity --volume ${vol}1:/data1 --volume ${vol}2:/data2 --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify volume was created for this container and you can touch a file inside the container"
	docker volume ls | grep ${vol}1
	docker volume ls | grep ${vol}2
	docker exec ${CName}4 df | egrep "/data1|^Filesystem"
	docker exec ${CName}4 df | egrep "/data2|^Filesystem"
	docker exec ${CName}4 touch /data1/file1
	docker exec ${CName}4 touch /data2/file2

	echo "## ---> ${S}.2. Stop container Verify unmount and remove volumes"
	docker stop ${CName}4
	mount |grep ubiquity  && exit ${S} || :
	docker rm ${CName}4
	docker volume rm ${vol}1
	docker volume rm ${vol}2
	docker volume ls | grep -v $vol || :
}

function fstype_basic_check()
{
    for fstype in ext4 xfs; do
        stepinc
        echo "####### ---> ${S}. Create volume with xfs run container and make sure the volume is $fstype"
        docker volume create --driver ubiquity --name $vol --opt size=5 --opt profile=${profile} --opt fstype=$fstype

        echo "## ---> ${S}.1. Verify volume info"
        docker volume ls | grep $vol
        docker volume inspect $vol | grep fstype | grep $fstype

        echo "## ---> ${S}.2 Run container with the volume and Verify it was mounted right"
        docker run -t -i -d --name ${CName}4 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine
        df | egrep "ubiquity|^Filesystem"
        mount |grep ubiquity | grep $fstype
        docker stop ${CName}4
        docker rm ${CName}4
        docker volume rm $vol
    done
}

function one_node_negative_tests()
{
	stepinc
	echo "####### ---> ${S}. some negative"
	echo "## ---> ${S}.1. Should fail to create volume with long name"
	long_vol_name=""; for i in `seq 1 63`; do long_vol_name="$long_vol_name${i}"; done
	docker volume create --driver ubiquity --name $long_vol_name --opt size=5 --opt profile=${profile} && exit 81 || :

	echo "## ---> ${S}.2. Should fail to create volume with wrong size"
	docker volume create --driver ubiquity --name $vol --opt size=10XX --opt profile=${profile} && exit 82 || :

	echo "## ---> ${S}.3. Should fail to create volume on wrong service"
	docker volume create --driver ubiquity --name $vol --opt size=10 --opt profile=${profile}XX && exit 83 || :
}


function tests_with_second_node()
{
	# Assuming plugin runs on second node and with storage connectivity
	echo ""
	echo "######### [2 nodes testing  node1=`hostname`, node2=`$node2`] ###########"
	
	stepinc
	echo "####### ---> ${S}. Run stateful container (should create and run the container)"
	docker run -t -i -d --name ${CName}4 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1. Verify volume was created for this container and you can touch a file inside the container"
	docker volume ls | grep $vol
	docker exec ${CName}4 touch /data/file_on_A9000_volume
	docker exec ${CName}4 ls -l /data/file_on_A9000_volume

	echo "## ---> ${S}.2. [$node2] : Verify volume is visible from second node"
	ssh root@$node2 "docker volume ls | grep $vol"

	echo "## ---> ${S}.3. [$node2] : Verify that you can NOT run container with $vol on second node"
	ssh root@$node2 "docker run -t -i -d --name ${CName}5 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine" && exit 1 || :
	ssh root@$node2 "docker stop ${CName}5"
	ssh root@$node2 "docker rm ${CName}5"

	echo "## ---> ${S}.4. [$node2] : Verify that you can NOT delete the volume $vol from the second node because its already attached to first node"
	ssh root@$node2 "docker volume rm $vol" && exit 1 || :
	ssh root@$node2 "docker volume ls | grep -v $vol" # volume should still be visible on the remote
	docker volume ls| grep -v $vol # and also visible on the local node, so we sure the volume was deleted
	
	stepinc
	echo "####### ---> ${S} Stop the container (so next step can run it on second node)"
	docker stop ${CName}4
	docker rm ${CName}4
	sleep 2 && echo "finished sleep 2 seconds"  # just waiting for detach to complite

	stepinc
	echo "####### ---> ${S} [$node2] : Start the container with the same vol on the second node"
	ssh root@$node2 "docker run -t -i -d --name ${CName}5 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine"

	echo "## ---> ${S}.1 [$node2] : Verify data presiste after migration to second node."
	ssh root@$node2 "docker exec ${CName}5 ls -l /data/file_on_A9000_volume"

	echo "## ---> ${S}.2 [$node2] : And add new file inside the volume."
	ssh root@$node2 "docker exec ${CName}5 touch /data/file_on_A9000_volume_from_node2"

	stepinc
	echo "####### ---> ${S}  [$node2] Stop the container on second node"
	ssh root@$node2 "docker stop ${CName}5"
	ssh root@$node2 "docker rm ${CName}5"

	stepinc
	echo "####### ---> ${S} [$node2] : Start the container with the same vol on the first node"
	docker run -t -i -d --name ${CName}6 --volume-driver ubiquity --volume $vol:/data --entrypoint /bin/sh alpine

	echo "## ---> ${S}.1 Verify data presiste after migration back to first node(check 2 files)."
	docker exec ${CName}6 ls -l /data/file_on_A9000_volume
	docker exec ${CName}6 ls -l /data/file_on_A9000_volume_from_node2

	stepinc
	echo "####### ---> ${S} Stop container and delete vol $vol"
	docker stop ${CName}6
	docker rm ${CName}6
	docker volume rm $vol

	echo "## ---> ${S}.1. [$node2] : Verify volume is no longer visible on the second node"
	ssh root@$node2 "docker volume ls | grep -v $vol " 
}

function stepinc() { S=`expr $S + 1`; }

function setup()
{
    # clean acceptance containers and volumes before start the test and also validate ssh connection to second node if needed.
     conlist=`docker ps -a | grep $CName || :`
    if [ -n "$conlist" ]; then
       echo "Found $CName on the host `hostname`, try to stop and kill them before start the test"
       docker ps -a | grep $CName      
       conlist2=`docker ps -a | sed '1d' | grep $CName | awk '{print $1}'|| :`
       docker stop $conlist2
       docker rm $conlist2
    fi

     volist=`docker volume ls -q | grep $CName || :`
    if [ -n "$volist" ]; then
       echo "Found $CName on the host, try to remove them"
       docker volume rm $volist
    fi

    if [ -n "$node2" ]; then 
	ssh root@$node2 hostname || { echo "Cannot ssh to second host $node2, Aborting."; exit 1; }
        ssh root@$node2 "docker ps -aq | grep $CName" && { echo "need to clean $CName containers on remote node $node2"; exit 2; } || :
        ssh root@$node2 "docker volume ls | grep $CName" && { echo "need to clean $CName volumes on remote node $node2"; exit 3; } || :
    fi
}
[ "$1" = "-h" ] && { echo "$0 can get the following envs :"; echo "        ACCEPTANCE_PROFILE, ACCEPTANCE_WITH_NEGATIVE, ACCEPTANCE_WITH_SECOND_NODE"; exit 0; }

S=0 # steps counter

[ -n "$ACCEPTANCE_PROFILE" ] && profile=$ACCEPTANCE_PROFILE || profile=gold  
[ -n "$ACCEPTANCE_WITH_NEGATIVE" ] && withnegative=$ACCEPTANCE_WITH_NEGATIVE || withnegative=""  
[ -n "$ACCEPTANCE_WITH_SECOND_NODE" ] && node2=$ACCEPTANCE_WITH_SECOND_NODE || node2=""  


CName=acceptance # name of the containers in the script
vol=${CName}Vol
echo "Start Acceptance Test for IBM Block Storage"

setup # Verifications and clean up before the test

basic_tests_on_one_node
fstype_basic_check
[ -n "$withnegative" ] && one_node_negative_tests
[ -n "$node2" ] && tests_with_second_node

echo ""
echo "======================================================"
echo "Successfully Finish The Acceptance test ([$S] steps). Running stateful container on IBM Block Storage."

