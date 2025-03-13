/*
Copyright © 2025 Paisanos <hola@paisanos.com>
*/
package main

import (
	"fmt"
	"math/rand"
	"paisanos-cli/cmd"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	width       = 70 // Increased width to accommodate spacing
	centerText  = "🥷 neovim ninja detected"
	sideSpacing = 11 // Spaces between text and binary on each side
)

type model struct {
	top         []rainChar
	bottom      []rainChar
	middleLeft  []rainChar
	middleRight []rainChar
	tick        int
}

type rainChar struct {
	char     string
	age      int
	updateAt int // determines update frequency
}

func initialModel() model {
	// Calculate text positioning with spacing
	textWithSpacing := strings.Repeat(" ", sideSpacing) + centerText + strings.Repeat(" ", sideSpacing)
	textStart := (width - len(textWithSpacing)) / 2
	textEnd := textStart + len(textWithSpacing)

	// Initialize top line
	top := make([]rainChar, width)
	for i := range top {
		top[i] = rainChar{
			char:     randomBinary(),
			age:      rand.Intn(10),
			updateAt: rand.Intn(3) + 1,
		}
	}

	// Initialize bottom line
	bottom := make([]rainChar, width)
	for i := range bottom {
		bottom[i] = rainChar{
			char:     randomBinary(),
			age:      rand.Intn(10),
			updateAt: rand.Intn(3) + 1,
		}
	}

	// Middle line left side (before text start)
	middleLeft := make([]rainChar, textStart)
	for i := range middleLeft {
		middleLeft[i] = rainChar{
			char:     randomBinary(),
			age:      rand.Intn(10),
			updateAt: rand.Intn(3) + 1,
		}
	}

	// Middle line right side (after text end)
	middleRight := make([]rainChar, width-textEnd)
	for i := range middleRight {
		middleRight[i] = rainChar{
			char:     randomBinary(),
			age:      rand.Intn(10),
			updateAt: rand.Intn(3) + 1,
		}
	}

	return model{
		top:         top,
		bottom:      bottom,
		middleLeft:  middleLeft,
		middleRight: middleRight,
		tick:        0,
	}
}

func randomBinary() string {
	if rand.Intn(2) == 0 {
		return "0"
	}
	return "1"
}

type tickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tickMsg:
		m.tick++

		// Update all lines of binary characters
		updateLine := func(line []rainChar) {
			for i := range line {
				if m.tick%line[i].updateAt == 0 {
					line[i].age++
					// Occasionally change the character
					if rand.Intn(3) == 0 || line[i].age > 15 {
						line[i].char = randomBinary()
						if line[i].age > 15 {
							line[i].age = 0
						}
					}
				}
			}
		}

		updateLine(m.top)
		updateLine(m.bottom)
		updateLine(m.middleLeft)
		updateLine(m.middleRight)

		return m, tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	// Render a line of binary characters with color based on age
	renderBinary := func(line []rainChar) {
		for i := range line {
			brightness := 255 - (line[i].age * 15)
			if brightness < 50 {
				brightness = 50
			}
			hexColor := fmt.Sprintf("#00%02X00", brightness)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(hexColor))
			sb.WriteString(style.Render(line[i].char))
		}
	}

	// Top line (binary rain)
	renderBinary(m.top)
	sb.WriteString("\n")

	// Middle line with text and binary on sides
	// First the left side binary
	renderBinary(m.middleLeft)

	// Then the centered text with spacing
	textStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	sb.WriteString(strings.Repeat(" ", sideSpacing+1))
	sb.WriteString(textStyle.Render(centerText))
	sb.WriteString(strings.Repeat(" ", sideSpacing+1))

	// Then the right side binary
	renderBinary(m.middleRight)
	sb.WriteString("\n")

	// Bottom line (binary rain)
	renderBinary(m.bottom)

	return sb.String()
}

func main() {
	cmd.Execute()
}
