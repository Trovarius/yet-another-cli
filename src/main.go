package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	// "encoding/json"
	"io/ioutil"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Plugin struct {
	Version  string    `yaml:"version"`
	Name     string    `yaml:"name"`
	Desc     string    `yaml:"description"`
	Commands []Command `yaml:"commands"`
}

type Command struct {
	Name     string `yaml:"name"`
	Command  string `yaml:"command"`
	Executer string `yaml:"executer"`
}

type CommandArg struct {
	Arg         string `yaml:"name"`
	Description string `yaml:"description"`
}

func (c Command) Execute(path string) {
	fmt.Printf("\n\nExeucting %s:\n", c.Name)

	cmdString := fmt.Sprintf("%s/%s", path, c.Command)
	cmd := exec.Command("/bin/sh", cmdString)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error %s", err)
	}
}

type item struct {
	title, desc string
	cmd         Command
}

func (p item) FilterValue() string { return p.title }
func (p item) Title() string       { return p.title }
func (p item) Description() string { return p.desc }
func (p item) Command() Command    { return p.cmd }

type model struct {
	list     map[string]list.Model
	focused  []string
	selected list.Model
}

func initialModel(plugins []Plugin, focused string) model {
	fmt.Printf("Focused: %v", focused)
	if focused == "" {
		focused = "plugins"
	}

	items := make([]list.Item, 0, len(plugins))
	all := make(map[string]list.Model)

	for _, v := range plugins {
		commands := make([]list.Item, 0, len(v.Commands))

		for _, y := range v.Commands {
			// fmt.Printf("Comands %s: %s\n", x, y)
			commands = append(commands, item{title: y.Name, desc: v.Desc, cmd: y})
		}
		cmdList := list.New(commands, list.NewDefaultDelegate(), 50, 14)
		cmdList.Title = fmt.Sprintf("Commands from %v plugin:", v.Name)
		all[v.Name] = cmdList

		items = append(items, item{title: v.Name, desc: v.Desc})
	}

	pluginList := list.New(items, list.NewDefaultDelegate(), 50, 14)
	pluginList.Title = "My Plugins"
	all[focused] = pluginList

	m := model{list: all, focused: []string{focused}, selected: all[focused]}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			i, ok := m.selected.SelectedItem().(item)

			if ok {
				if i.cmd != (Command{}) {
					i.cmd.Execute(fmt.Sprintf("./plugins/%s", m.focused))

					return m, tea.Quit
				}

				m.focused = append(m.focused, i.title)
				m.selected = m.list[i.title]
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyLeft, tea.KeyCtrlBackslash:
			if len(m.focused) > 1 {
				m.focused = m.focused[:len(m.focused)-1]
				m.selected = m.list[m.focused[len(m.focused)-1]]
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		list := m.selected
		list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd

	m.selected, cmd = m.selected.Update(msg)
	return m, cmd
}

func (m model) View() string {

	return m.selected.View()
}

func WalkMatch(root, pattern string) ([]Plugin, error) {
	var matches []Plugin
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			yfile, err := ioutil.ReadFile(path)

			if err != nil {
				log.Fatal(err)
			}

			var plugin Plugin

			err2 := yaml.Unmarshal(yfile, &plugin)

			if err2 != nil {

				log.Fatal(err2)
			}

			matches = append(matches, plugin)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func init() {
	if os.Getenv("YACPATH") == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Home folder not fould: %v", err)
		}

		env := path.Join(home, ".yet-another-cli", "plugin")
		errDir := os.MkdirAll(env, 0777)
		if errDir != nil {
			fmt.Printf("Directory creation wnet wrong %v", err)
		}

		os.Setenv("YACPATH", env)
		fmt.Printf("folderpath: %v\n\n", os.Getenv("YACPATH"))
	}
}

func main() {
	var focused string
	if len(os.Args) > 1 {
		focused = os.Args[1]
	}

	var folderPath string = "./plugins"

	if os.Getenv("YACPATH") != "" {
		folderPath = os.Getenv("YACPATH")
	}

	plugins, err := WalkMatch(folderPath, "*.yml")
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}

	p := tea.NewProgram(initialModel(plugins, focused))
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
