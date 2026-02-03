package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Catppuccin Macchiato Colors
	primaryColor   = lipgloss.Color("#8aadf4") // Blue
	secondaryColor = lipgloss.Color("#eed49f") // Yellow
	successColor   = lipgloss.Color("#a6da95") // Green
	dangerColor    = lipgloss.Color("#ed8796") // Red
	warningColor   = lipgloss.Color("#f5a97f") // Peach
	mutedColor     = lipgloss.Color("#6e738d") // Overlay0
	whiteColor     = lipgloss.Color("#cad3f5") // Text
	blackColor     = lipgloss.Color("#24273a") // Base
	surfaceColor   = lipgloss.Color("#363a4f") // Surface0
	lavenderColor  = lipgloss.Color("#b7bdf8") // Lavender
	mauveColor     = lipgloss.Color("#c6a0f6") // Mauve
	tealColor      = lipgloss.Color("#8bd5ca") // Teal
	subtextColor   = lipgloss.Color("#a5adcb") // Subtext0

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(blackColor).
			Background(primaryColor).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2)

	TabBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(mutedColor)

	// List item styles
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(whiteColor)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	CriticalItemStyle = lipgloss.NewStyle().
				Foreground(dangerColor).
				Bold(true)

	CriticalSelectedStyle = lipgloss.NewStyle().
				Foreground(dangerColor).
				Bold(true).
				Background(lipgloss.Color("#331111"))

	CheckedStyle = lipgloss.NewStyle().
			Foreground(successColor)

	UncheckedStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Size styles
	SizeNormalStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Width(12).
			Align(lipgloss.Right)

	SizeCriticalStyle = lipgloss.NewStyle().
				Foreground(dangerColor).
				Width(12).
				Align(lipgloss.Right)

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(whiteColor).
			MarginBottom(1)

	// Confirmation dialog styles
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2).
			Width(50)

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(warningColor).
				MarginBottom(1)

	DialogButtonStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Background(mutedColor).
				Padding(0, 2).
				MarginRight(1)

	DialogActiveButtonStyle = lipgloss.NewStyle().
				Foreground(blackColor).
				Background(primaryColor).
				Padding(0, 2).
				MarginRight(1)

	DialogDangerButtonStyle = lipgloss.NewStyle().
				Foreground(whiteColor).
				Background(dangerColor).
				Padding(0, 2).
				MarginRight(1)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(primaryColor)

	// Spinner style
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Container styles
	ViewportStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Footer/summary styles
	SummaryStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Fire emoji for critical items >10GB
	FireEmoji = lipgloss.NewStyle().
			SetString("\U0001F525")
)

// FormatSize formats bytes to human-readable GB format
func FormatSize(bytes int64) string {
	gb := float64(bytes) / (1 << 30)
	if gb >= 1.0 {
		return lipgloss.NewStyle().Render(
			lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(
				formatFloat(gb) + " GB",
			),
		)
	}
	mb := float64(bytes) / (1 << 20)
	return lipgloss.NewStyle().Render(
		lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(
			formatFloat(mb) + " MB",
		),
	)
}

func formatFloat(f float64) string {
	if f >= 10 {
		return lipgloss.NewStyle().Render(
			lipgloss.NewStyle().Render(
				intToStr(int(f)),
			),
		)
	}
	// One decimal place for smaller numbers
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	return intToStr(whole) + "." + intToStr(frac)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
