package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	yaml "gopkg.in/yaml.v3"

	_ "github.com/hinshun/snipedit/tui"
)

const listHeight = 14

type model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				m.choice = i.Snippet
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

type Config struct {
	Items    []Item
	Includes []string
}

type Item struct {
	Name    string
	Snippet string
}

func (i Item) Title() string       { return i.Name }
func (i Item) Description() string { return i.Snippet }
func (i Item) FilterValue() string { return i.Name }

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var items []list.Item

	rootCfg := "snipsearch.yaml"
	// rootCfg := fmt.Sprintf("%s/.config/snipsearch/snipsearch.yaml", os.Getenv("HOME"))
	_, err := os.Stat(rootCfg)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("No snippet config defined: %v", err)
		}
		return err
	}

	visited := make(map[string]struct{})
	stack := []string{rootCfg}

	for len(stack) > 0 {
		next := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		absPath, err := filepath.Abs(next)
		if err != nil {
			return err
		}
		if _, ok := visited[absPath]; ok {
			continue
		}
		visited[absPath] = struct{}{}

		f, err := os.Open(absPath)
		if err != nil {
			return err
		}
		defer f.Close()

		dt, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		var cfg Config
		err = yaml.Unmarshal(dt, &cfg)
		if err != nil {
			return err
		}

		for _, item := range cfg.Items {
			items = append(items, item)
		}

		for _, include := range cfg.Includes {
			stack = append(stack, include)
		}
	}

	if len(items) == 0 {
		return fmt.Errorf("No snippets found!")
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 18)
	l.Title = "Snippet Search"

	p := tea.NewProgram(model{list: l}, tea.WithOutput(os.Stderr))

	m, err := p.StartReturningModel()
	if err != nil {
		return err
	}

	fmt.Print(m.(model).choice)
	return nil
}
