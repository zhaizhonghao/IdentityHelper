# Registrar Help
Description:
This tool has following functions with user-friendly interface:

(1) help user config the docker-compose.yaml. The parameters can be set in docker-compsose.yaml are:
* username and password for initailizing the CA. The admin will use them to enroll and get its identity.
* service name (such as ca_org1)
* fabric_ca_server_ca_name (consistent with the hostname)
* fabric_ca_server_TLS_enabled (such as true)
* fabric_ca_server_port (such as 7054)
* fabric_ca_home (the path to store the material for fabric-ca-server)
* container_name
* hostname (the hostname is consistent with the container name, such as ca.org1.example.com) 

(2) start the fabric-ca container

(3) check the status of the container

(4) shut down the container

## Fabric-ca-server的Docker容器的配置流程(纯手动)
步骤1：配置docker文件

步骤2：启动docker，在挂在的目录下生成fabric-ca-server必要的身份信息和fabric-ca-server-config.yaml文件

步骤3：修改fabric-ca-server-config.yaml文件

步骤4：删除之前生成的所有有关fabric-ca-server的身份信息

步骤5：重新启动docker,在挂在的目录下重新生成fabric-ca-server必要的身份信息