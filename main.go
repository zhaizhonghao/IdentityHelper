package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/zhaizhonghao/registrarHelper/services/config"
	"github.com/zhaizhonghao/registrarHelper/services/dockerCompose"
)

type Success struct {
	Payload string `json:"Payload"`
	Message string `json:"Message"`
}

var tpl *template.Template

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/fabricca/config", configContainer).Methods("POST", http.MethodOptions)
	router.HandleFunc("/fabricca/start", startContainer).Methods("POST", http.MethodOptions)
	router.HandleFunc("/fabricca/status", checkContainer).Methods("GET", http.MethodOptions)
	router.HandleFunc("/fabricca/stop", stopContainer).Methods("POST", http.MethodOptions)
	router.HandleFunc("/fabricca/enroll", enroll).Methods("POST", http.MethodOptions)
	router.HandleFunc("/fabricca/register", register).Methods("POST", http.MethodOptions)
	router.HandleFunc("/fabricca/enrollForTLS", enrollForTLS).Methods("POST", http.MethodOptions)

	router.Use(mux.CORSMethodMiddleware(router))

	fmt.Println("Registrar helper is listenning on localhost:8383")

	http.ListenAndServe(":8383", router)

}

func configContainer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Configing the docker-compose.yaml")
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}

	var containerConfiguration = dockerCompose.ContainerConfiguration{}
	err := json.NewDecoder(r.Body).Decode(&containerConfiguration)
	if err != nil {
		fmt.Println("failed to decode", err)
	}

	//Generate docker-compose.yaml
	tpl = template.Must(template.ParseGlob("templates/dockerCompose/*.yaml"))
	file, err := os.Create("docker-compose.yaml")
	if err != nil {
		fmt.Println("Fail to create file!")
	}
	defer file.Close()
	err = dockerCompose.GenerateDockerComposeTemplate(containerConfiguration, tpl, file)
	if err != nil {
		fmt.Println("Fail to generate docker-compose.yaml", err)
	}

	//读取文件,并返回内容
	content, err := ioutil.ReadFile("docker-compose.yaml")
	if err != nil {
		fmt.Println("fail to read file", err)
		return
	} else {
		success := Success{
			Payload: string(content),
			Message: "200 OK",
		}
		json.NewEncoder(w).Encode(success)
		return
	}
}

func startContainer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("starting the CA")
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	//检查是否有docker-compose.yaml文件
	if !isFileExist("docker-compose.yaml") {
		fmt.Println("docker-compose.yaml doesn't exist!")
		json.NewEncoder(w).Encode("docker-compose.yaml doesn't exist!")
		return
	}
	//Creating or starting the docker containers
	err, wout := runCMD("docker-compose", "up", "-d")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	success := Success{
		Payload: string(wout.Bytes()),
		Message: "200 OK",
	}
	json.NewEncoder(w).Encode(success)
	return
}

type NodeState struct {
	Name  string `json:"name"`
	State bool   `json:"state"`
}

func checkContainer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("check the status of the CA container")
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	//检查已经已有的docker容器的状态
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	nodes := []NodeState{}

	for _, container := range containers {
		node := NodeState{}
		node.Name = strings.Split(container.Names[0], "/")[1]
		if container.State == "running" {
			node.State = true
		} else {
			node.State = false
		}
		nodes = append(nodes, node)
	}
	fmt.Println(nodes)
	//如果有容器在运行，则返回所有正在运行的容器名称和状态；如果没有容器运行，则返回空。
	json.NewEncoder(w).Encode(nodes)
	return
}

type ContainerInfo struct {
	Name string `json:"Name"`
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
	fmt.Println("check the status of the CA container")
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	var containerInfo = ContainerInfo{}
	err := json.NewDecoder(r.Body).Decode(&containerInfo)
	if err != nil {
		fmt.Println("failed to decode", err)
		return
	}
	//stop the docker container
	err, wout := runCMD("docker", "stop", containerInfo.Name)
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	success := Success{
		Payload: "shutdown " + containerInfo.Name + " the container successfully!",
		Message: "200 OK",
	}
	json.NewEncoder(w).Encode(success)
	return
}

//MSPDir 一定要是绝对路径
//TODO CSRHosts实际上是一个list
type EnrollmentInfo struct {
	Username         string `json:"Username"`
	Password         string `json:"Password"`
	Address          string `json:"Address"`
	CAName           string `json:"CAName"`
	CommonName       string `json:"CommonName"`
	Country          string `json:"Country"`
	State            string `json:"State"`
	Organization     string `json:"Organization"`
	OrganizationUnit string `json:"OrganizationUnit"`
	MSPDir           string `json:"MSPDir"`
	CSRHosts         string `json:"CSRHosts"`
	PathOfCATLSCert  string `json:"PathOfCATLSCert"`
}

//fabric-ca-client enroll -u https://admin:adminpw@localhost:7054 --caname ca.org1.example.com --mspdir ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/msp --csr.hosts Admin@org1.example.com --tls.certfiles ${PWD}/fabric-ca/org1/tls-cert.pem
//fabric-ca-client enroll -u https://peer0:peer0pw@localhost:7054 --caname ca.org1.example.com --mspdir ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/msp --csr.hosts peer0.org1.example.com --tls.certfiles ${PWD}/fabric-ca/org1/tls-cert.pem
func enroll(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	pwd := os.Getenv("PWD")
	fmt.Println("current path", pwd)
	var enrollmentInfo = EnrollmentInfo{}
	err := json.NewDecoder(r.Body).Decode(&enrollmentInfo)
	if err != nil {
		fmt.Println("failed to decode", err)
		return
	}
	//变成绝对路径
	enrollmentInfo.MSPDir = pwd + enrollmentInfo.MSPDir
	enrollmentInfo.PathOfCATLSCert = pwd + enrollmentInfo.PathOfCATLSCert

	url := fmt.Sprintf("https://%s:%s@%s",
		enrollmentInfo.Username,
		enrollmentInfo.Password,
		enrollmentInfo.Address)
	fmt.Println(enrollmentInfo.Username, "enrolling")
	//enroll
	//TODO这里的localhost不应该写死
	err, wout := runCMD("fabric-ca-client", "enroll",
		"--url", url,
		"--caname", enrollmentInfo.CAName,
		"--csr.hosts", enrollmentInfo.CSRHosts,
		"--csr.hosts", "localhost",
		"--csr.names", "C="+enrollmentInfo.Country,
		"--csr.names", "ST="+enrollmentInfo.State,
		"--csr.names", "O="+enrollmentInfo.Organization,
		"--csr.names", "OU="+enrollmentInfo.OrganizationUnit,
		"--mspdir", enrollmentInfo.MSPDir,
		"--tls.certfiles", enrollmentInfo.PathOfCATLSCert)
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}
	address := parseAddress(enrollmentInfo.Address)
	caName := parseCAName(enrollmentInfo.CAName)
	//补全MSP(在msp中放入config.yaml文件，下面是其内容)
	tpl = template.Must(template.ParseGlob("templates/config/*.yaml"))
	file, err := os.Create(enrollmentInfo.MSPDir + "/config.yaml")
	if err != nil {
		fmt.Println("Fail to create config.yaml file!")
	}
	defer file.Close()
	configInfo := config.ConfigInfo{}
	configInfo.CertName = address + "-" + caName
	err = config.GenerateDockerComposeTemplate(configInfo, tpl, file)
	if err != nil {
		fmt.Println("Fail to generate config.yaml file!", err)
	}

	success := Success{
		Payload: "enroll successfully!",
		Message: "200 OK",
	}
	//返回结果
	json.NewEncoder(w).Encode(success)
	return
}

func parseAddress(address string) string {
	return strings.Replace(address, ":", "-", -1)
}

func parseCAName(caName string) string {
	return strings.Replace(caName, ".", "-", -1)
}

func enrollForTLS(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	pwd := os.Getenv("PWD")
	fmt.Println("current path", pwd)
	var enrollmentInfo = EnrollmentInfo{}
	err := json.NewDecoder(r.Body).Decode(&enrollmentInfo)
	if err != nil {
		fmt.Println("failed to decode", err)
		return
	}
	//变成绝对路径
	enrollmentInfo.MSPDir = pwd + enrollmentInfo.MSPDir
	enrollmentInfo.PathOfCATLSCert = pwd + enrollmentInfo.PathOfCATLSCert

	url := fmt.Sprintf("https://%s:%s@%s",
		enrollmentInfo.Username,
		enrollmentInfo.Password,
		enrollmentInfo.Address)

	fmt.Println(enrollmentInfo.Username, "enrolling for tls")
	//enroll
	//TODO这里的localhost不应该写死
	err, wout := runCMD("fabric-ca-client", "enroll",
		"--url", url,
		"--caname", enrollmentInfo.CAName,
		"--csr.hosts", enrollmentInfo.CSRHosts,
		"--csr.hosts", "localhost",
		"--csr.names", "C="+enrollmentInfo.Country,
		"--csr.names", "ST="+enrollmentInfo.State,
		"--csr.names", "O="+enrollmentInfo.Organization,
		"--csr.names", "OU="+enrollmentInfo.OrganizationUnit,
		"--enrollment.profile", "tls",
		"--mspdir", enrollmentInfo.MSPDir,
		"--tls.certfiles", enrollmentInfo.PathOfCATLSCert)
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	//补全MSP(拷贝文件)
	//cp ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/tlscacerts/* ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt
	err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/tlscacerts/* "+enrollmentInfo.MSPDir+"/ca.crt")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	//cp ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/signcerts/* ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/server.crt
	err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/signcerts/* "+enrollmentInfo.MSPDir+"/server.crt")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	//cp ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/keystore/* ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/server.key
	err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/keystore/* "+enrollmentInfo.MSPDir+"/server.key")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	//将msp的tlscacerts补全
	err, wout = runCMD("mkdir", enrollmentInfo.MSPDir+"/../msp/tlscacerts")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}
	//将msp的admincerts补全
	err, wout = runCMD("mkdir", enrollmentInfo.MSPDir+"/../msp/admincerts")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}
	//cp ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/tlscacerts/* ${PWD}/crypto-config-ca/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/msp/tlscacerts/tlscacert.pem
	err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/tlscacerts/* "+enrollmentInfo.MSPDir+"/../msp/tlscacerts/tlscacert.pem")
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	//如果是Admin把组织的msp给补全了

	if strings.Contains(enrollmentInfo.Username, "admin") {
		//创建msp目录
		err, wout = runCMD("mkdir", "-p", enrollmentInfo.MSPDir+"/../../../ca")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		err, wout = runCMD("mkdir", "-p", enrollmentInfo.MSPDir+"/../../../tlsca")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		err, wout = runCMD("mkdir", "-p", enrollmentInfo.MSPDir+"/../../../msp/cacerts")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		err, wout = runCMD("mkdir", "-p", enrollmentInfo.MSPDir+"/../../../msp/tlscacerts")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		//将admin的msp目录下的cacerts里面文件考到组织msp的目录中
		err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/../msp/cacerts/* "+enrollmentInfo.MSPDir+"/../../../msp/cacerts/")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/../msp/cacerts/* "+enrollmentInfo.MSPDir+"/../../../ca/")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		//将admin的msp目录下的tlscacerts里面文件考到组织msp的目录中
		err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/../msp/tlscacerts/* "+enrollmentInfo.MSPDir+"/../../../msp/tlscacerts/")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/../msp/tlscacerts/* "+enrollmentInfo.MSPDir+"/../../../tlsca/")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
		//将admin的msp目录下的config.yaml考到组织msp的目录中
		err, wout = runCMD("/bin/sh", "-c", "cp "+enrollmentInfo.MSPDir+"/../msp/config.yaml "+enrollmentInfo.MSPDir+"/../../../msp/config.yaml")
		if err != nil {
			json.NewEncoder(w).Encode(string(wout.Bytes()))
			return
		}
	}

	success := Success{
		Payload: "enroll for tls successfully!",
		Message: "200 OK",
	}

	//返回结果
	json.NewEncoder(w).Encode(success)
	return
}

type RegisterInfo struct {
	Username        string `json:"Username"`
	Password        string `json:"Password"`
	Type            string `json:"Type"`
	Address         string `json:"Address"`
	MSPDir          string `json:"MSPDir"`
	CSRHosts        string `json:"CSRHosts"`
	PathOfCATLSCert string `json:"PathOfCATLSCert"`
}

//./fabric-ca-client register --id.name org1admin --id.secret org1adminpw --url https://example.com:7054 --mspdir ./org1-ca/msp --id.type admin --tls.certfiles ../tls/tls-ca-cert.pem --csr.hosts 'host1,*.example.com'
func register(w http.ResponseWriter, r *http.Request) {
	setHeader(w)
	if (*r).Method == "OPTIONS" {
		fmt.Println("Options request discard!")
		return
	}
	pwd := os.Getenv("PWD")
	fmt.Println("current path", pwd)
	var registerInfo = RegisterInfo{}
	err := json.NewDecoder(r.Body).Decode(&registerInfo)
	if err != nil {
		fmt.Println("failed to decode", err)
		return
	}
	//转为绝对路径
	registerInfo.MSPDir = pwd + registerInfo.MSPDir
	registerInfo.PathOfCATLSCert = pwd + registerInfo.PathOfCATLSCert

	fmt.Println("registering", registerInfo.Username)
	url := fmt.Sprintf("https://%s",
		registerInfo.Address)

	err, wout := runCMD("fabric-ca-client", "register",
		"--id.name", registerInfo.Username,
		"--id.secret", registerInfo.Password,
		"--id.type", registerInfo.Type,
		"--url", url,
		"--mspdir", registerInfo.MSPDir,
		"--tls.certfiles", registerInfo.PathOfCATLSCert)
	if err != nil {
		json.NewEncoder(w).Encode(string(wout.Bytes()))
		return
	}

	success := Success{
		Payload: string(wout.Bytes()),
		Message: "200 OK",
	}
	json.NewEncoder(w).Encode(success)
	return
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "PUT,POST,GET,DELETE,OPTIONS")
	w.Header().Set("X-Powered-By", "3.2.1")
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
}

func runCMD(name string, args ...string) (error, *bytes.Buffer) {
	cmd := exec.Command(name, args...)
	arrString := strings.Join(args, "")
	fmt.Println("Executing", name, arrString)
	wout := bytes.NewBuffer(nil)
	cmd.Stderr = wout

	err := cmd.Run()
	if err != nil {
		fmt.Println("Execute Command failed:" + err.Error())
		//将错误提示输出
		fmt.Printf("Stderr: %s\n", string(wout.Bytes()))
		return err, wout
	}

	//输出执行提示
	fmt.Printf("std: %s\n", string(wout.Bytes()))
	return nil, wout
}
