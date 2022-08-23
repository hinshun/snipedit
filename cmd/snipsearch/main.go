package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	_ "github.com/hinshun/snipedit/tui"
)

const listHeight = 14

var (
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type model struct {
	list     list.Model
	choice   *string
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
			i, ok := m.list.SelectedItem().(item)
			if ok {
				*m.choice = i.desc
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

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	items := []list.Item{
		item{"Remove a single git commit from history and its changes", "git rebase --onto %commit%^ %commit%"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, listHeight)
	l.Title = "Snippet Search"

	var choice string
	m := model{list: l, choice: &choice}

	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	err := p.Start()
	if err != nil {
		return err
	}

	fmt.Println(choice)
	return nil
}
