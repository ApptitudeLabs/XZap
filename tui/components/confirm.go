package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmResult is sent when the user makes a choice
type ConfirmResult struct {
	Confirmed bool
	ID        string // Optional ID to identify which confirmation this is
}

// Confirm is a confirmation dialog component
type Confirm struct {
	Title      string
	Message    string
	ID         string
	focused    int // 0 = Cancel, 1 = Confirm
	width      int
	visible    bool
	isDanger   bool
	confirmTxt string
	cancelTxt  string
}

// NewConfirm creates a new confirmation dialog
func NewConfirm(title, message string) Confirm {
	return Confirm{
		Title:      title,
		Message:    message,
		focused:    0,
		visible:    false,
		isDanger:   false,
		confirmTxt: "Yes",
		cancelTxt:  "No",
	}
}

// NewDangerConfirm creates a danger-styled confirmation dialog
func NewDangerConfirm(title, message string) Confirm {
	c := NewConfirm(title, message)
	c.isDanger = true
	c.confirmTxt = "Delete"
	c.cancelTxt = "Cancel"
	return c
}

// SetSize sets dialog width
func (c *Confirm) SetSize(width int) {
	c.width = width
}

// Show makes the dialog visible
func (c *Confirm) Show() {
	c.visible = true
	c.focused = 0
}

// Hide hides the dialog
func (c *Confirm) Hide() {
	c.visible = false
}

// IsVisible returns visibility state
func (c Confirm) IsVisible() bool {
	return c.visible
}

// SetButtonText sets custom button labels
func (c *Confirm) SetButtonText(confirm, cancel string) {
	c.confirmTxt = confirm
	c.cancelTxt = cancel
}

// Update handles input
func (c *Confirm) Update(msg tea.Msg) tea.Cmd {
	if !c.visible {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			c.focused = 0
		case "right", "l":
			c.focused = 1
		case "tab":
			c.focused = (c.focused + 1) % 2
		case "enter":
			c.visible = false
			return func() tea.Msg {
				return ConfirmResult{
					Confirmed: c.focused == 1,
					ID:        c.ID,
				}
			}
		case "esc", "n":
			c.visible = false
			return func() tea.Msg {
				return ConfirmResult{
					Confirmed: false,
					ID:        c.ID,
				}
			}
		case "y":
			c.visible = false
			return func() tea.Msg {
				return ConfirmResult{
					Confirmed: true,
					ID:        c.ID,
				}
			}
		}
	}
	return nil
}

// View renders the dialog
func (c Confirm) View() string {
	if !c.visible {
		return ""
	}

	width := c.width
	if width < 40 {
		width = 50
	}

	// Styles - Catppuccin Macchiato
	borderColor := lipgloss.Color("#8aadf4") // Blue
	if c.isDanger {
		borderColor = lipgloss.Color("#ed8796") // Red
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

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cad3f5")). // Text
		MarginBottom(1)

	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cad3f5")). // Text
		Background(lipgloss.Color("#494d64")). // Surface1
		Padding(0, 2).
		MarginRight(2)

	activeButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#24273a")). // Base
		Background(lipgloss.Color("#8aadf4")). // Blue
		Padding(0, 2).
		MarginRight(2)

	dangerButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#24273a")). // Base
		Background(lipgloss.Color("#ed8796")). // Red
		Padding(0, 2).
		MarginRight(2)

	// Build content
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(c.Title))
	b.WriteString("\n\n")

	// Message
	b.WriteString(messageStyle.Render(c.Message))
	b.WriteString("\n\n")

	// Buttons
	var cancelBtn, confirmBtn string

	if c.focused == 0 {
		cancelBtn = activeButtonStyle.Render(c.cancelTxt)
	} else {
		cancelBtn = buttonStyle.Render(c.cancelTxt)
	}

	if c.focused == 1 {
		if c.isDanger {
			confirmBtn = dangerButtonStyle.Render(c.confirmTxt)
		} else {
			confirmBtn = activeButtonStyle.Render(c.confirmTxt)
		}
	} else {
		confirmBtn = buttonStyle.Render(c.confirmTxt)
	}

	b.WriteString(cancelBtn)
	b.WriteString(confirmBtn)

	return dialogStyle.Render(b.String())
}

// CenterInView returns the dialog centered in a viewport of given dimensions
func (c Confirm) CenterInView(viewWidth, viewHeight int) string {
	dialog := c.View()
	if dialog == "" {
		return ""
	}

	// Get dialog dimensions
	dialogHeight := strings.Count(dialog, "\n") + 1
	dialogWidth := c.width
	if dialogWidth < 40 {
		dialogWidth = 50
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
