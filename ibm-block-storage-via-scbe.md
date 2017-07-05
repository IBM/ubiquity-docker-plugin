# IBM Block Storage System via Spectrum Control Base Edition

## Configuring Docker host for IBM block storage systems
Configure the following steps(1-4) on each node in the Docker Swarm cluster that requires access to Ubiquity volumes.

#### 1. Installing connectivity packages 
The plugin supports FC or iSCSI connectivity to the storage systems.

  * Redhat \ SLES
  
```bash
   sudo yum -y install sg3_utils
   sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
```

#### 2. Configuring multipathing 
The plugin requires multipath devices. Configure the `multipath.conf` file according to the storage system requirments.
  * Redhat \ SLES
  
```bash
   yum install device-mapper-multipath
   sudo modprobe dm-multipath

   cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
   systemctl start multipathd
   systemctl status multipathd  # Make sure its active
   multipath -ll  # Make sure no error appear.
```

#### 3. Configure storage system connectivity
  *  Verify that the hostname of the Docker node is defined on the relevant storage systems with the valid WWPNs or IQN of the node. Otherwise, you will not be able to run stateful containers.

  *  For iSCSI, discover and log in to the iSCSI targets of the relevant storage systems:

```bash
   iscsiadm -m discoverydb -t st -p ${storage system iSCSI portal IP}:3260 --discover   # To discover targets
   iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login              # To log in to targets
```
            
### 4. Configuring Ubiquity Docker volume plugin for SCBE

The ubiquity-client.conf must be created in the /etc/ubiquity directory. Configure the plugin by editing the file, as illustrated below.

 
 ```toml
 logPath = "/tmp/ubiquity"            # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.
 backends = ["scbe"]                  # The Storage system backend to be used with Ubiquity to create and manage volumes. In this we configure Docker plugin to create volumes using IBM Block Storage system via SCBE.
 
 [DockerPlugin]
 port = 9000                                # Port to serve docker plugin functions
 pluginsDirectory = "/etc/docker/plugins/"  # Point to the location of the configured Docker plugin directory (create if not already created by Docker)
 
 
 [UbiquityServer]
 address = "IP"  # IP/hostname of the Ubiquity Service
 port = 9999     # TCP port on which the Ubiquity Service is listening
 ```
 
## Plugin usage example

### Creating a Docker volume
Docker volume creation template:
```bash
docker volume create --driver ubiquity --name [VOL NAME] --opt size=[number in GB] --fstype=[xfs|ext4] --opt profile=[SCBE service name]
```

For example, to create a volume named volume1 with 10gb size from the gold SCBE storage service, such as a pool from IBM FlashSystem A9000R and with QoS capability:

```bash
#> docker volume create --driver ubiquity --name volume1 --opt size=10 --opt fstype=xfs --opt profile=gold
```

### Display a Docker volume

You can list and inspect the newly created volume by the following command :
```bash
#> docker volume ls
DRIVER              VOLUME NAME
ubiquity            volume1


#> docker volume inspect demo1
[
    {
        "Driver": "ubiquity",
        "Labels": {},
        "Mountpoint": "/",
        "Name": "volume1",
        "Options": {
            "fstype": "xfs",
            "profile": "gold",
            "size": "10"
        },
        "Scope": "local",
        "Status": {
            "LogicalCapacity": "10000000000",
            "Name": "u_instance_volume1",
            "PhysicalCapacity": "10234101760",
            "PoolName": "gold_ubiquity",
            "Profile": "gold",
            "StorageName": "A9000R system1",
            "StorageType": "2810XIV",
            "UsedCapacity": "10485760",
            "Wwn": "6001738CFC9035EB0000000000CFF306",
            "fstype": "xfs"
        }
    }
]

```

### Run a Docker container with a volume
Docker run template:
```bash
#> docker run -it -d --name [CONTAINER NAME] --volume-driver ubiquity -v [VOL NAME]:[PATH TO MOUNT] [DOCKER IMAGE] [CMD]
```

For example, to run a container `container1` with the created volume `volume1` based on alpine Docker image and running `bash` command for the fun.

```bash
#> docker run -it -d --name container1 --volume-driver ubiquity -v volume1:/data alpine bash
```

You can display the new mountpoint and multipath device inside the container and of cause to write data inside this A9000R presistant volume
```bash
#> docker exec container1 df | egrep "/data|^Filesystem"
Filesystem           1K-blocks      Used Available Use% Mounted on
/dev/mapper/mpathacg   9755384     32928   9722456   0% /data

#> docker exec container1 mount | egrep "/data"
/dev/mapper/mpathacg on /data type xfs (rw,seclabel,relatime,attr2,inode64,noquota)

#> docker exec container1 touch /data/FILE
#> docker exec container1 ls /data/FILE
```

You can also see the new attached volume on the host
```bash
#> multipath -ll
mpathacg (36001738cfc9035eb0000000000cbb306) dm-8 IBM     ,2810XIV         
size=9.3G features='1 queue_if_no_path' hwhandler='0' wp=rw
`-+- policy='service-time 0' prio=1 status=active
  |- 8:0:0:1 sdb 8:16 active ready running
  `- 9:0:0:1 sdc 8:32 active ready running

#> mount |grep ubiquity
/dev/mapper/mpathacg on /ubiquity/6001738CFC9035EB0000000000CBB306 type xfs (rw,relatime,seclabel,attr2,inode64,noquota)

#> df | egrep "ubiquity|^Filesystem" 
Filesystem                       1K-blocks    Used Available Use% Mounted on
/dev/mapper/mpathacg               9755384   32928   9722456   1% /ubiquity/6001738CFC9035EB0000000000CFF306

#> docker inspect --format '{{ index .Mounts }}' container1
[{volume volume1 /ubiquity/6001738CFC9035EB0000000000CBB306 /data ubiquity  true }]

```

### Stop a Docker container with a volume
Docker stop   (Ubiquity will detach the volume from the host)
```bash
#> docker stop container1
```

### Remove a Docker volume
Note : to remove a volume you first need to remove the container.
```bash
#> docker rm container1
container1

#> docker volume rm volume1
volume1
```
