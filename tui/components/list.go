package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in the selectable list
type ListItem struct {
	ID         string
	Title      string
	Subtitle   string
	Size       int64
	IsCritical bool
	Selected   bool
}

// List is a selectable list component with multi-select support
type List struct {
	Items       []ListItem
	cursor      int
	width       int
	height      int
	focused     bool
	showSize    bool
	multiSelect bool
}

// NewList creates a new list component
func NewList(items []ListItem, multiSelect bool, showSize bool) List {
	return List{
		Items:       items,
		cursor:      0,
		multiSelect: multiSelect,
		showSize:    showSize,
		focused:     true,
	}
}

// SetSize sets the dimensions of the list
func (l *List) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// Focus sets the focus state
func (l *List) Focus() {
	l.focused = true
}

// Blur removes focus
func (l *List) Blur() {
	l.focused = false
}

// Cursor returns current cursor position
func (l List) Cursor() int {
	return l.cursor
}

// SetCursor sets the cursor position
func (l *List) SetCursor(pos int) {
	if pos >= 0 && pos < len(l.Items) {
		l.cursor = pos
	}
}

// SelectedItems returns all selected items
func (l List) SelectedItems() []ListItem {
	var selected []ListItem
	for _, item := range l.Items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// SelectedCount returns the count of selected items
func (l List) SelectedCount() int {
	count := 0
	for _, item := range l.Items {
		if item.Selected {
			count++
		}
	}
	return count
}

// SelectedSize returns total size of selected items
func (l List) SelectedSize() int64 {
	var total int64
	for _, item := range l.Items {
		if item.Selected {
			total += item.Size
		}
	}
	return total
}

// TotalSize returns total size of all items
func (l List) TotalSize() int64 {
	var total int64
	for _, item := range l.Items {
		total += item.Size
	}
	return total
}

// Toggle toggles selection of current item
func (l *List) Toggle() {
	if len(l.Items) > 0 && l.cursor < len(l.Items) {
		l.Items[l.cursor].Selected = !l.Items[l.cursor].Selected
	}
}

// SelectAll selects all items
func (l *List) SelectAll() {
	for i := range l.Items {
		l.Items[i].Selected = true
	}
}

// SelectAllCritical selects all critical items
func (l *List) SelectAllCritical() {
	for i := range l.Items {
		if l.Items[i].IsCritical {
			l.Items[i].Selected = true
		}
	}
}

// DeselectAll deselects all items
func (l *List) DeselectAll() {
	for i := range l.Items {
		l.Items[i].Selected = false
	}
}

// MoveUp moves cursor up
func (l *List) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
	}
}

// MoveDown moves cursor down
func (l *List) MoveDown() {
	if l.cursor < len(l.Items)-1 {
		l.cursor++
	}
}

// CurrentItem returns the item at cursor
func (l List) CurrentItem() *ListItem {
	if len(l.Items) > 0 && l.cursor < len(l.Items) {
		return &l.Items[l.cursor]
	}
	return nil
}

// Update handles input
func (l *List) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			l.MoveUp()
		case "down", "j":
			l.MoveDown()
		case " ":
			if l.multiSelect {
				l.Toggle()
			}
		case "a":
			if l.multiSelect {
				l.SelectAllCritical()
			}
		case "A":
			if l.multiSelect {
				l.SelectAll()
			}
		case "d":
			if l.multiSelect {
				l.DeselectAll()
			}
		}
	}
	return nil
}

// View renders the list
func (l List) View() string {
	if len(l.Items) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render("No items")
	}

	var b strings.Builder

	// Calculate visible window - use a sensible default
	visibleHeight := l.height
	if visibleHeight < 5 {
		visibleHeight = 20 // Reasonable default
	}

	// Always limit to visible height
	start := 0
	end := len(l.Items)

	// Keep cursor in view with scrolling
	if l.cursor < start {
		start = l.cursor
	} else if l.cursor >= start+visibleHeight {
		start = l.cursor - visibleHeight + 1
	}

	// Ensure start is valid
	if start < 0 {
		start = 0
	}
	if start > len(l.Items)-visibleHeight && len(l.Items) > visibleHeight {
		start = len(l.Items) - visibleHeight
	}

	// Calculate end
	end = start + visibleHeight
	if end > len(l.Items) {
		end = len(l.Items)
	}

	for i := start; i < end; i++ {
		item := l.Items[i]
		isActive := i == l.cursor && l.focused

		// Build the line
		var line strings.Builder

		// Cursor indicator (compact) - Catppuccin Macchiato
		if isActive {
			line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4")).Render("> "))
		} else {
			line.WriteString("  ")
		}

		// Circular selection indicator
		if l.multiSelect {
			if item.Selected {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95")).Render("● "))
			} else {
				line.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render("○ "))
			}
		}

		// Title
		titleStyle := lipgloss.NewStyle()
		if item.IsCritical {
			if isActive {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#ed8796")).Bold(true)
			} else {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#ed8796"))
			}
		} else {
			if isActive {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#8aadf4")).Bold(true)
			} else {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#cad3f5"))
			}
		}

		// Calculate available width for title
		titleWidth := l.width - 25 // Reserve space for checkbox, cursor, size
		if titleWidth < 20 {
			titleWidth = 20
		}

		title := item.Title
		if len(title) > titleWidth {
			title = title[:titleWidth-3] + "..."
		}
		line.WriteString(titleStyle.Render(title))

		// Size - display right after title with small spacing
		if l.showSize {
			sizeStr := formatSizeCompact(item.Size)
			line.WriteString("  ") // Just 2 spaces between title and size

			sizeStyle := lipgloss.NewStyle()
			if item.IsCritical {
				sizeStyle = sizeStyle.Foreground(lipgloss.Color("#ed8796"))
			} else {
				sizeStyle = sizeStyle.Foreground(lipgloss.Color("#a6da95"))
			}
			line.WriteString(sizeStyle.Render(sizeStr))

			// Fire emoji for >10GB
			if item.Size > 10<<30 {
				line.WriteString(" 🔥")
			}
		}

		b.WriteString(line.String())
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func formatSizeCompact(bytes int64) string {
	gb := float64(bytes) / (1 << 30)
	if gb >= 1.0 {
		return fmt.Sprintf("%6.2f GB", gb)
	}
	mb := float64(bytes) / (1 << 20)
	return fmt.Sprintf("%6.2f MB", mb)
}
