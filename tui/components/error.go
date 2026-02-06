package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorDismissed is sent when the user dismisses the error dialog
type ErrorDismissed struct{}

// ErrorDialog is an error/info dialog component with a single dismiss button
type ErrorDialog struct {
	Title   string
	Message string
	width   int
	visible bool
	isError bool // true for error styling, false for info styling
}

// NewErrorDialog creates a new error dialog
func NewErrorDialog(title, message string) ErrorDialog {
	return ErrorDialog{
		Title:   title,
		Message: message,
		visible: true,
		isError: true,
	}
}

// NewInfoDialog creates a new info dialog
func NewInfoDialog(title, message string) ErrorDialog {
	return ErrorDialog{
		Title:   title,
		Message: message,
		visible: true,
		isError: false,
	}
}

// SetSize sets dialog width
func (e *ErrorDialog) SetSize(width int) {
	e.width = width
}

// Show makes the dialog visible
func (e *ErrorDialog) Show() {
	e.visible = true
}

// Hide hides the dialog
func (e *ErrorDialog) Hide() {
	e.visible = false
}

// IsVisible returns visibility state
func (e ErrorDialog) IsVisible() bool {
	return e.visible
}

// Update handles input
func (e *ErrorDialog) Update(msg tea.Msg) tea.Cmd {
	if !e.visible {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q", " ":
			e.visible = false
			return func() tea.Msg {
				return ErrorDismissed{}
			}
		}
	}
	return nil
}

// View renders the dialog
func (e ErrorDialog) View() string {
	if !e.visible {
		return ""
	}

	width := e.width
	if width < 40 {
		width = 55
	}

	// Styles - Catppuccin Macchiato
	borderColor := lipgloss.Color("#8aadf4") // Blue for info
	if e.isError {
		borderColor = lipgloss.Color("#ed8796") // Red for error
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#eed49f")). // Yellow
		MarginBottom(1)

	if e.isError {
		titleStyle = titleStyle.Foreground(lipgloss.Color("#ed8796")) // Red
	}

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cad3f5")). // Text
		MarginBottom(1)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#24273a")). // Base
		Background(lipgloss.Color("#8aadf4")). // Blue
		Padding(0, 3)

	// Build content
	var b strings.Builder

	// Title with icon
	icon := "i"
	if e.isError {
		icon = "!"
	}
	b.WriteString(titleStyle.Render(icon + " " + e.Title))
	b.WriteString("\n\n")

	// Message
	b.WriteString(messageStyle.Render(e.Message))
	b.WriteString("\n\n")

	// Button
	b.WriteString(buttonStyle.Render("OK"))

	return dialogStyle.Render(b.String())
}

// CenterInView returns the dialog centered in a viewport of given dimensions
func (e ErrorDialog) CenterInView(viewWidth, viewHeight int) string {
	dialog := e.View()
	if dialog == "" {
		return ""
	}

	// Get dialog dimensions
	dialogHeight := strings.Count(dialog, "\n") + 1
	dialogWidth := e.width
	if dialogWidth < 40 {
		dialogWidth = 55
	}

	// Calculate padding
	topPadding := (viewHeight - dialogHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}
	leftPadding := (viewWidth - dialogWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	// Add vertical padding
	var result strings.Builder
	for i := 0; i < topPadding; i++ {
		result.WriteString("\n")
	}

	// Add horizontal padding to each line
	lines := strings.Split(dialog, "\n")
	padding := strings.Repeat(" ", leftPadding)
	for i, line := range lines {
		result.WriteString(padding)
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
