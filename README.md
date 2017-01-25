# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity Volume Service for the creation and management of volumes in the storage system.  Once created, a volume can be used by either Kubernetes or Docker. 

This plugin is a REST service that must be running on each of your Docker hosts (or Docker Swarm hosts).

This plugin can support a variety of storage systems.  See 'Available Storage Systems' for more details.


## Installation

### General Prerequisites
* [Ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service must be running
* Install [docker](https://docs.docker.com/engine/installation/)
* Install [golang](https://golang.org/)

### Storage Prerequisites
The correct storage software must be installed and configured on each of the hosts.

For example:
* Spectrum-Scale - Ensure the Spectrum Scale client (NSD client) is installed and part of a Spectrum Scale cluster.
* NFS - Ensure hosts support mounting NFS file systems.


### Getting started
- Configure go - GOPATH environment variable needs to be correctly set before starting the build process. Create a new directory and set it as GOPATH 
```bash
mkdir -p $HOME/workspace
export GOPATH=$HOME/workspace
```
- Configure ssh-keys for github.ibm.com - go tools require password less ssh access to github. If you have not already setup ssh keys for your github.ibm profile, please follow steps in 
(https://help.github.com/enterprise/2.7/user/articles/generating-an-ssh-key/) before proceeding further. 
- Build Ubiquity docker plugin from source (can take several minutes based on connectivity)
```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-docker-plugin.git
cd ubiquity-docker-plugin
./scripts/build

```

### Running the Plugin on each Host
On each host, you need to start the plugin as follow,

```bash
./bin/ubiquity-docker-plugin [-configFile <configFile>]
```
where:
* configFile: Configuration file to use (defaults to `./ubiquity-client.conf`)

### Configuring the Plugin

Unless otherwise specified by the `configFile` command line parameter, the Ubiquity Docker Plugin will
look for a file named `ubiquity-client.conf` for its configuration.

The following snippet shows a sample configuration file:

```toml
logPath = "/tmp"            # The Ubiquity Docker Plugin will write logs to file "ubiquity-docker-plugin.log" in this path.
backend = "spectrum-scale"  # The storage backend to use. Valid values include "spectrum-scale-nfs" and "spectrum-scale". 

[DockerPlugin]
address = "0.0.0.0"                        # IP address of plugin
port = 9000                                # Port to serve docker plugin functions
pluginsDirectory = "/etc/docker/plugins/"  # Docker plugin directory (create if not already created by Docker)


[UbiquityServer]
address = "UbiquityServiceHostname"  # IP/hostname of the Ubiquity Service
port = 9999            # TCP port on which the Ubiquity Service is listening

[SpectrumNfsRemoteConfig]  # Only relevant for use with "spectrum-scale-nfs" backend.
CIDR = "192.168.1.0/24"    # This is the subnet mask to which the NFS volumes will be exported.  Access to created volumes will be limited to this subnet.
```

After the plugin has been started, restart the Docker engine daemon so it can discover the Ubiquity Docker Plugin:
```bash
service docker restart
```

#### Common errors
If any of docker volume management commands responds with following errors message, it is highly likely that ubiquity-docker-plugin and ubiquity service are not able to communicate
with each other. Please check the storageApiURL specified while starting the plugin
```bash
Error response from daemon: create fdsfdsf: create fdsfdsf: Error looking up volume plugin spectrum-scale: Plugin does not implement the requested driver
```

### Supported Volume Types

The volume driver supports creation of three types of volumes in Spectrum Scale:

***1. Fileset Volume***

Fileset Volume is a volume which maps to a fileset in Spectrum Scale. Fileset Volume is the default type of creating
a volume.

***2. Lightweight Volume***

Lightweight Volume is a volume which maps to a sub-directory within an existing fileset in Spectrum Scale.

***3. Fileset With Quota Volume***

Fileset with Quota Volume is a volume which maps to a fileset, along with quota limit set on it, in Spectrum Scale.<br/>
Quota, especially fileset based quota, must be enabled on the file system.

### Usage

***_Example:_***

##### Creating Fileset Volumes

Create a fileset volume named demo1,  using volume driver, on the gold GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo1 --opt filesystem=gold
```

Alternatively, we can create the same volume demo1 by also passing a type option :

```bash
docker volume create -d spectrum-scale --name demo1 --opt type=fileset --opt filesystem=gold
```

Similarly, to create a fileset volume named demo2, using volume driver, on the silver GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo2 --opt filesystem=silver
```

Create a fileset volume named demo3, using volume driver, on the default existing GPFS filesystem :

```bash
docker volume create -d spectrum-scale --name demo3
```

Create a fileset volume named demo4, using volume driver and an existing fileset modelingData, on the gold GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo4 --opt fileset=modelingData --opt filesystem=gold
```

Alternatively, we can create the same volume named demo4 by also passing a type option :

```bash
docker volume create -d spectrum-scale --name demo4 --opt type=fileset --opt fileset=modelingData --opt filesystem=gold
```

##### Creating Lightweight Volumes

Create a lightweight volume named demo5, using volume driver, within an existing fileset 'LtWtVolFileset' in the gold GPFS filesystem :

```bash
docker volume create -d spectrum-scale --name demo5 --opt type=lightweight --opt fileset=LtWtVolFileset --opt filesystem=gold
```

Create a lightweight volume named demo6, using volume driver, within an existing fileset 'LtWtVolFileset' having a sub-directory 'dir1' in the gold GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo6 --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold
```

Alternatively, we can create the same volume named demo6 by also passing a type option :

```bash
docker volume create -d spectrum-scale --name demo6 --opt type=lightweight --opt fileset=LtWtVolFileset --opt directory=dir1 --opt filesystem=gold
```

##### Creating Fileset With Quota Volumes

Create a fileset with quota volume named demo7, using volume driver, with a quota limit of 1GB in the silver GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo7 --opt quota=1G --opt filesystem=silver
```

Alternatively, we can create the same volume named demo7 by also passing a type option :

```bash
docker volume create -d spectrum-scale --name demo7 --opt type=fileset --opt quota=1G --opt filesystem=silver
```

Create a fileset with quota volume named demo8, using volume driver and an existing fileset 'filesetQuota' having a quota limit of 1G, in the silver GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo8 --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver
```

Alternatively, we can also create the same volume named demo8 by also passing a type option :

```bash
docker volume create -d spectrum-scale --name demo8 --opt type=fileset --opt fileset=filesetQuota --opt quota=1G --opt filesystem=silver
```

#### List Docker Volumes

We can list the volumes created using the ubiquity docker plugin and the output should be as given below :
It lists volumes across all the volume plugins running on that system. Each volume created is listed along with the the volume driver used to create it

```bash
 $ docker volume ls
DRIVER                  VOLUME NAME
spectrum-scale          demo1
spectrum-scale          demo2
```

#### Running Docker Containers

Run a container and mount the volume created above by specifying the name of the volume name and the volume driver used to create that volume.

```bash
docker run -t -i --volume-driver spectrum-scale --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```

Similarly, if the volume was created using the spectrum-scale-nfs backend, the same command should read

```bash
docker run -t -i --volume-driver spectrum-scale-nfs --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```

**_Example_**

let's run a docker image of Alpine Linux, a lightweight Linux Distribution, inside a container and mounting demo1 volume inside the container.

```bash
docker run -t -i --volume-driver spectrum-scale --volume demo1:/data --entrypoint /bin/sh alpine
```
Here demo1 was created using the volume driver spectrum-scale, a volume plugin for the gold GPFS file system. We specify that volume demo1 must be mounted at /data inside the container

#### Removing volume
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
