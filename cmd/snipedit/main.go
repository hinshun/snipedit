package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	CursorLocationPattern = regexp.MustCompile(`\%([^%]+)\%`)
	focusedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle           = focusedStyle.Copy()
	noStyle               = lipgloss.NewStyle()
)

func main() {
	err := run(context.Background(), os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	snippet, err := NewSnippet(strings.Join(args, " "))
	if err != nil {
		return err
	}

	p := tea.NewProgram(initialModel(snippet), tea.WithOutput(os.Stderr))
	err = p.Start()
	if err != nil {
		return err
	}

	fmt.Print(snippet.view(false))
	return nil
}

type Position struct {
	ID         string
	Start, End int
}

type Snippet struct {
	text string

	focusID   string
	sortedPos []string
	posMap    map[string][]Position
	posText   map[string]string
}

func NewSnippet(text string) (*Snippet, error) {
	matches := CursorLocationPattern.FindAllStringSubmatchIndex(text, -1)

	var sortedPos []string
	posMap := make(map[string][]Position)
	for _, match := range matches {
		if len(match) != 4 {
			continue
		}

		start, end := match[0], match[1]
		id := text[match[2]:match[3]]
		if _, ok := posMap[id]; !ok {
			sortedPos = append(sortedPos, id)
		}

		posMap[id] = append(posMap[id], Position{id, start, end})
	}

	// for i, pos := range sortedPos {
	// 	fmt.Printf("%s[%d] %v\n", i, pos, posMap[pos])
	// }

	return &Snippet{
		text:      text,
		sortedPos: sortedPos,
		posMap:    posMap,
		posText:   make(map[string]string),
	}, nil
}

func (s *Snippet) View() string {
	return s.view(true)
}

func (s *Snippet) view(style bool) string {
	var allPositions []Position
	for _, positions := range s.posMap {
		allPositions = append(allPositions, positions...)
	}
	sort.Slice(allPositions, func(i, j int) bool {
		return allPositions[i].Start < allPositions[j].Start
	})

	text := s.text
	offset := 0
	for _, pos := range allPositions {
		sub, ok := s.posText[pos.ID]
		if !ok || sub == "" {
			sub = fmt.Sprintf("%%%s%%", pos.ID)
		}
		if style && pos.ID == s.focusID {
			sub = focusedStyle.Render(sub)
		}

		text = text[:pos.Start+offset] + sub + text[pos.End+offset:]
		offset += len(sub) - (pos.End - pos.Start)
	}

	return text
}

type tickMsg struct{}

type model struct {
	focusIndex int
	inputs     []textinput.Model
	snippet    *Snippet
	textInput  textinput.Model
}

func initialModel(snippet *Snippet) model {
	m := model{
		snippet: snippet,
		inputs:  make([]textinput.Model, len(snippet.sortedPos)),
	}

	for i := range m.inputs {
		t := textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32
		t.Prompt = fmt.Sprintf("%s> ", snippet.sortedPos[i])

		m.inputs[i] = t
	}
	_ = m.updateFocus()

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs)-1 {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			return m, m.updateFocus()
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m model) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i <= len(m.inputs)-1; i++ {
		if i == m.focusIndex {
			// Set focused state
			m.snippet.focusID = m.snippet.sortedPos[i]
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = focusedStyle
			m.inputs[i].TextStyle = focusedStyle
			continue
		}
		// Remove focused state
		m.inputs[i].Blur()
		m.inputs[i].PromptStyle = noStyle
		m.inputs[i].TextStyle = noStyle
	}
	return tea.Batch(cmds...)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		id := m.snippet.sortedPos[i]
		m.snippet.posText[id] = m.inputs[i].Value()
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.snippet.View() + "\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	return b.String() + "\n"
}
