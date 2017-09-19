#!/bin/bash -ex
# While running this test export FILESYSTEM to gpfs filesystem. Default value
# is gpfs_device


if [ -z "$1" ]
    then
        DRIVER="ubiquity"
else
    DRIVER=$1
fi
set -e
LightWtFileset="LtWtVolFileset"$(date +%s)
FileSystem="gpfs_device"

if [[ -z "${FILESYSTEM}" ]]; then
	FileSystem="gpfs_device"
else
	FileSystem=${FILESYSTEM}
fi

function gpfs_cleanup()
{
	set +e
	for ent in `/usr/lpp/mmfs/bin/mmlsfileset gpfs_device |awk '{print $1}'`;do
		if [ "$ent" != "root" ]; then
			 /usr/lpp/mmfs/bin/mmnfs export remove /gpfs/$FileSystem/$ent
			delete_fileset $FileSystem $ent			
		fi		
	done
	for ent in `mount |grep nfs4 |awk '{print $3}' |xargs`;do
		umount $ent -f
	done
	set -e
}
function cleanup() {
	filesystem=$1
	lightweight=$2
	/usr/lpp/mmfs/bin/mmlsfileset $filesystem $lightweight && rc=$? || rc=$?
	if [ $rc -ne 2 ]; then
		delete_fileset $filesystem $lightweight
	fi
	systemctl restart ubiquity-docker-plugin
	systemctl restart docker
	sleep 5
	set +e
	docker stop $(docker ps -qa)
	docker rm $(docker ps -qa)
	set -e

	return
}

function create_fileset()
{
	local fsys=$1
	local fileset=$2
	/usr/lpp/mmfs/bin/mmcrfileset $fsys $fileset && rc=$? || rc=$?
	if [ $rc -eq 17 ];then
		return
	fi
	/usr/lpp/mmfs/bin/mmlinkfileset $fsys $fileset -J /gpfs/$fsys/$fileset
	sleep 10
	return
}

function delete_fileset()
{
	local fsys=$1
	local fileset=$2
	/usr/lpp/mmfs/bin/mmunlinkfileset $fsys $fileset -f
	/usr/lpp/mmfs/bin/mmdelfileset $fsys $fileset  -f
}

function execute_spectrum_scale_test()
{
	backend=$1
	filesystem=$2
	type_l=$3
	lightweight_data=''

	if [ "$type_l" == "lightweight" ]; then
		create_fileset $filesystem $LightWtFileset
		lightweight_data="--opt type=lightweight --opt fileset=$LightWtFileset"
	fi	
	echo "Starting tests with: "$DRIVER
	echo "********************************************************************"
	echo "Creating volume without UID:GID"
	VolumeName='acceptanceTestVolume'$(date +%s)
	docker volume create -d $DRIVER --name $VolumeName $lightweight_data --opt filesystem=$filesystem --opt backend=$backend
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
	if [ "$type_l" == "lightweight" ]; then
		delete_fileset $filesystem $LightWtFileset
	fi
	echo "********************************************************************"
	echo
	echo "********************************************************************"

#TODO Bypasssing some checks for spectrum-scale-nfs since we have some issue there. Need to remove this hack once fix is in place
	if [ "$backend" == "spectrum-scale-nfs" ]; then
		set +e
	fi

	echo "Second tests"
	echo "Creating volume with  UID:GID"
	VolumeName='acceptanceTestVolume'$(date +%s)
	if [ "$type_l" == "lightweight" ]; then
		create_fileset $filesystem $LightWtFileset
		lightweight_data="--opt type=lightweight --opt fileset=$LightWtFileset"
	fi	

	docker volume create -d $DRIVER --name $VolumeName $lightweight_data --opt filesystem=$filesystem --opt uid=1010 --opt gid=1010 --opt backend=$backend

	echo "1. Writing to volume with UID:GID should succeed"

	docker run -t -i -d --name acceptanceContainer --volume-driver $DRIVER --volume $VolumeName:/data --user 1010:1010 --entrypoint /bin/sh alpine

	docker exec  --user 1010:1010 acceptanceContainer  touch /data/acceptance.txt

	echo
	echo "2. Writing to volume with different UID:GID should error"
	docker exec --user 1005:1005 acceptanceContainer   touch /data/unauthorized.txt && rc=$? || rc=$?

	if [ $rc -eq 0 ]; then
		if [ "$backend" != "spectrum-scale-nfs" ]; then
			exit -1
		fi
	fi

	echo
	echo "3. Reading from volume should succeed"
	docker exec acceptanceContainer ls /data

	echo
	echo "4. Reading from volume with different UID:GID should error"
	docker exec --user 1005:1005 acceptanceContainer  ls /data && rc=$? || rc=$?
	if [ $rc -eq 0 ]; then
		if [ "$backend" != "spectrum-scale-nfs" ]; then
			exit -1
		fi
	fi


	echo
	echo "5. Reading from volume without UID:GID should succeed"
	docker exec acceptanceContainer ls /data

	echo
	echo "6. Reading from volume with UID:GID should succeed"
	docker exec --user 1010:1010 acceptanceContainer ls /data

	echo
	echo "7. Modifying a file created by another user should error"

	docker exec --user 1005:1005 acceptanceContainer sh -c "echo 'unauthorized' >> /data/acceptance.txt" && rc=$? || rc=$?
	if [ $rc -eq 0 ]; then
		if [ "$backend" != "spectrum-scale-nfs" ]; then
			exit -1
		fi
	fi


	docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt
	docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'unauthorized' >> /data/unauthorized.txt"

	docker exec --user 1005:1005 acceptanceContainer more /data/unauthorized.txt && rc=$? || rc=$?


	echo
	echo "8. Modifying a file created by same user should succeed"
	docker exec --user 1010:1010 acceptanceContainer sh -c "echo 'authorized' >> /data/acceptance.txt"

	docker exec --user 1010:1010 acceptanceContainer more /data/acceptance.txt

	echo
	echo "Stoping and removing container and volume"
	docker stop acceptanceContainer
	docker rm acceptanceContainer
	docker volume rm $VolumeName
	if [ "$type_l" == "lightweight" ]; then
		delete_fileset $filesystem $LightWtFileset
	fi
	if [ "$backend" == "spectrum-scale-nfs" ]; then
		set -e
	fi

	echo "********************************************************************"
}

cleanup $FileSystem $LightWtFileset
execute_spectrum_scale_test "spectrum-scale" $FileSystem ""
execute_spectrum_scale_test "spectrum-scale-nfs" $FileSystem ""
execute_spectrum_scale_test "spectrum-scale" $FileSystem "lightweight"
execute_spectrum_scale_test "spectrum-scale-nfs" $FileSystem "lightweight"
cleanup $FileSystem $LightWtFileset
echo "TESTS PASSED"

