# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity Volume Service for the creation and management of volumes in the storage system.  Once created, a volume can be used by either Kubernetes or Docker. 

This plugin is a REST service that must be running on each of your Docker hosts (or Docker Swarm hosts).

This code is provided "AS IS" and without warranty of any kind.  Any issues will be handled on a best effort basis.


## Installation

### Build Prerequisites
* Install [golang](https://golang.org/) (>=1.6)
* Install git


### Deployment Prerequisites
* [Ubiquity](https://github.com/ibm/ubiquity) service must be running
* Install [docker](https://docs.docker.com/engine/installation/)
* Install [golang](https://golang.org/) and setup your go path
* Install [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
* The correct storage software must be installed and configured on each of the hosts. For example:
  * Spectrum-Scale - Ensure the Spectrum Scale client (NSD client) is installed and part of a Spectrum Scale cluster.
  * NFS - Ensure hosts support mounting NFS file systems.
  * IBM Block Storage - Configuring storage connectivity and multipathing as mentioned below

     The plugin supports FC or iSCSI connectivity to the storage systems.
      - Install OpeniSCSI and SCSI utilities.
        * Ubuntu
        ```bash
         sudo apt-get update
         sudo apt-get -y install scsitools
         sudo apt-get install -y open-iscsi  # only if you need iSCSI
         ```
         * Redhat
         ```bash
         sudo yum -y install sg3_utils
         sudo yum -y install iscsi-initiator-utils  # only if you need iSCSI
         ```

     - Install and configure multipathing.
         * Ubuntu
        ```bash
         sudo apt-get multipath-tools
         cp multipath.conf /etc/multipath.conf
         multipath -l  # Check no errors appear.
        ```

         * Redhat
        ```bash
         yum install device-mapper-multipath
         sudo modprobe dm-multipath

         cp multipath.conf /etc/multipath.conf  # Default file can be copied from  /usr/share/doc/device-mapper-multipath-*/multipath.conf to /etc
         systemctl start multipathd
         systemctl status multipathd  # Make sure its active
         multipath -ll  # Make sure no error appear.
        ```

     - Verify that the hostname of the docker node is defined on the relevant storage systems with the valid WWPNs or IQN of the node.

     - For iSCSI - Discover and login to the iSCSI targets of the relevant storage systems:
         * Discover iSCSI targets of the storage systems portal on the host

            ```bash
               iscsiadm -m discoverydb -t st -p ${Storage System iSCSI Portal IP}:3260 --discover
            ```
         * Log in to iSCSI ports. You must have at least two communication paths from your host to the storage system to achieve multipathing.

            ```bash
               iscsiadm -m node  -p ${storage system iSCSI portal IP/hostname} --login
            ```


### Download and build the code
- Configure [golang](https://golang.org/) - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH 
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
- Configure ssh-keys for github.ibm.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github.ibm profile, please follow steps in 
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further. 
- Build Ubiquity docker plugin from source (can take several minutes based on connectivity)
```bash
mkdir -p $GOPATH/src/github.com/IBM
cd $GOPATH/src/github.com/IBM
git clone git@github.com:IBM/ubiquity-docker-plugin.git
cd ubiquity-docker-plugin
./scripts/build

```

### Configuring the Plugin

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity Docker Plugin will
look for a file named `ubiquity-client.conf` for its configuration.

The following snippet shows a sample configuration file:

```toml
logPath = "/tmp/ubiquity"            # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.

[DockerPlugin]
port = 9000                                # Port to serve docker plugin functions
pluginsDirectory = "/etc/docker/plugins/"  # Point to the location of the configured Docker plugin directory (create if not already created by Docker)


[UbiquityServer]
address = "UbiquityServiceHostname"  # IP/hostname of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening

[SpectrumNfsRemoteConfig]  # Only relevant for use with "spectrum-scale-nfs" backend.
ClientConfig = "192.0.2.0/20(Access_Type=RW,Protocols=3:4);198.51.100.0/20(Access_Type=RO,Protocols=3:4,Transports=TCP:UDP)"    # Mandatory. Declares the client specific settings for NFS volume exports. Access will be limited to the specified client subnet(s) and protocols.
```

The plugin must be started prior to the Docker daemon on the host.  Therefore, if Docker is already running, after the plugin has been started, restart the Docker engine daemon so it can discover the Ubiquity Docker Plugin:
```bash
service docker restart
```
### Two Options to Install and Run

#### Option 1: systemd

This option assumes that the system that you are using has support for systemd (e.g., ubuntu 14.04 does not have native support to systemd, ubuntu 16.04 does.)
Please note that the script will try to start the service as user `ubiquity`. So before proceeding, please create the user ubiquity as described in [Ubiquity documentation](https://github.com/IBM/ubiquity).


1) Inside the ubiquity-docker-plugin/scripts directory, execute the following command 
```bash
./setup
```

This command will copy ubiquity-docker-plugin binary to /usr/bin, ubiquity-client-docker.conf and ubiquity-docker-plugin.env  to /etc/ubiquity location. It will also enable ubiquity-docker-plugin service.

2) Make appropriate changes to /etc/ubiquity/ubiquity-client-docker.conf e.g. server ip/port, plugin directory etc.

3) Edit /etc/ubiquity/ubiquity-docker-plugin.env  to add/remove command line options to Ubiquity docker plugin

4) Start or stop the Ubiquity docker plugin service using the following command
```bash
systemctl start/stop/restart ubiquity-docker-plugin
```

#### Option 2: Manual
##### Running the Plugin on each Host
On each host, you need to start the plugin as follows.  Note that the service will stop if the shell in which it is running exits.  To run as a service, please use systemd above.

```bash
./bin/ubiquity-docker-plugin [-configFile <configFile>]
```
where:
* configFile: Configuration file to use (defaults to `./ubiquity-client.conf`)




### Common errors
#### Plugins Directory
Ensure that pluginsDirectory specified in ubiquity-client.conf file exists on the host before starting the plugin. Default localtion is /etc/docker/plugins/.

#### Communication Error
If any of docker volume management commands responds with following errors message, it is highly likely that ubiquity-docker-plugin and ubiquity service are not able to communicate
with each other. Please check the storageApiURL specified while starting the plugin
```bash
Error response from daemon: create fdsfdsf: create fdsfdsf: Error looking up volume plugin spectrum-scale: Plugin does not implement the requested driver
```


#### Sample Spectrum Scale Usage

##### Creating Fileset Volumes

Create a fileset volume named demo1,  using volume driver, on the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo1 --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume demo1 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo1 --opt type=fileset --opt filesystem=gold --opt backend=spectrum-scale
```

Similarly, to create a fileset volume named demo2, using nfs volume driver, on the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --opt backend=spectrum-scale-nfs --name demo2 --opt filesystem=silver
```

Create a fileset volume named demo3, using volume driver, on the default existing Spectrum Scale filesystem :

```bash
docker volume create -d ubiquity --name demo3 --opt backend=spectrum-scale
```

Create a fileset volume named demo4, using volume driver and an existing fileset modelingData, on the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo4 --opt fileset=modelingData --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo4 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo4 --opt type=fileset --opt fileset=modelingData --opt filesystem=gold --opt backend=spectrum-scale
```

##### Creating Independent Fileset Volumes

Create an independent fileset volume named demo5, using volume driver, on the gold Spectrum Scale file system

```bash
docker volume create -d ubiquity --name demo5 --opt type=fileset --opt filesystem=gold --opt fileset-type=independent
```

Create an independent fileset volume named demo6 having an inode limit of 1024, using volume driver, on the gold Spectrum Scale file system

```bash
docker volume create -d spectrum-scale --name demo6 --opt type=fileset --opt filesystem=gold --opt fileset-type=independent --opt inode-limit=1024
```

##### Creating Lightweight Volumes

Create a lightweight volume named demo7, using volume driver, within an existing fileset 'LtWtVolFileset' in the gold Spectrum Scale filesystem :

```bash
docker volume create -d ubiquity --name demo7 --opt type=lightweight --opt fileset=LtWtVolFileset --opt filesystem=gold --opt backend=spectrum-scale
```

Create a lightweight volume named demo8, using volume driver, within an existing fileset 'LtWtVolFileset' having a sub-directory 'dir1' in the gold Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo8 --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo8 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo8 --opt type=lightweight --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold --opt backend=spectrum-scale
```

##### Creating Fileset With Quota Volumes

Create a fileset with quota volume named demo9, using volume driver, with a quota limit of 1GB in the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo9 --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Alternatively, we can create the same volume named demo9 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo9 --opt type=fileset --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Create a fileset with quota volume named demo10, using volume driver and an existing fileset 'filesetQuota' having a quota limit of 1G, in the silver Spectrum Scale file system :

```bash
docker volume create -d ubiquity --name demo10 --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

Alternatively, we can also create the same volume named demo10 by also passing a type option :

```bash
docker volume create -d ubiquity --name demo10 --opt type=fileset --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver --opt backend=spectrum-scale
```

#### Sample IBM Block Storage Usage
[Ubiquity service](https://github.com/IBM/ubiquity) communicates with the IBM block storage systems through IBM Spectrum Control Base Edition([SCBE](http://www.ibm.com/support/knowledgecenter/STWMS9/landing/IBM_Spectrum_Control_Base_Edition_welcome_page.html)).
The plugin can provision a volume from a delegated SCBE storage service by using the --opt=<SCBE storage service name> flag.

##### Creating volume on gold SCBE storage service
Create a volume named demo11 with 10gb size from the gold SCBE storage service (the gold service could be, for example, a pool from IBM FlashSystem A9000\R and with high QoS capability) :

```bash
docker volume create -d ubiquity --name demo11 --opt size=10 --opt profile=gold
```

## General Examples
### List Docker Volumes

We can list the volumes created using the ubiquity docker plugin and the output should be as given below :
It lists volumes across all the volume plugins running on that system. Each volume created is listed along with the the volume driver used to create it

```bash
 $ docker volume ls
DRIVER                  VOLUME NAME
ubiquity                    demo1
ubiquity                    demo2
```

#### Running Docker Containers and Using Volumes.  

Run a container and mount the volume created above by specifying the name of the volume name and the volume driver used to create that volume.  Note that local and ubiquity volumes can be passed into a container.

```bash
docker run -t -i --volume-driver ubiquity --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```

Similarly, if the volume was created using the spectrum-scale-nfs backend, the same command should read

```bash
docker run -t -i --volume-driver ubiquity --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```

**_Example_**

let's run a docker image of Alpine Linux, a lightweight Linux Distribution, inside a container and mounting demo1 volume inside the container.

```bash
docker run -t -i --volume-driver ubiquity --volume demo1:/data --entrypoint /bin/sh alpine
```
Here demo1 was created using the volume driver ubiquity, a volume plugin for the gold Spectrum Scale file system. We specify that volume demo1 must be mounted at /data inside the container

### Remove a Volume
**_Pre-Conditions :_** Make sure the volume is not being used by any running containers

```bash
docker volume rm <VOLUME-NAME>
```

**_Example:_**

To Remove volume demo1, created above :
```bash
docker volume rm demo1
```

**_NOTE: If an error occurs try removing any stale docker entries by running the following command and then try removing the volume again:_**

```bash
docker rm `docker ps -aq`
```

**_NOTE: Whether data is actually removed or not is controlled by the forceDelete config option on the Ubiquity service.

### Update a Volume

Currently there is no way to update the options set on a volume through the Docker CLI.  In order to change its name or features, the native storage system APIs must be used.  For example, if the name of a Spectrum Scale fileset or directory (that maps to a lightweight volume) is change, Ubiquity will no longer recognize it.  If a name must be changed and the data must be kept, it can always be deleted from Ubiquity (assuming forceDelete = false on the server) and then re-added with the new name.

## Suggestions and Questions

For any questions, suggestions, or issues, please use github.
