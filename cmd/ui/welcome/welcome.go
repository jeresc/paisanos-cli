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

const brandColor = "190"

// primaryBg is the Lip Gloss style used to render both the flag and text.
var (
	primaryBg = lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(0)).
			Background(lipgloss.Color(brandColor))
	primary = lipgloss.NewStyle().
		Foreground(lipgloss.Color(brandColor)).
		Background(lipgloss.ANSIColor(0))
	username, _ = getUserName()
)

// Define separate tick message types.
type (
	flagTickMsg time.Time
	textTickMsg time.Time
)

// Define phase constants for the typewriter effect.
const (
	phaseTyping       = "typing"
	phaseWaiting      = "waiting"      // Waiting after the first message
	phaseWaitingClear = "waitingClear" // Waiting after the second message before clearing
	phaseCleared      = "cleared"      // Cleared state (nothing is rendered)
)

// model defines the Bubble Tea model.
type model struct {
	frames          []string
	currentFrame    int
	startTime       time.Time
	phase           string    // phase: typing, waiting, waitingClear, cleared
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
		// If we've cleared everything, don't continue the animation.
		if m.phase == phaseCleared {
			return m, nil
		}
		// Update the flag frame at a 150ms rate.
		m.currentFrame = (m.currentFrame + 1) % len(m.frames)
		return m, tickFlagCmd()

	case textTickMsg:
		switch m.phase {
		case phaseTyping:
			// Append one character.
			if m.typewriterIndex < len(m.messages[m.currentMessage]) {
				m.typewriterIndex++
			} else {
				// Finished typing the current message.
				if m.currentMessage == 0 {
					// After the first message, wait 1 second before moving on.
					m.phase = phaseWaiting
					m.waitStart = time.Now()
				} else {
					// After the second message, wait 1 second and then clear.
					m.phase = phaseWaitingClear
					m.waitStart = time.Now()
				}
			}
		case phaseWaiting:
			// After the first message, if 1 second has elapsed, switch to the second message.
			if time.Since(m.waitStart) >= 1*time.Second {
				m.currentMessage = 1
				m.typewriterIndex = 0
				m.phase = phaseTyping
			}
		case phaseWaitingClear:
			// After the second message, if 1 second has elapsed, clear everything.
			if time.Since(m.waitStart) >= 1*time.Second {
				m.phase = phaseCleared
			}
		case phaseCleared:
			return m, tea.Quit
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
// When in "cleared" phase, it returns an empty view.
func (m model) View() string {
	if m.phase == phaseCleared {
		return ""
	}
	// Render the animated flag.
	flagView := primaryBg.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBackground(lipgloss.Color(brandColor)).
		BorderForeground(lipgloss.Color("0")).
		Render(m.frames[m.currentFrame])
	// Static label "Paisabot:".
	label := primary.Render("Paisabot:")
	// Build the dynamic text using the typewriter effect.
	var dynamicText string
	if m.typewriterIndex <= len(m.messages[m.currentMessage]) {
		dynamicText = m.messages[m.currentMessage][:m.typewriterIndex]
	} else {
		dynamicText = m.messages[m.currentMessage]
	}
	// Join the label and dynamic text vertically.
	textBlock := lipgloss.JoinVertical(lipgloss.Left, label, "  "+dynamicText)
	// Join the flag and text block horizontally with a margin.
	combined := lipgloss.JoinHorizontal(lipgloss.Top, flagView+"\n", "\n  "+textBlock)
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

// WelcomeCmd is the root Cobra command.
var WelcomeCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
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
	},
}
