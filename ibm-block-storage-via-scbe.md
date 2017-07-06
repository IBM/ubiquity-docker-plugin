# IBM Block Storage System via Spectrum Control Base Edition

IBM block storage can be used as persistent storage for Docker via Ubiquity service.
Ubiquity communicates with the IBM storage systems through [IBM Spectrum Control Base Edition](https://www.ibm.com/support/knowledgecenter/en/STWMS9) (SCBE) 3.2.0. SCBE creates a storage profile (for example, gold, silver or bronze) and makes it available for Ubiquity Docker volume plugin.
Avilable IBM block storage systems for Docker volume plugin are listed in the [Ubiquity Service](https://github.com/IBM/ubiquity/).

## Configuring Docker host for IBM block storage systems
Perform the following installation and configuration procedures on each node in the Docker Swarm cluster that requires access to Ubiquity volumes.


#### 1. Installing connectivity packages 
The plugin supports FC or iSCSI connectivity to the storage systems.

  * RHEL, SLES
  
```bash
   sudo yum -y install sg3_utils
   sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
```

#### 2. Configuring multipathing 
The plugin requires multipath devices. Configure the `multipath.conf` file according to the storage system requirments.
  * RHEL, SLES
  
```bash
   yum install device-mapper-multipath
   sudo modprobe dm-multipath

   cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
   systemctl start multipathd
   systemctl status multipathd  # Make sure its active
   multipath -ll  # Make sure no error appear.
```

#### 3. Configure storage system connectivity
  *  Verify that the hostname of the Docker node is defined on the relevant storage systems with the valid WWPNs or IQN of the node. The hostname on the storage system must be the same as the output of `hostname` command on the Docker node. Otherwise, you will not be able to run stateful containers.

  *  For iSCSI, discover and log in to the iSCSI targets of the relevant storage systems:

```bash
   iscsiadm -m discoverydb -t st -p ${storage system iSCSI portal IP}:3260 --discover   # To discover targets
   iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login              # To log in to targets
```

### 4. Opening TCP ports to Ubiquity server
Ubiquity server listens on TCP port (by default 9999) to receive plugin requests, such as creating a new volume. Verify that the Docker node can access this Ubiquity server port.

### 5. Configuring Ubiquity Docker volume plugin for SCBE

The ubiquity-client.conf must be created in the /etc/ubiquity directory. Configure the plugin by editing the file, as illustrated below.

 
 ```toml
 logPath = "/var/tmp"                 # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.
 backends = ["scbe"]                  # The Storage system backend to be used with Ubiquity to create and manage volumes. In this we configure Docker plugin to create volumes using IBM Block Storage system via SCBE.
 
 [DockerPlugin]
 port = 9000                                # Port to serve docker plugin functions
 pluginsDirectory = "/etc/docker/plugins/"  # Point to the location of the configured Docker plugin directory (create if not already created by Docker)
 
 
 [UbiquityServer]
 address = "IP"  # IP/hostname of the Ubiquity Service
 port = 9999     # TCP port on which the Ubiquity Service is listening
 ```
  * Verify that the logPath, exists on the host before you start the plugin.
  * Verify that the pluginsDirectory, exists on the host before you start the plugin. Default location is /etc/docker/plugins/.
  ```bash
        mkdir /etc/docker/plugins
 ```
  
 

## Plugin usage example

### Simple flow of stateful container with Ubiquity volume
The flow is:
1. Create volume `demoVol` on `gold` SCBE storage service, 
2. Run docker `container1` with volume `demoVol` on `/data` mountpoint inside the `container1`
3. Create file `/data/myDATA` in the `container1`
4. Exit `container1` and then start new `container2` and the file `/data/myDATA` exist!
5. Clean up by exist `container2`, remove the containers and delete the volume `demoVol`

```bash
[root@docker-c1-n1 tmp]# docker volume create --driver ubiquity --name demoVol --opt size=10 --opt fstype=xfs --opt profile=gold
demoVol

[root@docker-c1-n1 tmp]# docker volume ls
DRIVER              VOLUME NAME
ubiquity            demoVol
[root@docker-c1-n1 tmp]# docker run -it --name container1 --volume-driver ubiquity -v demoVol:/data alpine sh

/ # df | egrep "/data|^Filesystem"
Filesystem           1K-blocks      Used Available Use% Mounted on
/dev/mapper/mpathaci   9755384     32928   9722456   0% /data

/ # touch /data/myDATA
/ # ls -l /data/myDATA
-rw-r--r--    1 root     root             0 Jul  6 07:59 /data/myDATA
/ # exit

[root@docker-c1-n1 tmp]# docker run -it --name container2 --volume-driver ubiquity -v demoVol:/data alpine sh

/ # ls -l /data/myDATA
-rw-r--r--    1 root     root             0 Jul  6 07:59 /data/myDATA
/ # exit

[root@docker-c1-n1 tmp]# docker rm container1 container2
container1
container2

[root@docker-c1-n1 tmp]# docker volume rm demoVol
demoVol
```

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
#> docker run -it -d --name container1 --volume-driver ubiquity -v volume1:/data alpine sh
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

Now you can start the same container to see that the data still p
```bash
#> docker run -it -d --name container1 --volume-driver ubiquity -v volume1:/data alpine bash
```

### Remove a Docker volume
Note : to remove a volume you first need to remove the container.
```bash
#> docker rm container1
container1

#> docker volume rm volume1
volume1
```

### Docker Compose 
Below a docker compose that describe web app and postgress app.

```bash
version: "3"

volumes:
   postgres:
      driver: "ubiquity"
      driver_opts:
        size: "2"
        profile: "gold"

services:
   web:
     image: shaybery/docker_and_ibm_storage
     ports:
        -  "80:80"
     environment:
        - "USE_POSTGRES_HOST=postgres"
        - "POSTGRES_USER=ubiquity"
        - "POSTGRES_PASSWORD=ubiquitydemo"
     network_mode: "bridge"
     links:
        - "postgres:postgres"
   postgres:
     image: postgres:9.5
     ports:
        -  "5432:5432"
     environment:
        - "POSTGRES_USER=ubiquity"
        - "POSTGRES_PASSWORD=ubiquitydemo"
        - "POSTGRES_DB=postgres"
        - "PGDATA=/var/lib/postgresql/data/data"
     network_mode: "bridge"
     volumes:
        - 'postgres:/var/lib/postgresql/data'
```

docker-compose -f docker-compose.yml up

## Troubleshooting
### Server error
If the `bad status code 500 INTERNAL SERVER ERROR` error is displayed, check the `/var/log/sc/hsgsvr.log` log file on the SCBE node for explanation.
