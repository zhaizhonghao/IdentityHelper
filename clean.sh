docker stop ca.org1.example.com
docker rm ca.org1.example.com

docker stop ca.org2.example.com
docker rm ca.org2.example.com

docker stop ca.orderer.example.com
docker rm ca.orderer.example.com

docker stop ca.tls.example.com
docker rm ca.tls.example.com


sudo rm -rf crypto-config-ca/ fabric-ca/