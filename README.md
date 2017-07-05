# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity storage volume service for the creation and management of volumes in storage system.  Once created, a volume can be used by Docker. 

The plugin must be installed on each of your Docker hosts. The plugin must be configured to operate with the Ubiquity storage service.

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.


## Installing the Ubiquity Docker volume plugin

### 1. Prerequisites
  * Ubiquity Docker volume plugin is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+
  * The following sudoers configuration `/etc/sudoers` is required to run the plugin as root user: 
  
     ```
        Defaults !requiretty
     ```
     For non-root users, such as USER, configure the sudoers as follows: 

     ```
         USER ALL= NOPASSWD: /usr/bin/, /bin/
         Defaults:%USER !requiretty
         Defaults:%USER secure_path = /sbin:/bin:/usr/sbin:/usr/bin
     ```
  * Verify that the pluginsDirectory, specified in ubiquity-client.conf file, exists on the host. Default localtion is /etc/docker/plugins/.

  * Ubiquity needs access to the management of the required storage backends. See [Available Storage Systems](supportedStorage.md) for connectivity details.

  
```bash
        mkdir /etc/docker/plugins
 ```

### 2. Downloading and installing the plugin

* Download and unpack the application package.
```bash
mkdir -p /etc/ubiquity
cd /etc/ubiquity
curl https://github.com/IBM/ubiquity-docker-plugin/releases/download/v0.3.0/ubiquity-docker-plugin-0.3.0.tar.gz | tar xf -
chmod u+x ubiquity-docker-plugin
#chown USER:GROUP ubiquity  # Run this command only a non-root should run ubiquity (fill up the USER and GROUP)
cp ubiquity-docker-plugin /usr/bin                         
cp ubiquity-docker-plugin.service /usr/lib/systemd/system/ 
```
   * To run the plugin as non-root users, you must add to the `/usr/lib/systemd/system/ubiquity-docker-plugin.service` file this line `User=USER` under the [Service] item.
   
   * Enable the plugin service
   
```bash 
systemctl enable ubiquity-docker-plugin.service      
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
Examples of how to manage Ubiquiy Docker volumes, such as volume create\remove\list and start\stop stateful containers, details in [Available Storage Systems](supportedStorage.md).

## Troubleshooting
### Communication Error
If any of docker volume management commands responds with following errors message, it is highly likely that ubiquity-docker-plugin and ubiquity service are not able to communicate
with each other. Please check the storageApiURL specified while starting the plugin
```bash
Error looking up volume plugin ubiquity: Plugin does not implement the requested driver
```

### Updating a Docker volume
Currently there is no way to update the options set on a volume through the Docker CLI.  In order to change its name or features, the native storage system APIs must be used. If a name must be changed and the data must be kept, it can always be deleted from Ubiquity (assuming forceDelete = false on the server) and then re-added with the new name.

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
