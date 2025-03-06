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
	FlagColor          color.Attribute
	LogoBackground     color.Attribute
	LogoForeground     color.Attribute
	TitleBackground    color.Attribute
	TitleForeground    color.Attribute
	HomebrewBackground color.Attribute
	HomebrewForeground color.Attribute
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
		FlagColor:          color.FgHiGreen,
		LogoBackground:     color.BgHiGreen,
		LogoForeground:     color.FgBlack,
		TitleBackground:    color.BgHiCyan,
		TitleForeground:    color.FgBlack,
		HomebrewBackground: color.BgYellow,
		HomebrewForeground: color.FgBlack,
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

	paisanosLogo := formatTitle("paisanos", config.ColorScheme.LogoBackground, config.ColorScheme.LogoForeground)

	// Clear the screen and draw initial elements
	clearScreen()
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
	flagColor := color.New(config.ColorScheme.FlagColor).SprintFunc()
	fmt.Printf("│ %s │ ", flagColor("█▀▀▃▃█"))
	printSlowly("¡Juntos vamos a conquistar el mundo!", config.AnimationDelay)
	time.Sleep(MessageTransitionDelay)

	// Show loading sequences
	clearScreen()
	fmt.Print("\n")
	fmt.Print(paisanosLogo + "  Secuencia de setup iniciada.\n")
	time.Sleep(1 * time.Second)

	fmt.Print("\n")
	showLoadingSequence("brew", "Instalando homebrew", "Homebrew instalado correctamente.", 4,
		config.ColorScheme.HomebrewBackground, config.ColorScheme.HomebrewForeground)
	time.Sleep(2 * time.Second)

	fmt.Print("\n\n")
	showLoadingSequence("setup", "Instalando programas", "Programas instalados correctamente.", 3,
		config.ColorScheme.TitleBackground, config.ColorScheme.TitleForeground)

	fmt.Print("\n\n")
	return nil
}

func padString(str string, length int) string {
	return strings.Repeat(" ", length) + str
}

// showLoadingSequence displays an animated loading sequence
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

// drawPaisanosBox prints the ASCII art box with logo
func drawPaisanosBox() {
	flagColor := color.New(config.ColorScheme.FlagColor).SprintFunc()
	boldColorText := color.New(config.ColorScheme.FlagColor, color.Bold).SprintFunc()

	fmt.Println("╭────────╮")
	fmt.Printf("│ %s │ %s \n", flagColor("█▀▀▃▃▃"), boldColorText("Paisabot:"))
	fmt.Printf("│ %s │ \n", flagColor("█▀▀▃▃█"))
	fmt.Printf("│ %s      │\n", flagColor("█"))
	fmt.Println("╰────────╯")
}

// Helper functions

// printSlowly prints text gradually, character by character
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

// getUserName retrieves the current user's name
func getUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error getting current user: %w", err)
	}
	return currentUser.Username, nil
}

// formatTitle formats a title with the specified colors
func formatTitle(title string, bgColor, fgColor color.Attribute) string {
	return color.New(bgColor, fgColor).Sprintf(" %s ", title)
}

// clearCurrentLine clears the current terminal line
func clearCurrentLine() {
	fmt.Print("\r")     // Return to beginning of line
	fmt.Print("\033[K") // Clear from cursor to end of line
}

// clearScreen clears the entire terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
