# Spectrum Scale Volume Plugin
Spectrum Scale volume plugin provides access to persistent storage, utilizing Spectrum Scale, within Docker containers

# Installation
#### Prerequisites
* Provision a system running GPFS client or NSD server, but preferably running a GPFS client
* Install [docker](https://docs.docker.com/engine/installation/) 
* Install [golang](https://golang.org/)


#### *go get* the repository
Assuming you have a working installation of *golang* and the GOPATH is set correctly:

```bash
go get github.ibm.com/almaden-containers/spectrum-container-plugin.git
```

#### Creating the executables
Assuming the GOPATH and GOBIN variables are set correctly:

```bash
cd $GOPATH/github.ibm.com/almaden-containers/spectrum-container-plugin.git
go install main.go
```


# Usage
#### Run the Spectrum Scale Volume Plugin
Running docker with the spectrum scale volume plugin.
Instantiate a spectrum-container-plugin server for each GPFS filesystem that you wish to use to create docker volumes. Each instance of server must be listening on separate ports.
Assuming that the GOBIN directory is on the PATH:
```bash
main -listenAddr 127.0.0.1 -listenPort <PORT> -pluginsDirectory /etc/docker/plugins -filesystem <GPFS-FILESYSTEM-NAME> -mountpath <GPFS-FILESYSTEM-MOUNTPOINT>
```

***_Example:_***

For a GPFS client having 3 GPFS file systems(gold, silver and bronze) mounted as shown below, we run instance of spectrum-container-plugin server to handle each GPFS file system

```bash
df -Th | grep gpfs
/dev/gold      gpfs      140G  789M  139G   1% /gpfs/gold
/dev/silver    gpfs      8.0G  457M  7.6G   6% /gpfs/silver
/dev/bronze    gpfs      8.0G  457M  7.6G   6% /gpfs/bronze
```
**_Run the server for each GPFS file system_**
```bash
main -listenAddr 127.0.0.1 -listenPort 9001 -pluginsDirectory /etc/docker/plugins -filesystem gold -mountpath /gpfs/gold

main -listenAddr 127.0.0.1 -listenPort 9002 -pluginsDirectory /etc/docker/plugins -filesystem silver -mountpath /gpfs/silver

main -listenAddr 127.0.0.1 -listenPort 9003 -pluginsDirectory /etc/docker/plugins -filesystem bronze -mountpath /gpfs/bronze
```
#### Restart Docker Engine
Restart the docker engine daemon so that it can discover the plugins in the plugin directory (/etc/docker/plugins)

```bash
service docker restart
```

#### Create Docker Volumes
Create docker volumes using the volume plugins as the volume driver.

```bash 
docker volume create -d spectrum-scale-<GPFS-FILESYSTEM-NAME> --name <DOCKER-VOLUME-NAME>
```
**NOTE: The docker volume name must be unique across all the volume drivers**

**_Example_**

Create a volume named demo1 using volume driver for the gold GPFS file system :
 
 ```bash
docker volume create -d spectrum-scale-gold --name demo1
```
Similarly, to create a volume named demo2 using volume driver for the silver GPFS file system :

```bash
 docker volume create -d spectrum-scale-silver --name demo2
```

#### List Docker Volumes

We can list the volumes created using the spectrum scale plugin and the output should be as given below :
It lists volumes across all the volume plugins running on that system. Each volume created is listed along with the the volume driver used to create it

```bash
 $ docker volume ls
DRIVER                  VOLUME NAME
spectrum-scale-gold     demo1
spectrum-scale-silver   demo2
```
   
#### Running Docker Containers

Run a container and mount the volume created above by specifiying the name of the volume name and the volume driver used to create that volume.

```bash
docker run -t -i --volume-driver spectrum-scale-<GPFS-FILESYSTEM-NAME> --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```
**_Example_**

let's run a docker image of Alpine Linux, a lightweight Linux Distribution, inside a container and mounting demo1 volume inside the container. 

```bash       
docker run -t -i --volume-driver spectrum-scale-gold --volume demo1:/data --entrypoint /bin/sh alpine
```
Here demo1 was created using the volume driver spectrum-scale-gold , a volume plugin for the gold GPFS file system. We specify that volume demo1 must be mounted at /data inside the container

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

#### Running Unit Tests

```bash
./scripts/run_unit.sh
```

#### Running Integration Tests

```bash
./scripts/run_integration.sh
```

# Dependency Management

For dependency management, spectrum-container-plugin uses [godep](https://github.com/tools/godep).

#### Get *godep*

```bash
go get github.com/tools/godep
```
Assuming the GOPATH, GOBIN, and PATH variables are set correctly, run godep:
```bash
godep version
```
**Make sure you have version `godep v74`** or greater.

#### Modifying the code of a dependency

For example, for modifying spectrum-common and make spectrum-container-plugin use the modified revision:

1. Get the latest spectrum-common into your *go workspace* (get the tip of the *master* branch):<br />
`go get github.ibm.com/almaden-containers/spectrum-common.git` (use the `-u` flag if already present in workspace)
2. Optional: Checkout desired branch<br />
e.g. `cd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-common.git && git checkout develop`
3. Modify spectrum-common, commit changes (note your commit SHA), push to authoritative remote repo.<br />
e.g. `cd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-common.git && vim README.md && git add . && git commit && git push origin`
4. `cd` back to spectrum-container-plugin and run *godep update*<br />
e.g. `cd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git && godep update github.ibm.com/almaden-containers/...`<br />
Wildcard use (*`...`*) is **encouraged** as godep sometimes seems to have problems with full import path specifications of dependencies.
5. Check with `git diff` if the correct revision (i.e., the commit SHA from Step 3) has been put into the `Godeps/Godeps.json` file and if the changes have been correctly propagated to the `vendor` folder (`$GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git/vendor`).
6. Commit the updated dependencies.<br />
e.g. `git add Godeps/Godeps.json && git add vendor/ && git commit -m "Updated spectrum-common dependency" && git push origin`

#### Adding a new dependency

1. `go get github.com/sample/sample` to get the library you want to add into your go workspace.
2. Reference the library in your go code, i.e., use it.
3. In the root folder of your repository, run `godep save` *without parameters*.<br />
e.g., `cd $GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git && godep save`
4. Check with `git diff` if the new dependency has been put into the `Godeps/Godeps.json` file and if the changes have been correctly propagated to the `vendor` folder (`$GOPATH/src/github.ibm.com/almaden-containers/spectrum-container-plugin.git/vendor`).
5. Commit the updated dependency info.<br />
e.g. `git add Godeps/Godeps.json && git add vendor/ && git commit -m "Added sample dependency" && git push origin`
