package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"

	"github.com/manifoldco/promptui"
)

var (
	gcloud = flag.String("gcloud", "gcloud", "gcloud command")
)

func main() {
	flag.Parse()
	configs := listConfigs()
	displayAndChooseConfig(configs)
}

type gcloudConfig struct {
	IsActive   bool                   `json:"is_active"`
	Name       string                 `json:"name"`
	Properties gcloudConfigProperties `json:"properties"`
}

type gcloudConfigProperties struct {
	ApiEndpointOverrides gcloudConfigApiEndpointOverrides `json:"api_endpoint_overrides"`
	Core                 gcloudConfigCore                 `json:"core"`
}

type gcloudConfigApiEndpointOverrides struct {
	Dataproc string `json:"dataproc"`
}

type gcloudConfigCore struct {
	Account string `json:"account"`
	Project string `json:"project"`
}

func listConfigs() []gcloudConfig {
	b, err := exec.Command(*gcloud, "config", "configurations", "list", "--format=json").CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("listing configurations: %w", err))
	}
	var configs []gcloudConfig
	if err := json.Unmarshal(b, &configs); err != nil {
		panic(fmt.Errorf("json unmarshalling bytes: %q, %w", string(b), err))
	}
	return configs
}

func useConfig(c gcloudConfig) []byte {
	b, err := exec.Command(*gcloud, "config", "configurations", "activate", c.Name).CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("activating configuration: %q, %w", string(b), err))
	}
	return b
}

func displayAndChooseConfig(configs []gcloudConfig) {
	var activeIndex int
	for i, c := range configs {
		if c.IsActive {
			activeIndex = i
			break
		}
	}

	prompt := promptui.Select{
		Label: "gcloud configuration",
		Items: configs,
		Templates: &promptui.SelectTemplates{
			Active: "{{ .Name | cyan | underline }}" +
				"{{ if .IsActive }} {{- \" (Active)\" | cyan | underline }} {{ end }}",
			Inactive: "{{ .Name }}{{ if .IsActive }} {{- \" (Active)\" }} {{ end }}",
			Details: "Account: {{ .Properties.Core.Account }}\t" +
				"Project: {{ .Properties.Core.Project }}" +
				"{{- if .Properties.ApiEndpointOverrides.Dataproc }} " +
				"\tDataproc: {{ .Properties.ApiEndpointOverrides.Dataproc }} " +
				"{{ end }}",
		},
		Size: len(configs),
		HideSelected: true,
	}
	i, _, err := prompt.RunCursorAt(activeIndex, 0)
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return
	}
	b := useConfig(configs[i])
	fmt.Print(string(b))
}
