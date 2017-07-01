# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity Volume Service for the creation and management of volumes in the storage system.  Once created, a volume can be used by either Kubernetes or Docker. 

This plugin is a REST service that must be running on each of your Docker hosts (or Docker Swarm hosts).  The plugin must be configured to access the single deployed instance of the Ubiquity service.

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
* The correct storage software must be installed and configured on each of the hosts. See [Available Storage Systems](supportedStorage.md) for more information on installing and configuring storage software specific to a storage system supported by ubiquity.
  
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

For more information on configuring Plugin with Ubiquity Service and one of its supported storage system see [Available Storage Systems](supportedStorage.md)

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

#### Volume Creation
Ubiquity Docker Plugin communicates with Ubiquity Service to create volumes on one of the storage systems supported by it.  Storage system specific options can be provided using the 'opt' option on the command line as a set of key-value pairs.

For more information examples of volume creation specific to Ubiquity supported storage systems see [Available Storage Systems](supportedStorage.md)  

### List Docker Volumes

We can list the volumes created using the ubiquity docker plugin:

```bash
 $ docker volume ls
DRIVER                  VOLUME NAME
ubiquity                    demo1
ubiquity                    demo2
```

Please note that the 'volume ls' command will list all volumes across all the volume plugins (including plugins other than Ubiquity) running on the host on which the command was executed.

#### Running Docker Containers and Using Volumes.  

Run a container and mount the volume created above by specifying the name of the volume name and the volume driver used to create that volume.  Note that local and ubiquity volumes can be passed into a container.

```bash
docker run -t -i --volume-driver ubiquity --volume <VOLUME-NAME>:<CONTAINER-MOUNTPOINT> --entrypoint /bin/sh alpine
```

**_Example_**

let's run a docker image of Alpine Linux, a lightweight Linux Distribution, inside a container and mounting demo1 volume inside the container.

```bash
docker run -t -i --volume-driver ubiquity --volume demo1:/data --entrypoint /bin/sh alpine
```
Here demo1 was created using the volume driver ubiquity. We specify that volume demo1 must be mounted at /data inside the container

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

Currently there is no way to update the options set on a volume through the Docker CLI.  In order to change its name or features, the native storage system APIs must be used. If a name must be changed and the data must be kept, it can always be deleted from Ubiquity (assuming forceDelete = false on the server) and then re-added with the new name.

## Suggestions and Questions

For any questions, suggestions, or issues, please use github.
