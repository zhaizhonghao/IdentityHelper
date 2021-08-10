package dockerCompose

import (
	"fmt"
	"io"
	"text/template"
)

type ContainerConfiguration struct {
	Username                 string `json:"Username"`
	Password                 string `json:"Password"`
	ContainerName            string `json:"ContainerName"`
	FabricCAServerTLSEnabled bool   `json:"FabricCAServerTLSEnabled"`
	FabricCAServerPort       string `json:"FabricCAServerPort"`
	FabricCAHome             string `json:"FabricCAHome"`
	HostName                 string `json:"HostName"`
	CommonName               string `json:"CommonName"`
	Country                  string `json:"Country"`
	State                    string `json:"State"`
	Organization             string `json:"Organization"`
	OrganizationUnit         string `json:"OrganizationUnit"`
}

//GenerateConfigTxTemplate To write the config to the w according to the template tpl
func GenerateDockerComposeTemplate(containerConfiguration ContainerConfiguration, tpl *template.Template, w io.Writer) error {
	err := tpl.Execute(w, containerConfiguration)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
