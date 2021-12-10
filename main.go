package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os/exec"

	"github.com/rivo/tview"
)

var (
	gcloud = flag.String("gcloud", "gcloud", "gcloud command")
)

func main() {
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
	app := tview.NewApplication()
	list := tview.NewList()

	selected := func(i int) {
		b := useConfig(configs[i])
		app.Stop()
		fmt.Println(string(b))
	}

	shortcut := 'a'
	for i, c := range configs {
		i := i
		activeText := ""
		if c.IsActive {
			activeText = " (active)"
		}
		secondaryText := fmt.Sprintf("Account: %s | Project: %s",
			c.Properties.Core.Account, c.Properties.Core.Project)
		if dp := c.Properties.ApiEndpointOverrides.Dataproc; dp != "" {
			secondaryText = fmt.Sprintf("%s | Dataproc: %s", secondaryText, dp)
		}
		list.AddItem(
			fmt.Sprintf("%s%s", c.Name, activeText),
			secondaryText,
			shortcut,
			func() {
				selected(i)
			})
		shortcut += 1
		// 'q' is reserved for quit.
		if shortcut == 'q' {
			shortcut += 1
		}
	}

	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
		fmt.Println("Exiting without changing configurations")
	})

	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(fmt.Errorf("running app: %w", err))
	}
}
