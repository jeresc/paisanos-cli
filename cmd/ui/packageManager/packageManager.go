package packageManager

import (
	"fmt"
	"math/rand"
	"paisanos-cli/cmd/program"
	"paisanos-cli/utils"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BrewPackageType string

const (
	Formula BrewPackageType = "formula" // Normal packages
	Cask    BrewPackageType = "cask"    // GUI or cask packages
)

type Package struct {
	DisplayName string
	BrewName    string
	Kind        BrewPackageType
}

type model struct {
	packages []Package
	index    int
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

func (m model) Init() tea.Cmd {
	return tea.Batch(downloadAndInstall(m.packages[m.index]), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case installedPkgMsg:
		pkg := m.packages[m.index]
		if m.index >= len(m.packages)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s %s", checkMark, pkg.DisplayName), // print the last success message
				tea.Quit, // exit the program
			)
		}

		// Update progress bar
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.packages)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s %s", checkMark, pkg.DisplayName), // print success message above our program
			downloadAndInstall(m.packages[m.index]),         // download the next package
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	n := len(m.packages)
	w := lipgloss.Width(fmt.Sprintf("%d", n))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("¡Setup completado! %d paquetes instalados.\n", n))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.index, w, n)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := utils.Max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	pkgName := currentPkgNameStyle.Render(m.packages[m.index].DisplayName)
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Instalando " + pkgName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

type installedPkgMsg string

func downloadAndInstall(pkg Package) tea.Cmd {
	// This is where you'd do i/o stuff to download and install packages. In
	// our case we're just pausing for a moment to simulate the process.
	d := time.Millisecond * time.Duration(rand.Intn(500)) //nolint:gosec
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return installedPkgMsg(pkg.BrewName)
	})
}

func InitialModelPkgManager(packages []Package, program *program.Project) model {
	return model{
		packages: packages,
		spinner:  spinner.New(),
		progress: progress.New(
			progress.WithScaledGradient("#000000", "#efff00"),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		),
	}
}
