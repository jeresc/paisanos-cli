package flag

import (
	"fmt"
	"os/user"
	"paisanos-cli/cmd/program"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Estilos utilizados para el flag y el texto.
var (
	primaryBg = lipgloss.NewStyle().
			Foreground(lipgloss.ANSIColor(0)).
			Background(lipgloss.Color("190"))
	primary = lipgloss.NewStyle().
		Foreground(lipgloss.Color("190")).
		Background(lipgloss.ANSIColor(0))
	username, _ = getUserName()
)

// flagFrames contiene los frames de la animación del flag.
var flagFrames = []string{
	" █▀▀▃▃▃ \n █▀▀▃▃█ \n █      ",
	" █▀▀▀▃▃ \n █▀▀▀▃█ \n █      ",
	" █▀▀▀▀▃ \n █▀▀▀▀█ \n █      ",
	" █▃▀▀▀█ \n █▃▀▀▀█ \n █      ",
	" █▃▃▀▀█ \n █▃▃▀▀▀ \n █      ",
	" █▃▃▃▀█ \n █▃▃▃▀▀ \n █      ",
	" █▃▃▃▃█ \n █▃▃▃▃▀ \n █      ",
	" █▀▃▃▃▃ \n █▀▃▃▃█ \n █      ",
}

// Definición de mensajes de tick para el flag y el texto.
type (
	flagTickMsg time.Time
	textTickMsg time.Time
)

// Constantes para las fases del efecto typewriter.
const (
	phaseTyping       = "typing"
	phaseWaiting      = "waiting"      // Espera luego del primer mensaje
	phaseWaitingClear = "waitingClear" // Espera antes de limpiar luego del segundo mensaje
	phaseCleared      = "cleared"      // Estado final (nada se renderiza)
)

// model define el modelo de Bubble Tea para la animación del flag.
type model struct {
	frames          []string  // frames de la animación
	currentFrame    int       // frame actual
	startTime       time.Time // tiempo de inicio
	phase           string    // fase actual: typing, waiting, waitingClear, cleared
	typewriterIndex int       // índice en el mensaje actual para el efecto typewriter
	messages        []string  // mensajes a mostrar (secuenciales)
	currentMessage  int       // índice del mensaje actual
	waitStart       time.Time // momento en que se inició la espera
	exit            *bool
}

// getUserName obtiene el nombre del usuario actual.
func getUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error obteniendo el usuario actual: %w", err)
	}
	return currentUser.Username, nil
}

// InitialModelFlag inicializa y retorna el modelo para el flag.
// Se reciben los frames y los mensajes a mostrar.
func InitialModelFlag(program *program.Project) model {
	return model{
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
		exit:           &program.Exit,
	}
}

// Init programa los ticks para la animación del flag y del texto.
func (m model) Init() tea.Cmd {
	return tea.Batch(tickFlagCmd(), tickTextCmd())
}

// Update procesa los mensajes (ticks y keys) y actualiza el modelo.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flagTickMsg:
		if m.phase == phaseCleared {
			return m, nil
		}
		m.currentFrame = (m.currentFrame + 1) % len(m.frames)
		return m, tickFlagCmd()

	case textTickMsg:
		switch m.phase {
		case phaseTyping:
			if m.typewriterIndex < len(m.messages[m.currentMessage]) {
				m.typewriterIndex++
			} else {
				if m.currentMessage == 0 {
					m.phase = phaseWaiting
					m.waitStart = time.Now()
				} else {
					m.phase = phaseWaitingClear
					m.waitStart = time.Now()
				}
			}
		case phaseWaiting:
			if time.Since(m.waitStart) >= time.Second {
				m.currentMessage = 1
				m.typewriterIndex = 0
				m.phase = phaseTyping
			}
		case phaseWaitingClear:
			if time.Since(m.waitStart) >= time.Second {
				m.phase = phaseCleared
			}
		case phaseCleared:
			return m, tea.Quit
		}
		return m, tickTextCmd()

	case tea.KeyMsg:
		// Permitir salir con "q" o Ctrl+C.
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			*m.exit = true
			return m, tea.Quit
		}
		return m, nil
	}
	return m, nil
}

// View renderiza la animación del flag y el bloque de texto.
func (m model) View() string {
	if m.phase == phaseCleared {
		return ""
	}

	// Renderiza el flag usando el estilo primaryBg.
	flagView := primaryBg.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBackground(lipgloss.Color("190")).
		BorderForeground(lipgloss.Color("0")).
		Render(m.frames[m.currentFrame])

	// Etiqueta estática "Paisabot:" usando el estilo primary.
	label := primary.Render("Paisabot:")

	// Efecto typewriter para el mensaje actual.
	var dynamicText string
	if m.typewriterIndex <= len(m.messages[m.currentMessage]) {
		dynamicText = m.messages[m.currentMessage][:m.typewriterIndex]
	} else {
		dynamicText = m.messages[m.currentMessage]
	}

	// Une la etiqueta y el texto dinámico en un bloque vertical.
	textBlock := lipgloss.JoinVertical(lipgloss.Left, label, "  "+dynamicText)
	// Combina el flag y el bloque de texto horizontalmente.
	combined := lipgloss.JoinHorizontal(lipgloss.Top, flagView+"\n", "\n  "+textBlock)

	return combined
}

// tickFlagCmd programa el siguiente tick para el flag (120ms).
func tickFlagCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return flagTickMsg(t)
	})
}

// tickTextCmd programa el siguiente tick para el efecto typewriter (40ms).
func tickTextCmd() tea.Cmd {
	return tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg {
		return textTickMsg(t)
	})
}
