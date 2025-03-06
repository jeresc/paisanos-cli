package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------
// Global style definitions for text, spinner and editor list.
// ----------------------------------------------------------

var (
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	installing   = lipgloss.NewStyle().Foreground(lipgloss.Color("44")).Render
	installed    = lipgloss.NewStyle().Foreground(lipgloss.Color("29")).Render
	skipped      = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render

	// Styles for the editor selection list.
	editorTitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	editorItemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	editorSelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).
				Foreground(lipgloss.Color("170"))
	editorPaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	editorHelpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	editorQuitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

// ----------------------------------------------------------
// Installation lists (formulas and casks).
// ----------------------------------------------------------

var normalInstallations = []string{
	"neovim",
	"fnm",
}

var caskInstallations = []string{
	"figma",
	"notion",
	"slack",
	"google-chrome",
}

// installingDescription returns a formatted installation description.
func installingDescription(pkg string) string {
	return installing(fmt.Sprintf("Instalando %s...", pkg))
}

func alreadyInstalled(pkg string) string {
	return skipped(fmt.Sprintf("■ %s ya se encuentra instalado.", pkg))
}

// ----------------------------------------------------------
// Step definitions for setup.
// ----------------------------------------------------------

type step struct {
	description string   // A description of the step.
	command     string   // The command to execute.
	args        []string // Arguments for the command.
}

type commandResultMsg struct {
	stepIndex int
	err       error
}

// setupModel runs our installation steps with a spinner.
type setupModel struct {
	spinner     spinner.Model
	steps       []step
	currentStep int
	done        bool
	err         error
}

func (m *setupModel) Init() tea.Cmd {
	if len(m.steps) == 0 {
		m.done = true
		return nil
	}
	return tea.Batch(m.spinner.Tick, runCommand(m.steps[m.currentStep], m.currentStep))
}

func (m *setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case commandResultMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
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

func runCommand(s step, index int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(s.command, s.args...)
		if s.description == "Instalando Homebrew..." {
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

// fileExists returns true if the given file exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func successfullyInstalled(pkg string) string {
	return installed(fmt.Sprintf("■ %s se instaló correctamente.", pkg))
}

// ----------------------------------------------------------
// Editor selection phase using a list (new phase)
// ----------------------------------------------------------

// We'll create a simple list model to let the user pick the editor.

const editorListHeight = 6

type editorItem string

func (i editorItem) FilterValue() string { return "" }

type editorDelegate struct{}

func (d editorDelegate) Height() int  { return 1 }
func (d editorDelegate) Spacing() int { return 0 }
func (d editorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d editorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(editorItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, item)
	renderFn := editorItemStyle.Render
	if index == m.Index() {
		renderFn = func(s ...string) string {
			return editorSelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, renderFn(str))
}

type editorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m editorModel) Init() tea.Cmd {
	return nil
}

func (m editorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(editorItem)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m editorModel) View() string {
	if m.choice != "" {
		return editorQuitTextStyle.Render(
			fmt.Sprintf("Has seleccionado: %s", m.choice),
		)
	}
	if m.quitting {
		return editorQuitTextStyle.Render("No se seleccionó un editor.")
	}
	return "\n" + m.list.View()
}

func selectEditor() (string, error) {
	// Create list items for the available editors.
	items := []list.Item{
		editorItem("neovim"),
		editorItem("cursor.ai"),
		editorItem("vscode"),
	}

	const defaultWidth = 20
	l := list.New(items, editorDelegate{}, defaultWidth, editorListHeight)
	l.Title = "Selecciona el editor a instalar"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = editorTitleStyle
	l.Styles.PaginationStyle = editorPaginationStyle
	l.Styles.HelpStyle = editorHelpStyle

	m := editorModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}
	em, ok := finalModel.(editorModel)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}
	return em.choice, nil
}

// ----------------------------------------------------------
// SetupCmd Cobra command:
// First, it asks for an editor selection, then builds the install steps.
// ----------------------------------------------------------

var SetupCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure this command runs only on macOS.
		if runtime.GOOS != "darwin" {
			fmt.Println("Este comando solo funciona en macOS.")
			return
		}

		// ------------------------------
		// New Phase: Select Editor
		// ------------------------------
		editorChoice, err := selectEditor()
		if err != nil {
			fmt.Printf("Error selecting editor: %v\n", err)
			return
		}
		if editorChoice == "" {
			fmt.Println("No se seleccionó ningún editor, finalizando.")
			return
		}
		fmt.Printf("Editor seleccionado: %s\n", editorChoice)

		// Set flags based on the selected editor.
		selectedNeovim, selectedCursor, selectedVSCode := false, false, false
		switch strings.ToLower(editorChoice) {
		case "neovim":
			selectedNeovim = true
		case "cursor.ai":
			selectedCursor = true
		case "vscode":
			selectedVSCode = true
		}

		// ------------------------------
		// Continue with Setup
		// ------------------------------

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
			// Si el paquete es neovim y no fue seleccionado, se omite.
			if pkg == "neovim" && !selectedNeovim {
				fmt.Println("Omitiendo la instalación de neovim, no se seleccionó.")
				continue
			}
			if brewInstalled {
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

		// Append extra installation steps for the selected editor.
		if selectedVSCode {
			if brewInstalled {
				if err := exec.Command("brew", "list", "--cask", "visual-studio-code").
					Run(); err != nil {
					steps = append(steps, step{
						description: installingDescription("vscode"),
						command:     "brew",
						args:        []string{"install", "--cask", "visual-studio-code"},
					})
				} else {
					fmt.Println(alreadyInstalled("vscode"))
				}
			} else {
				steps = append(steps, step{
					description: installingDescription("vscode"),
					command:     "brew",
					args:        []string{"install", "--cask", "visual-studio-code"},
				})
			}
		}

		if selectedCursor {
			if brewInstalled {
				if err := exec.Command("brew", "list", "--cask", "cursor").Run(); err != nil {
					steps = append(steps, step{
						description: installingDescription("cursor.ai"),
						command:     "brew",
						args:        []string{"install", "--cask", "cursor"},
					})
				} else {
					fmt.Println(alreadyInstalled("cursor.ai"))
				}
			} else {
				steps = append(steps, step{
					description: installingDescription("cursor.ai"),
					command:     "brew",
					args:        []string{"install", "--cask", "cursor"},
				})
			}
		}

		// Create and start the Bubble Tea program with our steps.
		m := newSetupModel(steps)
		program := tea.NewProgram(m)
		if _, err := program.Run(); err != nil {
			fmt.Printf("Error during setup: %v\n", err)
			os.Exit(1)
		}
	},
}
