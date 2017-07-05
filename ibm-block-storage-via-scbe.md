# IBM Block Storage System via SCBE

* [Deployment Prerequisities](#deployment-prerequisities)
* [Configuration](#configuring-ubiquity-docker-plugin-with-ubiquity-and-scbe)
* [Volume Creation](#volume-creation-using-scbe-supported-ibm-block-storage-system)

## Plugin installation prerequisities for IBM block storage system
Configuring storage connectivity and multipathing as mentioned below

#### 1. Storage connectivity packages to install
The plugin supports FC or iSCSI connectivity to the storage systems.
Install SCSI utilities (and OpeniSCSI if iscsi is needed)

  * Redhat \ SLES
  
      ```bash
      sudo yum -y install sg3_utils
      sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
      ```
  * Ubuntu
  
      ```bash
      sudo apt-get update
      sudo apt-get -y install scsitools
      sudo apt-get install -y open-iscsi  # only if you need iSCSI
      ```

#### 2. Multipathing configuration 
The plugin works only with multipath devices. Therefor make sure you configure the multipath.conf compatible to the required storage system.
     - Install and configure multipathing.
  * Redhat \ SLES
  
        ```bash
         yum install device-mapper-multipath
         sudo modprobe dm-multipath

         cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
         systemctl start multipathd
         systemctl status multipathd  # Make sure its active
         multipath -ll  # Make sure no error appear.
        ```

  * Ubuntu
  
        ```bash
         sudo apt-get multipath-tools
         cp multipath.conf /etc/multipath.conf
         multipath -l  # Check no errors appear.
        ```        

#### 3. Storage connectivity setup
  *  Verify that the hostname of the docker node is defined on the relevant storage systems with the valid WWPNs or IQN of the node.

  *  For iSCSI - Discover and login to the iSCSI targets of the relevant storage systems:
     *  Discover iSCSI targets of the storage systems portal on the host
     

         ```bash
            iscsiadm -m discoverydb -t st -p ${Storage System iSCSI Portal IP}:3260 --discover
         ```
     *  Log in to iSCSI ports. You must have at least two communication paths from your host to the storage system to achieve multipathing.
     

         ```bash
            iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login
         ```
            
## Configuring Ubiquity Docker Plugin with Ubiquity and SCBE

Create this configuration file `/etc/ubiquity/ubiquity-client.conf` as mentioned below (there is nothing special for setup the SCBE as the backend of the plugin):
 
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
