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
Volume Creation using SCBE supported IBM Block Storage system
[Ubiquity service](https://github.com/IBM/ubiquity) communicates with the IBM block storage systems through IBM Spectrum Control Base Edition([SCBE](http://www.ibm.com/support/knowledgecenter/STWMS9/landing/IBM_Spectrum_Control_Base_Edition_welcome_page.html)).
The plugin can provision a volume from a delegated SCBE storage service by using the --opt=<SCBE storage service name> flag.

### Creating volume on gold SCBE storage service
Create a volume named demo11 with 10gb size from the gold SCBE storage service (the gold service could be, for example, a pool from IBM FlashSystem A9000\R and with high QoS capability) :

```bash
docker volume create -d ubiquity --name demo11 --opt size=10 --opt profile=gold
```
