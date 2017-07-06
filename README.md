# Ubiquity Docker Volume Plugin
The Ubiquity Docker volume plugin provides access to persistent storage for Docker containers.  This plugin communicates with the Ubiquity storage volume service for the creation and management of volumes in storage system.  Once created, a volume can be used by Docker. 

The plugin must be installed on each of your Docker hosts. The plugin must be configured to operate with the Ubiquity storage service.

The code is provided as is, without warranty. Any issue will be handled on a best-effort basis.


## Installing the Ubiquity Docker volume plugin
Install and configure the plugin on each node in the Docker Swarm cluster that requires access to Ubiquity volumes.

### 1. Prerequisites
  * Ubiquity Docker volume plugin is supported on the following operating systems:
    - RHEL 7+
    - SUSE 12+

  * Ubiquity Docker volume plugin requires Docker version 17+.

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

  * The Docker node must have access to the storage backends. Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.
   

  

### 2. Downloading and installing the plugin

* Download and unpack the application package.
```bash
mkdir -p /etc/ubiquity
cd /etc/ubiquity
curl https://github.com/IBM/ubiquity-docker-plugin/releases/download/v0.3.0/ubiquity-docker-plugin-0.3.0.tar.gz | tar xf -
chmod u+x ubiquity-docker-plugin
#chown USER:GROUP ubiquity       ### Run this command only a non-root user.
cp ubiquity-docker-plugin /usr/bin                         
cp ubiquity-docker-plugin.service /usr/lib/systemd/system/ 
```
   * To run the plugin as non-root user, add the `User=USER` line under the [Service] item in the  `/usr/lib/systemd/system/ubiquity-docker-plugin.service` file.
   
   * Enable the plugin service.
   
```bash 
systemctl enable ubiquity-docker-plugin.service      
```

### 3. Configuring the plugin
Before running the plugin service, you must create and configure the `/etc/ubiquity/ubiquity-client.conf` file, according to your storage system type.
Follow the configuration procedures detailed in the [Available Storage Systems](supportedStorage.md) section.


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
For examples on how to create, remove, list Ubiquity Docker volumes, as well as start and stop stateful containers, refer to the [Available Storage Systems](supportedStorage.md) section, according to your storage system type.

## Troubleshooting
### Communication failure
If the  `Error looking up volume plugin ubiquity: Plugin does not implement the requested driver` error is displayed and the `Error in activate remote call &url.Error` message is stored in the `ubiquity-docker-plugin.log` file, verify comminication link between the plugin and Ubiqutiy server nodes. The loss of communication may occur if the relevant TCP  ports are not open. The port numbers are detailed in the plugin and Ubiquity server configuration files.

## Support
For any questions, suggestions, or issues, use github.

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
