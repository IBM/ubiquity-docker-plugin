#Ubiquity Docker Volume Plugin
Ubiquity volume plugin provides access to persistent storage for Docker containers utilizing Ubiquity service integrated with Spectrum Scale 

### Installation
#### Prerequisites
* [Ubiquity](https://github.ibm.com/almaden-containers/ubiquity) service must be running
* Install Spectrum-Scale client and connect to Spectrum-Scale cluster
* Install [docker](https://docs.docker.com/engine/installation/)
* Install [golang](https://golang.org/)


#### Getting started
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

#### Running the plugin
```bash
./out/ubiquity-docker-plugin -listenAddr <> -listenPort <> -pluginsDirectory <> -ubiquityServerIP <> -ubiquityServerPort <> -logPath <>
```
where:
* listenAddr: IP address of plugin (preferably 127.0.0.1) 
* listenPort: Port to serve docker plugin functions
* pluginsDirectory: Directory where docker looks for plugins (please create if not already created by docker)
* logPath: Path to create ubiquity-docker-plugin log file
* ubiquityServerIP: IP address where ubiquity server is running
* ubiquityServerPort: Port where ubiquity server is listening
Examples invocation of binary:
```bash
./out/ubiquity-docker-plugin -listenAddr 127.0.0.1 -listenPort 9000 -pluginsDirectory /etc/docker/plugins -ubiquityServerIP 127.0.0.1 -ubiquityServerPort 8999 -logPath /tmp
```



Restart the docker engine daemon so that it can discover the plugins in the plugin directory (/etc/docker/plugins)
```bash
service docker restart
```

#### Common errors
If any of docker volume management commands responds with following errors message, it is highly likely that ubiquity-docker-plugin and ubiquity service are not able to communicate 
with each other. Please check the storageApiURL specified while starting the plugin
```bash
Error response from daemon: create fdsfdsf: create fdsfdsf: Error looking up volume plugin spectrum-scale: Plugin does not implement the requested driver
```


#### Supported Volume Types

The volume driver supports creation of three types of volumes:

***1. Fileset Volume***

Fileset Volume is a volume which maps to a fileset in spectrum scale. Fileset Volume is the default type of creating
a volume 
 
***2. Lightweight Volume***

Lightweight Volume is a volume which maps to a sub-directory within an existing fileset in spectrum scale.

***3. Fileset With Quota Volume***

Fileset with Quota Volume is a volume which maps to a fileset, along with quota limit set on it, in spectrum scale.<br/>
Quota, especially fileset based quota, must be enabled on the file system


Create docker volumes using the volume plugins as the volume driver.

```bash
docker volume create -d spectrum-scale --name <DOCKER-VOLUME-NAME>
```
**NOTE: The docker volume name must be unique across all the volume drivers**

### Usage

***_Example:_***

Create a fileset volume named demo1,  using volume driver, on the gold GPFS file system :

```bash
docker volume create -d spectrum-scale --name demo1 --opt filesystem=gold
```

Alternatively, we can create the same volume demo1 by also passing a type option

```bash
docker volume create -d spectrum-scale --name demo1 --opt type=fileset --opt filesystem=gold
```

Similarly, to create a fileset volume named demo2, using volume driver, on the silver GPFS file system :

```bash
 docker volume create -d spectrum-scale --name demo2 --opt filesystem=silver
```

Create a lightweight volume named demo3, using volume driver, within an existing fileset 'LtWtVolFileset' in the gold GPFS filesystem :

```bash
docker volume create -d spectrum-scale --name demo3 --opt type=lightweight --opt fileset=LtWtVolFileset --opt filesystem=gold
```

Create a fileset with quota volume named demo4, using volume driver, with a quota limit of 1GB in the silver file system:

```bash
docker volume create -d spectrum-scale --name demo4 --opt type=fileset --opt quota=1G --opt filesystem=silver
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