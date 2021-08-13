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

## 组织成员的注册与登录流程

步骤1：组织管理员进行登录

a) fabric-ca-client将在本地为其生成公钥和私钥

b) 管理员利用fabric-ca-server初始化时的用户和密码登录（发送用户名、密码、生成的公钥和CSR给fabric-ca-server）

c) 如果用户名和密码正确，fabric-ca-server会为其颁发证书并返回

步骤2：组织管理元再次进行登录获取TLS的身份信息，admin的TLS获取完成之后，会补全组织的MSP目录（包括组织信任的CA有哪些，组织信任的TLS ca有哪些？）

步骤3：组织管理为组织成员进行注册，组织成员的类型包括peer和client

a) 为peer类型的组织成员(peer0,peer1)注册身份

b) 该peer(peer0,peer1)利用刚才注册的用户名和密码获取身份信息

c) 该peer(peer0,peer1)利用刚才注册的用户名和密码获取TLS身份信息

d) 为client类型的组织成员注册身份

e) 该client类型的成员利用刚才注册的用户名和密码获取身份信息

f) 该client类型的成员利用刚才注册的用户名和密码获取TLS身份信息

## 为多个组织配置CA（为每个组织CA重复下面的步骤）

步骤1：配置docker文件

步骤2：启动docker，在挂在的目录下生成fabric-ca-server必要的身份信息和fabric-ca-server-config.yaml文件

步骤3：修改fabric-ca-server-config.yaml文件

步骤4：删除之前生成的所有有关fabric-ca-server的身份信息

步骤5：重新启动docker,在挂在的目录下重新生成fabric-ca-server必要的身份信息

结束！

## 为每个组织的成员注册并登录获取身份信息

步骤1：为组织1的成员注册和登录

步骤2：为组织2的成员注册和登录

步骤3：为组织orderer的成员注册和登录

步骤4：为每个组织的成员注册和登录

结束！

