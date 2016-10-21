#Ubiquity Docker Volume Plugin
Ubiquity volume plugin provides access to persistent storage for Docker containers utilizing Ubiquity service integrated with Spectrum Scale 

### Installation
#### Prerequisites
* Install and start Ubiquity Service
* Install Spectrum-Scale client and connect to Spectrum-Scale cluster
* Install [docker](https://docs.docker.com/engine/installation/)
* Install [golang](https://golang.org/)


#### Build Ubiquity docker plugin from source
Assuming you have a working installation of *golang* and the GOPATH is set correctly:
```bash
mkdir -p $GOPATH/src/github.ibm.com/almaden-containers
cd $GOPATH/src/github.ibm.com/almaden-containers
git clone git@github.ibm.com:almaden-containers/ubiquity-docker-plugin.git
cd ubiquity-docker-plugin.git
./bin/build
```

This will create an out folder and build ubiquity-docker-plugin in it.

#### Running the plugin
```bash
./out/ubiquity-docker-plugin -listenAddr 127.0.0.1 -listenPort <PORT> -pluginsDirectory /etc/docker/plugins -storageApiURL "http://<ip for ubiquity service>:8999/ubiquity_storage" -logPath <>
```
Restart the docker engine daemon so that it can discover the plugins in the plugin directory (/etc/docker/plugins)
```bash
service docker restart
```

### Usage

***_Example:_***

#### Create Docker Volumes
Create docker volumes using the volume plugins as the volume driver.

```bash
docker volume create -d spectrum-scale --name <DOCKER-VOLUME-NAME>
```
**NOTE: The docker volume name must be unique across all the volume drivers**

**_Example_**

Create a volume named demo1 using volume driver for the gold GPFS file system :

 ```bash
docker volume create -d spectrum-scale --name demo1 --opt filesystem=gold
```
Similarly, to create a volume named demo2 using volume driver for the silver GPFS file system :

```bash
 docker volume create -d spectrum-scale --name demo2 --opt filesystem=silver
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
