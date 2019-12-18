package project

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/compose/loader"
	composetypes "github.com/docker/cli/cli/compose/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/manifoldco/promptui"
)

// Any type can be given to the select's item as long as the templates properly implement the dot notation
// to display it.
type config struct {
	Name        string
	Type        string
	Description string
	Config      composetypes.BuildConfig
}

func SelectConfig(projectDir string) composetypes.BuildConfig {
	cfg := composetypes.BuildConfig{Context: "."}
	files, err := ioutil.ReadDir(projectDir)
	if err != nil {
		return cfg
	}

	options := []config{}
	for _, f := range files {
		if f.IsDir() {
			if _, err := os.Stat(filepath.Join(projectDir, f.Name(), "docker-compose.yml")); err == nil {
				targets := getBuildTargets(f.Name(), filepath.Join(projectDir, f.Name(), "docker-compose.yml"))
				options = append(options, targets...)
			} else if _, err := os.Stat(filepath.Join(projectDir, f.Name(), "Dockerfile")); err == nil {
				o := config{Name: f.Name()}
				o.Type = "Dockerfile"
				cfg.Dockerfile = filepath.Join(projectDir, f.Name(), "Dockerfile")
				cfg.Context = "."
				if d, err := ioutil.ReadFile(filepath.Join(projectDir, f.Name(), "description.txt")); err == nil {
					o.Description = string(d)
				}
				o.Config = cfg
				options = append(options, o)
			}
		}
	}

	// The Active and Selected templates set a small config icon next to the name colored and the heat unit for the
	// active template. The details template is show at the bottom of the select's list and displays the full info
	// for that config in a multi-line template.
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   promptui.IconSelect + " {{ .Name | cyan }} ({{ .Type }})",
		Inactive: "  {{ .Name | cyan }} ({{ .Type }})",
		Selected: promptui.IconSelect + " {{ .Name | cyan }}",
		Details: `
{{ .Description | faint }}`,
	}

	// A searcher function is implemented which enabled the search mode for the select. The function follows
	// the required searcher signature and finds any config whose name contains the searched string.
	searcher := func(input string, index int) bool {
		opt := options[index]
		name := strings.Replace(strings.ToLower(opt.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select config",
		Items:     options,
		Templates: templates,
		Size:      4,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return composetypes.BuildConfig{}
	}

	fmt.Printf("Using config %q\n", options[i].Name)
	return options[i].Config
}

func SelectTemplate() string {
	home, _ := os.UserHomeDir()
	templateDir := filepath.Join(home, ".docker", "project-templates")
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return ""
	}

	options := []string{}
	for _, f := range files {
		if f.IsDir() {
			options = append(options, f.Name())
		}
	}

	prompt := promptui.Select{
		Label: "Select Docker template",
		Items: options,
		Size:  4,
	}

	_, dir, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	return filepath.Join(templateDir, dir)
}

func SelectComposeFile(projectDir string) string {
	files, err := ioutil.ReadDir(projectDir)
	if err != nil {
		return ""
	}

	options := []string{}
	for _, f := range files {
		if f.IsDir() {
			if _, err := os.Stat(filepath.Join(projectDir, f.Name(), "docker-compose.yml")); err == nil {
				options = append(options, f.Name())
			}
		}
	}

	// The Active and Selected templates set a small config icon next to the name colored and the heat unit for the
	// active template. The details template is show at the bottom of the select's list and displays the full info
	// for that config in a multi-line template.
	templates := &promptui.SelectTemplates{
		Label: "{{ . }}?",
	}

	prompt := promptui.Select{
		Label:     "Select stack",
		Items:     options,
		Templates: templates,
		Size:      4,
	}

	_, dir, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	fmt.Printf("Using stack %q\n", dir)
	return dir
}

func SelectImage(apiClient client.APIClient) string {
	imgs, err := apiClient.ImageList(context.TODO(), types.ImageListOptions{})
	if err != nil {
		return ""
	}
	images := []string{}
	for _, i := range imgs {
		for _, tag := range i.RepoTags {
			if tag != "<none>:<none>" {
				images = append(images, tag)
			}
		}
	}

	// A searcher function is implemented which enabled the search mode for the select. The function follows
	// the required searcher signature and finds any config whose name contains the searched string.
	searcher := func(input string, index int) bool {
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(images[index], input)
	}

	prompt := promptui.Select{
		Label:    "Select image",
		Items:    images,
		Size:     4,
		Searcher: searcher,
	}

	_, selected, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return ""
	}

	fmt.Printf("Using image %q\n", selected)
	return selected
}

func getBuildTargets(name, composefile string) []config {
	targets := []config{}
	data, err := ioutil.ReadFile(composefile)
	if err != nil {
		return targets
	}
	yamlData, err := loader.ParseYAML(data)
	if err != nil {
		fmt.Println(err.Error())
		return targets
	}
	cfg, err := loader.Load(composetypes.ConfigDetails{
		ConfigFiles: []composetypes.ConfigFile{
			{Config: yamlData, Filename: composefile},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return targets
	}
	if cfg.Services == nil {
		return targets
	}
	services := &cfg.Services
	for _, s := range *services {
		t := config{Name: name + " (" + s.Name + ")", Type: "Compose File Service"}
		t.Config = s.Build
		if t.Config.Dockerfile != "" {
			t.Config.Dockerfile = filepath.Join(filepath.Dir(composefile), t.Config.Dockerfile)
		}
		// HACK: USING NETWORK TO STORE TAG for now
		if s.Image != "" {
			t.Config.Network = s.Image
		}

		targets = append(targets, t)
	}
	return targets
}
