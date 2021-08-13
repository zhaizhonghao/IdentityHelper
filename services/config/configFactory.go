package config

import (
	"fmt"
	"io"
	"text/template"
)

type ConfigInfo struct {
	CertName string `json:"CertName"`
}

//GenerateConfigTxTemplate To write the config to the w according to the template tpl
func GenerateDockerComposeTemplate(configInfo ConfigInfo, tpl *template.Template, w io.Writer) error {
	err := tpl.Execute(w, configInfo)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
