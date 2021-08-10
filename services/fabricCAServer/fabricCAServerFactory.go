package fabricCAServer

import (
	"fmt"
	"io"
	"text/template"

	"github.com/zhaizhonghao/registrarHelper/services/dockerCompose"
)

//GenerateConfigTxTemplate To write the config to the w according to the template tpl
func GenerateConfigTemplate(containerConfiguration dockerCompose.ContainerConfiguration, tpl *template.Template, w io.Writer) error {
	err := tpl.Execute(w, containerConfiguration)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
