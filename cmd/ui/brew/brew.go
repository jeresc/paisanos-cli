package brew

import (
	"paisanos-cli/cmd/program"

	tea "github.com/charmbracelet/bubbletea"
)

// model define el modelo de Bubble Tea para la animación del flag.
type model struct {
	normalPkgs []string
	caskPkgs   []string
	exit       *bool
}

func InitialModelBrew(program *program.Project) model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Batch()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			*m.exit = true
			return m, tea.Quit
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	return ""
}
