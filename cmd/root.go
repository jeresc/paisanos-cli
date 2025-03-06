package cmd

import (
	"fmt"
	"os"
	"os/user"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var border = lipgloss.Border{
	Left:        "│",
	Right:       "│",
	Top:         "─",
	Bottom:      "─",
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
}

// flagFrames holds the individual frames of the flag animation.
var flagFrames = []string{
	` █▀▀▃▃▃ 
 █▀▀▃▃█ 
 █      `,
	` █▀▀▀▃▃ 
 █▀▀▀▃█ 
 █      `,
	` █▀▀▀▀▃ 
 █▀▀▀▀█ 
 █      `,
	` █▃▀▀▀█ 
 █▃▀▀▀█ 
 █      `,
	` █▃▃▀▀█ 
 █▃▃▀▀▀ 
 █      `,
	` █▃▃▃▀█ 
 █▃▃▃▀▀ 
 █      `,
	` █▃▃▃▃█ 
 █▃▃▃▃▀ 
 █      `,
	` █▀▃▃▃▃ 
 █▀▃▃▃█ 
 █      `,
}

// primaryBg is the Lip Gloss style used to render both the flag and text.
var (
	primaryBg   = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(0)).Background(lipgloss.Color("226"))
	primary     = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	username, _ = getUserName()
)

// Define separate tick message types.
type (
	flagTickMsg time.Time
	textTickMsg time.Time
)

// Define phase constants for the typewriter effect.
const (
	phaseTyping  = "typing"
	phaseWaiting = "waiting"
)

// model defines the Bubble Tea model.
type model struct {
	frames          []string
	currentFrame    int
	startTime       time.Time
	phase           string    // phase can be "typing" or "waiting"
	typewriterIndex int       // index into current message
	messages        []string  // two messages to display in sequence
	currentMessage  int       // index of the current message
	waitStart       time.Time // time when waiting started
}

func getUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error getting current user: %w", err)
	}
	return currentUser.Username, nil
}

// Init schedules both tick commands.
func (m model) Init() tea.Cmd {
	return tea.Batch(tickFlagCmd(), tickTextCmd())
}

// Update handles messages for flag and text ticks as well as keys.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flagTickMsg:
		// Update the flag frame at a 150ms rate.
		m.currentFrame = (m.currentFrame + 1) % len(m.frames)
		return m, tickFlagCmd()

	case textTickMsg:
		// Update the typewriter effect at a 50ms rate.
		switch m.phase {
		case phaseTyping:
			// Append one character.
			if m.typewriterIndex < len(m.messages[m.currentMessage]) {
				m.typewriterIndex++
			} else {
				// When the first message is complete, wait 1 second.
				if m.currentMessage == 0 {
					m.phase = phaseWaiting
					m.waitStart = time.Now()
				}
			}
		case phaseWaiting:
			// If 1 second has elapsed, immediately clear text and switch to the second message.
			if time.Since(m.waitStart) >= 2*time.Second {
				m.currentMessage = 1
				m.typewriterIndex = 0
				m.phase = phaseTyping
			}
		}
		return m, tickTextCmd()

	case tea.KeyMsg:
		// Allow quitting early with "q" or Ctrl+C.
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}
	return m, nil
}

// View renders the flag and the text block.
func (m model) View() string {
	// Render the animated flag.
	flagView := primary.BorderStyle(border).Render(m.frames[m.currentFrame])
	// Static label "Paisabot:".
	label := primary.Render("Paisabot:")
	// Build the dynamic text using the typewriter effect.
	dynamicText := ""
	if m.typewriterIndex <= len(m.messages[m.currentMessage]) {
		dynamicText = m.messages[m.currentMessage][:m.typewriterIndex]
	} else {
		dynamicText = m.messages[m.currentMessage]
	}
	// Join the label and dynamic text vertically.
	textBlock := lipgloss.JoinVertical(lipgloss.Left, label, "  "+dynamicText)
	// Join the flag and text block horizontally with a margin.
	combined := lipgloss.JoinHorizontal(lipgloss.Top, flagView, "\n  "+textBlock)
	return combined
}

// tickFlagCmd returns a command for the flag animation at 150ms.
func tickFlagCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return flagTickMsg(t)
	})
}

// tickTextCmd returns a command for the typewriter effect at 50ms.
func tickTextCmd() tea.Cmd {
	return tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg {
		return textTickMsg(t)
	})
}

// rootCmd is the root Cobra command.
var rootCmd = &cobra.Command{
	Use:   "flaganimator",
	Short: "Animates a flag with a typewriter text effect",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Flag animation starting...")
		// Initialize the model with two messages.
		m := model{
			frames:          flagFrames,
			currentFrame:    0,
			startTime:       time.Now(),
			phase:           phaseTyping,
			typewriterIndex: 0,
			messages: []string{
				fmt.Sprintf("Bienvenido a paisanos, %s.", username),
				"¡Juntos vamos a conquistar el mundo!",
			},
			currentMessage: 0,
		}
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Flag animation stopped")
	},
}

// Execute runs the Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
