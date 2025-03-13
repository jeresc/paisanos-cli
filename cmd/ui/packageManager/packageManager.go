package packageManager

import (
	"fmt"
	"os"
	"os/exec"
	"paisanos-cli/cmd/program"
	"paisanos-cli/utils"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	BrewPackageType string
	installedPkgMsg string
	skippedPkgMsg   string
)

const (
	Formula BrewPackageType = "formula" // Normal packages
	Cask    BrewPackageType = "cask"    // GUI or cask packages
)

type Package struct {
	DisplayName string
	BrewName    string
	Kind        BrewPackageType
	Disabled    bool
}

type model struct {
	packages []Package
	index    int
	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	done     bool
	exit     *bool
}

var (
	currentPkgNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
	skippedMark         = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).SetString("■")
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
			*m.exit = true
			return m, tea.Quit
		}
	case skippedPkgMsg:
		pkg := m.packages[m.index]
		if m.index >= len(m.packages)-1 {
			// Everything's been installed. We're done!
			m.done = true
			return m, tea.Sequence(
				tea.Printf("%s  ya se encuentra instalado.", skippedMark.Render(pkg.DisplayName)),
				tea.Quit, // exit the program
			)
		}

		// Update progress bar
		m.index++
		progressCmd := m.progress.SetPercent(float64(m.index) / float64(len(m.packages)))

		return m, tea.Batch(
			progressCmd,
			tea.Printf("%s ya se encuentra instalado.", skippedMark.Render(pkg.DisplayName)), // print success message above our program
			downloadAndInstall(pkg), // download the next package
		)
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
			downloadAndInstall(pkg),                         // download the next package
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

func downloadAndInstall(pkg Package) tea.Cmd {
	return func() tea.Msg {
		// Skip disabled packages
		if pkg.Disabled {
			return ""
		}

		if pkg.BrewName == "google-chrome" {
			if _, err := os.Stat("/Applications/Google Chrome.app"); err == nil {
				return skippedPkgMsg(pkg.BrewName)
			}
		}

		// Check if package is already installed
		args := []string{"list"}
		if pkg.Kind == Cask {
			args = append(args, "--cask")
		}
		args = append(args, pkg.BrewName)

		cmd := exec.Command("brew", args...)
		if err := cmd.Run(); err == nil {
			// Package is already installed
			return skippedPkgMsg(pkg.BrewName)
		}

		// Install the package
		installArgs := []string{"install"}
		if pkg.Kind == Cask {
			installArgs = append(installArgs, "--cask")
		}
		installArgs = append(installArgs, pkg.BrewName)

		installCmd := exec.Command("brew", installArgs...)

		// Capture output for logging/error reporting
		output, err := installCmd.CombinedOutput()
		if err != nil {
			// Return error message that could be handled in the Update method
			return installErrorMsg{
				pkg:    pkg,
				err:    err,
				output: string(output),
			}
		}

		return installedPkgMsg(pkg.BrewName)
	}
}

// Add this new type to handle installation errors
type installErrorMsg struct {
	pkg    Package
	err    error
	output string
}

func InitialModelPkgManager(packages []Package, program *program.Project) model {
	selectedPackage := []Package{}
	for _, pkg := range packages {
		if pkg.Disabled {
			continue
		}
		selectedPackage = append(selectedPackage, pkg)
	}

	return model{
		packages: selectedPackage,
		spinner:  spinner.New(),
		progress: progress.New(
			progress.WithScaledGradient("#000000", "#efff00"),
			progress.WithWidth(40),
			progress.WithoutPercentage(),
		),
		exit: &program.Exit,
	}
}
