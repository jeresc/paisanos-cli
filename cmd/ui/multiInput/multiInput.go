package multiInput

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
)

type Selection struct {
	Choice string
}

func (s *Selection) Update(value string) {
	s.Choice = value
}

type model struct {
	cursor   int
	choices  []string
	selected map[int]struct{}
	choice   *Selection
	header   string
}

func (m model) Init() tea.Cmd {
	return nil
}

func InitialModelMulti(choices []string, selection *Selection, header string) model {
	return model{
		choices:  choices,
		selected: make(map[int]struct{}),
		choice:   selection,
		header:   titleStyle(header),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			if len(m.selected) == 1 {
				m.selected = make(map[int]struct{})
			}
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "y":
			if len(m.selected) == 1 {
				for selectedKey := range m.selected {
					m.choice.Update(m.choices[selectedKey])
					m.cursor = selectedKey
				}
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := m.header + "\n\n"

	for i, choice := range m.choices {
		cursor := ""
		if i == m.cursor {
			cursor = focusedStyle(">")
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = focusedStyle("x")
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, checked, choice)
	}

	s += fmt.Sprintf("\n Press %s to confirm choice.", focusedStyle("y"))

	return s
}
