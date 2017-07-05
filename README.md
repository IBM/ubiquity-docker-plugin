# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity storage volume service for the creation and management of volumes in storage system.  Once created, a volume can be used by Docker. 

The plugin must be installed on each of your Docker hosts. The plugin must be configured to operate with the Ubiquity storage service.

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.


## Installing the Ubiquity Docker volume plugin

### 1. Prerequisites
  * Ubiquity Docker volume plugin is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+
  * Ubiquity needs access to the management of the required storage backends. See [Available Storage Systems](supportedStorage.md) for connectivity details.
  * The following sudoers configuration is required to run the plugin process:
  
        ```bash
        Defaults !requiretty
        ```
  * Verify that the pluginsDirectory, specified in ubiquity-client.conf file, exists on the host. Default localtion is /etc/docker/plugins/.
  
        ```bash
        mkdir /etc/docker/plugins
        ```

### 2. Downloading and installing the plugin

```bash
mkdir -p /etc/ubiquity
cd /etc/ubiquity
curl https://github.com/IBM/ubiquity-docker-plugin/releases/download/v0.3.0/ubiquity-docker-plugin-0.3.0.tar.gz | tar xf -
chmod u+x ubiquity-docker-plugin
cp ubiquity-docker-plugin /usr/bin                         # Copy the plugin binary file
cp ubiquity-docker-plugin.service /usr/lib/systemd/system/ # Copy the plugin systemd config to systemd directory
systemctl enable ubiquity-docker-plugin.service            # Enable plugin systemd service
```

### 3. Configuring the plugin
Configure plugin according to your storage backend requirements. Refer to 
[specific instructions](supportedStorage.md) for specific configuration needed by the storage backend. 
The configuration file must be named `ubiquity-client.conf` and placed in `/etc/ubiquity` directory.


### 4. Running the plugin service
  * Run the service.
```bash
systemctl start ubiquity-docker-plugin    
```
  * Restart the Docker engine daemon on the host to let it discover the new plugin. 
```bash
service docker restart
```

## Plugin usage examples
### Creating a volume
Ubiquity Docker Plugin communicates with Ubiquity Service to create volumes on one of the storage systems that was configured. Storage system specific options can be provided using the '--opt' option on the command line as a set of key-value pairs.

In the Creation For more information examples of volume creation specific to Ubiquity supported storage systems see [Available Storage Systems](supportedStorage.md)  

### Listing Docker volumes

We can list the volumes created using the ubiquity docker plugin:

```bash
 $ docker volume ls
DRIVER                  VOLUME NAME
ubiquity                    demo1
ubiquity                    demo2
```

Please note that the 'volume ls' command will list all volumes across all the volume plugins (including plugins other than Ubiquity) running on the host on which the command was executed.

### Running Docker containers using the Ubiquity volumes

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

### Removing a Docker volume
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

### Updating a Docker volume
Currently there is no way to update the options set on a volume through the Docker CLI.  In order to change its name or features, the native storage system APIs must be used. If a name must be changed and the data must be kept, it can always be deleted from Ubiquity (assuming forceDelete = false on the server) and then re-added with the new name.

### Troubleshooting
#### Communication Error
If any of docker volume management commands responds with following errors message, it is highly likely that ubiquity-docker-plugin and ubiquity service are not able to communicate
with each other. Please check the storageApiURL specified while starting the plugin
```bash
Error looking up volume plugin ubiquity: Plugin does not implement the requested driver
```

## Suggestions and Questions

For any questions, suggestions, or issues, please use github.

## Licensing

Copyright 2016, 2017 IBM Corp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
