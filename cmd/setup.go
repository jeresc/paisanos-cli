package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Global style definitions from your spinner example.
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

// Lists for normal (formula) and cask installations.
var normalInstallations = []string{
	"neovim",
}

var caskInstallations = []string{
	"google-chrome",
	"figma",
}

// step represents a single installation step.
type step struct {
	description string   // A description of the step.
	command     string   // The command to execute.
	args        []string // Arguments for the command.
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
	if len(m.steps) > 0 {
		return tea.Batch(m.spinner.Tick, runCommand(m.steps[m.currentStep], m.currentStep))
	}
	return m.spinner.Tick
}

// Update handles messages (spinner ticks and command results).
func (m *setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case commandResultMsg:
		// If any command fails, store the error and quit.
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		// Move on to the next step.
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
		return textStyle("\nSetup complete!\n")
	}
	desc := m.steps[m.currentStep].description
	return fmt.Sprintf("\n%s %s\n", m.spinner.View(), textStyle(desc))
}

// runCommand returns a Tea command that executes a step.
func runCommand(s step, index int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(s.command, s.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Include the command output for better debugging.
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
	// Use our single spinner with a fixed style (spinner.Line).
	sp.Style = spinnerStyle
	sp.Spinner = spinner.Line
	return &setupModel{
		spinner:     sp,
		steps:       steps,
		currentStep: 0,
		done:        false,
	}
}

// SetupCmd is a Cobra command that sets up your macOS environment.
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up macOS by installing Homebrew and required applications",
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure this command runs only on macOS.
		if runtime.GOOS != "darwin" {
			fmt.Println("This setup command only works on macOS.")
			return
		}

		// Retrieve current user's home directory.
		usr, err := user.Current()
		if err != nil {
			fmt.Printf("Error retrieving current user: %v\n", err)
			return
		}
		profilePath := usr.HomeDir + "/.zprofile"

		// Define the initial steps for installing Homebrew and configuring the shell.
		steps := []step{
			{
				description: "Installing Homebrew...",
				command:     "/bin/bash",
				args:        []string{"-c", `$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`},
			},
			{
				description: "Appending Homebrew environment to .zprofile...",
				command:     "/bin/bash",
				args: []string{
					"-c",
					fmt.Sprintf(`(echo; echo 'eval "$(/opt/homebrew/bin/brew shellenv)"') >> %s`,
						profilePath),
				},
			},
			{
				description: "Evaluating Homebrew environment...",
				command:     "/bin/bash",
				args:        []string{"-c", `eval "$(/opt/homebrew/bin/brew shellenv)"`},
			},
		}

		// Append formula installation steps.
		for _, pkg := range normalInstallations {
			steps = append(steps, step{
				description: fmt.Sprintf("Installing %s...", pkg),
				command:     "brew",
				args:        []string{"install", pkg},
			})
		}

		// Append cask installation steps.
		for _, pkg := range caskInstallations {
			steps = append(steps, step{
				description: fmt.Sprintf("Installing %s...", pkg),
				command:     "brew",
				args:        []string{"install", "--cask", pkg},
			})
		}

		// Create and start the Bubble Tea program with our single spinner.
		m := newSetupModel(steps)
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error during setup: %v\n", err)
			os.Exit(1)
		}
	},
}
