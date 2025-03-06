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

const (
	// Animation settings
	DefaultAnimationDelay  = 40 * time.Millisecond
	MessageTransitionDelay = 3 * time.Second

	// ASCII art dimensions
	BoxWidth       = 10
	MessageLinePos = 2
)

// Configuration for the application
type Config struct {
	AnimationDelay time.Duration
	NoAnimation    bool
	ShowVersion    bool
	ColorScheme    ColorScheme
}

// ColorScheme defines colors used in the application
type ColorScheme struct {
	FlagColor      color.Attribute
	LogoBackground color.Attribute
	LogoForeground color.Attribute
}

// Global configuration with defaults
var config = Config{
	AnimationDelay: DefaultAnimationDelay,
	NoAnimation:    false,
	ShowVersion:    false,
	ColorScheme: ColorScheme{
		FlagColor:      color.FgHiGreen,
		LogoBackground: color.BgHiGreen,
		LogoForeground: color.FgBlack,
	},
}

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

// getUserName gets the current user's name in a cross-platform way
func getUserName() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error getting current user: %w", err)
	}

	username := currentUser.Username
	// Handle Windows domain usernames
	if parts := strings.Split(username, "\\"); len(parts) > 1 {
		username = parts[1]
	}

	return username, nil
}

// drawPaisanosBox draws the ASCII art box with the paisanos logo
func drawPaisanosBox() {
	flagColor := color.New(config.ColorScheme.FlagColor).SprintFunc()
	boldColorText := color.New(config.ColorScheme.FlagColor, color.Bold).SprintFunc()

	fmt.Println("╭────────╮")
	fmt.Printf("│ %s │ %s \n", flagColor("█▀▀▃▃▃"), boldColorText("Paisabot:"))
	fmt.Printf("│ %s │ \n", flagColor("█▀▀▃▃█"))
	fmt.Printf("│ %s      │\n", flagColor("▀"))
	fmt.Println("╰────────╯")
}

// formatPaisanosLogo returns the styled "paisanos" text
func formatPaisanosLogo() string {
	return color.New(
		config.ColorScheme.LogoBackground,
		config.ColorScheme.LogoForeground,
	).Sprint("paisanos")
}

// clearMessageLine clears the current message line for a replacement message
func clearMessageLine() {
	fmt.Print("\r")     // Move to beginning of line
	fmt.Print("\033[K") // Clear from cursor to end of line
}

// displayWelcomeMessage shows the welcome message with animation and transition
func displayWelcomeMessage() error {
	username, err := getUserName()
	if err != nil {
		return err
	}

	flagColor := color.New(config.ColorScheme.FlagColor).SprintFunc()

	fmt.Print("\033[H\033[2J")

	// Draw the initial box
	drawPaisanosBox()

	// Position cursor for the message
	fmt.Print("\033[3A")  // Move up 3 lines to message line
	fmt.Print("\033[11C") // Move right to position after the box edge

	// Display the welcome message
	printSlowly("Bienvenido a ", config.AnimationDelay)
	fmt.Print(formatPaisanosLogo())
	printSlowly(" "+username+".", config.AnimationDelay)

	// Wait for the transition delay
	time.Sleep(MessageTransitionDelay)

	// Clear the message line and display the new message
	clearMessageLine()
	fmt.Printf("│ %s │ ", flagColor("█▀▀▃▃█"))
	printSlowly("¡Juntos vamos a conquistar el mundo!", config.AnimationDelay)

	// Reset cursor position
	fmt.Print("\n\n\n")

	fmt.Println("Hola")

	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "paisanos",
	Short: "Paisanos CLI tool - a friendly greeting tool",
	Long: `Paisanos CLI is a friendly greeting tool that displays a welcome message
with a colorful ASCII art logo. It's designed to provide a warm welcome
to users of your system or application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := displayWelcomeMessage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Define command-line flags
	rootCmd.Flags().BoolVar(&config.NoAnimation, "no-animation", false, "Disable text animation")
	rootCmd.Flags().DurationVar(&config.AnimationDelay, "delay", DefaultAnimationDelay, "Animation delay between characters (e.g., 50ms)")
}
