package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"paisanos-cli/cmd/program"
	"paisanos-cli/cmd/ui/flag"
	"paisanos-cli/cmd/ui/multiInput"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Global style definitions for text and spinner.
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	installing   = lipgloss.NewStyle().Foreground(lipgloss.Color("44")).Render
	installed    = lipgloss.NewStyle().Foreground(lipgloss.Color("29")).Render
	skipped      = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render
)

// Lists for normal (formula) and cask installations.
var normalInstallations = []string{
	"fnm",
}

var caskInstallations = []string{
	"figma",
	"notion",
	"slack",
	"google-chrome",
}

type Options struct {
	Editor *multiInput.Selection
}

type step struct {
	description string   // A description of the step.
	command     string   // The command to execute.
	args        []string // Arguments for the command.
}

// installingDescription returns the installation description for a package.
func installingDescription(pkg string) string {
	return installing(fmt.Sprintf("Instalando %s...", pkg))
}

func alreadyInstalled(pkg string) string {
	return skipped(fmt.Sprintf("■ %s ya se encuentra instalado.", pkg))
}

func successfullyInstalled(pkg string) string {
	return installed(fmt.Sprintf("✔  %s instalado correctamente.", pkg))
}

// commandResultMsg is the message returned when a command completes.
type commandResultMsg struct {
	stepIndex int
	err       error
}

// setupModel is the Bubble Tea model that runs our setup steps.
type setupModel struct {
	spinner     spinner.Model
	steps       []step
	currentStep int
	done        bool
	err         error
}

// Init starts the spinner and executes the first step.
func (m *setupModel) Init() tea.Cmd {
	if len(m.steps) == 0 {
		m.done = true

		return nil
	}

	if len(m.steps) > 0 {
		return tea.Batch(m.spinner.Tick, runCommand(m.steps[m.currentStep], m.currentStep))
	}
	return m.spinner.Tick
}

// Update handles messages (spinner ticks and command results).
func (m *setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If the setup is done, ignore further messages and quit.
	if m.done {
		return m, tea.Quit
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		// Only schedule new ticks if not done.
		cmds = append(cmds, cmd)

	case commandResultMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		// Print success message if appropriate.
		prevStep := m.steps[m.currentStep]
		if strings.HasPrefix(prevStep.description, "▶ Instalando ") &&
			!strings.Contains(prevStep.description, "Homebrew") {
			pkg := strings.TrimSuffix(strings.TrimPrefix(prevStep.description, "▶ Instalando "), "...")
			fmt.Println(successfullyInstalled(pkg))
		}
		m.currentStep++
		if m.currentStep < len(m.steps) {
			cmds = append(cmds, runCommand(m.steps[m.currentStep], m.currentStep))
		} else {
			m.done = true
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI of the setup.
func (m *setupModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n%s\n", textStyle(fmt.Sprintf("Error: %v", m.err)))
	}
	if m.done {
		return textStyle("\nTu setup se ha completado correctamente 🚀\n")
	}
	desc := m.steps[m.currentStep].description
	return fmt.Sprintf("\n%s %s\n", m.spinner.View(), textStyle(desc))
}

// runCommand returns a Tea command that executes a step.
// For the Homebrew installation step it sets NONINTERACTIVE=1.
func runCommand(s step, index int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(s.command, s.args...)
		if s.description == "Installing Homebrew..." {
			cmd.Env = append(os.Environ(), "NONINTERACTIVE=1")
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return commandResultMsg{
				stepIndex: index,
				err:       fmt.Errorf("%q failed: %v (%s)", s.description, err, output),
			}
		}
		return commandResultMsg{stepIndex: index, err: nil}
	}
}

// newSetupModel creates a new setup model with our steps and a single spinner.
func newSetupModel(steps []step) *setupModel {
	sp := spinner.New()
	sp.Style = spinnerStyle
	sp.Spinner = spinner.Line
	return &setupModel{
		spinner:     sp,
		steps:       steps,
		currentStep: 0,
		done:        false,
	}
}

type listOptions struct {
	options []string
}

// fileExists returns true if the given filename exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// SetupCmd is a Cobra command that sets up your macOS environment.
var SetupCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure this command runs only on macOS.
		if runtime.GOOS != "darwin" {
			fmt.Println("Este comando solo funciona en macOS.")
			return
		}

		program := program.Project{}

		tprogram := tea.NewProgram(flag.InitialModelFlag(&program))
		if _, err := tprogram.Run(); err != nil {
			fmt.Printf("Error during setup: %v\n", err)
			os.Exit(1)
		}

		listOfEditors := listOptions{
			options: []string{"neovim", "cursor", "visual-studio-code"},
		}

		options := Options{
			Editor: &multiInput.Selection{},
		}

		tprogram = tea.NewProgram(multiInput.InitialModelMulti(listOfEditors.options, options.Editor, "Selecciona tu editor de confianza", &program))
		if _, err := tprogram.Run(); err != nil {
			fmt.Printf("Error during setup: %v\n", err)
			os.Exit(1)
		}

		if options.Editor.Choice == "neovim" {
			fmt.Println("Ninja neovim detectado 🥷")
			normalInstallations = append(normalInstallations, "neovim")
		} else {
			caskInstallations = append(caskInstallations, options.Editor.Choice)
		}

		// Retrieve current user's home directory.
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Error retrieving current user: %v\n", err)
			return
		}
		profilePath := usr.HomeDir + "/.zprofile"

		var steps []step
		brewInstalled := false

		// Check if Homebrew is installed.
		if _, err := exec.LookPath("brew"); err != nil {
			// Homebrew is not installed; add installation steps.
			steps = append(steps, step{
				description: "Instalando Homebrew...",
				command:     "/bin/bash",
				args: []string{
					"-c",
					"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)",
				},
			})
			steps = append(steps, step{
				description: "Configurando Homebrew...",
				command:     "/bin/bash",
				args: []string{
					"-c",
					fmt.Sprintf(`(echo; echo 'eval "$(/opt/homebrew/bin/brew shellenv)"') >> %s`, profilePath),
				},
			})
			steps = append(steps, step{
				description: "Evaluando entorno de Homebrew...",
				command:     "/bin/bash",
				args:        []string{"-c", `eval "$(/opt/homebrew/bin/brew shellenv)"`},
			})
		} else {
			brewInstalled = true
			fmt.Println("Homebrew ya se encuentra instalada, saltando instalación.")
		}

		// Append normal (formula) installation steps.
		for _, pkg := range normalInstallations {
			if brewInstalled {
				// Check if the package is already installed.
				if err := exec.Command("brew", "list", pkg).Run(); err != nil {
					steps = append(steps, step{
						description: installingDescription(pkg),
						command:     "brew",
						args:        []string{"install", pkg},
					})
				} else {
					fmt.Println(alreadyInstalled(pkg))
				}
			} else {
				steps = append(steps, step{
					description: installingDescription(pkg),
					command:     "brew",
					args:        []string{"install", pkg},
				})
			}
		}

		// Append cask installation steps.
		for _, pkg := range caskInstallations {
			if brewInstalled {
				// Special check for Google Chrome.
				if pkg == "google-chrome" {
					if fileExists("/Applications/Google Chrome.app") {
						fmt.Println(alreadyInstalled(pkg))
						continue
					}
					if err := exec.Command("brew", "list", "--cask", pkg).Run(); err != nil {
						steps = append(steps, step{
							description: installingDescription(pkg),
							command:     "brew",
							args:        []string{"install", "--cask", pkg},
						})
					} else {
						fmt.Println(alreadyInstalled(pkg))
					}
				} else {
					if err := exec.Command("brew", "list", "--cask", pkg).Run(); err != nil {
						steps = append(steps, step{
							description: installingDescription(pkg),
							command:     "brew",
							args:        []string{"install", "--cask", pkg},
						})
					} else {
						fmt.Println(alreadyInstalled(pkg))
					}
				}
			} else {
				steps = append(steps, step{
					description: installingDescription(pkg),
					command:     "brew",
					args:        []string{"install", "--cask", pkg},
				})
			}
		}

		// Create and start the Bubble Tea program with our steps.
		m := newSetupModel(steps)

		tprogram = tea.NewProgram(m)
		if _, err := tprogram.Run(); err != nil {
			fmt.Printf("Error during setup: %v\n", err)
			os.Exit(1)
		}
	},
}
