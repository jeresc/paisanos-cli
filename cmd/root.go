package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Constants for application settings
const (
	// Animation timing
	DefaultAnimationDelay  = 40 * time.Millisecond
	MessageTransitionDelay = 2 * time.Second
	LoadingDuration        = 10 * time.Second
	SpinnerDelay           = 100 * time.Millisecond

	// ASCII art dimensions
	BoxWidth = 10
)

// ColorScheme defines custom colors used in various parts of the app
type ColorScheme struct {
	Primary    *color.Color
	PrimaryBg  *color.Color
	Foreground color.Attribute
	TitleBg    color.Attribute
}

// Config holds the application configuration
type Config struct {
	AnimationDelay time.Duration
	NoAnimation    bool
	ColorScheme    ColorScheme
}

// Default configuration
var config = Config{
	AnimationDelay: DefaultAnimationDelay,
	NoAnimation:    false,
	ColorScheme: ColorScheme{
		PrimaryBg:  color.BgRGB(255, 254, 3).Add(color.FgBlack),
		Primary:    color.RGB(255, 254, 3),
		Foreground: color.FgBlack,
		TitleBg:    color.BgCyan,
	},
}

// rootCmd represents the base command when called without subcommands
var rootCmd = &cobra.Command{
	Use:   "paisanos",
	Short: "Paisanos CLI tool – a friendly greeting tool",
	Long: `Paisanos CLI is a friendly greeting tool that displays a welcome message
with a colorful ASCII art logo. It provides a warm greeting for your system
or application users.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := displayWelcomeMessage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Define command-line flags
	rootCmd.Flags().BoolVar(&config.NoAnimation, "no-animation", false,
		"Disable text animation")
	rootCmd.Flags().DurationVar(&config.AnimationDelay, "delay",
		DefaultAnimationDelay,
		"Animation delay between characters (e.g., 50ms)")
}

// displayWelcomeMessage shows the animated welcome message
func displayWelcomeMessage() error {
	username, err := getUserName()
	if err != nil {
		return err
	}
	paisanosLogo := config.ColorScheme.PrimaryBg.Sprint(" paisanos ")

	drawPaisanosBox()

	// Position cursor for welcome message
	fmt.Print("\033[3A")  // Move up 3 lines
	fmt.Print("\033[11C") // Move right 11 columns

	// Display welcome messages with transition
	printSlowly("Bienvenido a ", config.AnimationDelay)
	fmt.Print(paisanosLogo)
	printSlowly(" "+username+".", config.AnimationDelay)
	time.Sleep(MessageTransitionDelay)

	// Display second message
	clearCurrentLine()
	flagColor := config.ColorScheme.Primary.SprintFunc()
	fmt.Printf("│ %s │ ", flagColor("█▀▀▃▃█"))
	printSlowly("¡Juntos vamos a conquistar el mundo!", config.AnimationDelay)
	time.Sleep(MessageTransitionDelay)

	clearPreviousMessages()

	fmt.Print(paisanosLogo + "  Secuencia de setup iniciada.\n")
	time.Sleep(1 * time.Second)

	fmt.Print("\n") // Move to a new line and add a blank line
	showLoadingSequence("setup", "Instalando programas", "Programas instalados correctamente.", 3,
		config.ColorScheme.TitleBg, config.ColorScheme.Foreground)

	return nil
}

func padString(str string, length int) string {
	return strings.Repeat(" ", length) + str
}

func showLoadingSequence(label, loadingMsg, completedMsg string, pad int, bgColor, fgColor color.Attribute) {
	if config.NoAnimation {
		fmt.Println("Loading complete!")
		return
	}

	spinner := []rune{
		'⠁', '⠃', '⠇', '⡇', '⣇', '⣧',
		'⣷', '⣾', '⣹', '⢹', '⠹', '⠙', '⠉',
	}

	startTime := time.Now()
	formattedLabel := padString(formatTitle(label, bgColor, fgColor), pad)

	for time.Since(startTime) < LoadingDuration {
		for _, char := range spinner {
			if time.Since(startTime) >= LoadingDuration {
				break
			}

			fmt.Printf("\r%s  %s  %c", formattedLabel, loadingMsg, char)
			time.Sleep(SpinnerDelay)
		}
	}

	clearCurrentLine()
	fmt.Printf("\r%s  %s", formattedLabel, completedMsg)
}

func drawPaisanosBox() {
	primaryColor := color.RGB(255, 254, 3).SprintFunc()

	fmt.Println("╭────────╮")
	fmt.Printf("│ %s │ %s \n", primaryColor("█▀▀▃▃▃"), primaryColor("Paisabot:"))
	fmt.Printf("│ %s │ \n", primaryColor("█▀▀▃▃█"))
	fmt.Printf("│ %s      │\n", primaryColor("█"))
	fmt.Println("╰────────╯")
}

func printSlowly(text string, delay time.Duration) {
	if config.NoAnimation {
		fmt.Print(text)
		return
	}

	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
}

func getUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error getting current user: %w", err)
	}
	return currentUser.Username, nil
}

func formatTitle(title string, bgColor, fgColor color.Attribute) string {
	return color.New(bgColor, fgColor).Sprintf(" %s ", title)
}

func clearCurrentLine() {
	fmt.Print("\r")     // Return to beginning of line
	fmt.Print("\033[K") // Clear from cursor to end of line
}

func clearPreviousMessages() {
	fmt.Print("\033[3A")
	fmt.Print("\r\n")
	clearCurrentLine()
	fmt.Print("\r\n")
	clearCurrentLine()
	fmt.Print("\r\n")
	clearCurrentLine()
	fmt.Print("\r\n")
	clearCurrentLine()
	fmt.Print("\r\n")
	clearCurrentLine()
	fmt.Print("\033[3A")
}
