# IBM Block Storage System via Spectrum Control Base Edition

## Configuring Docker host for IBM block storage systems

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
            
## Configuring Ubiquity Docker volume plugin for SCBE

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

For example, to create a volume named demo1 with 10gb size from the gold SCBE storage service, such as a pool from IBM FlashSystem A9000R and with QoS capability:

```bash
#> docker volume create --driver ubiquity --name demo1 --opt size=10 --opt fstype=xfs --opt profile=gold
```

You can list and inspect the newly created volume by the following command :
```bash
#> docker volume ls
DRIVER              VOLUME NAME
ubiquity            demo1
#> docker volume inspect demo1
[
    {
        "Driver": "ubiquity",
        "Labels": {},
        "Mountpoint": "/",
        "Name": "demo1",
        "Options": {
            "fstype": "xfs",
            "profile": "gold",
            "size": "10"
        },
        "Scope": "local",
        "Status": {
            "LogicalCapacity": "10000000000",
            "Name": "u_ubiquityPOC_demo1",
            "PhysicalCapacity": "10234101760",
            "PoolName": "gold_ubiquity_9.151.162.17",
            "Profile": "gold",
            "StorageName": "XIV Gen4d-67d",
            "StorageType": "2810XIV",
            "UsedCapacity": "0",
            "Wwn": "6001738CFC9035EB0000000000CBB305",
            "fstype": "xfs"
        }
    }
]
```
