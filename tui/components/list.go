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
	IsOrphaned bool
	Selected   bool
	Section    string // Group/section name for grouped lists
}

// List is a selectable list component with multi-select support
type List struct {
	Items        []ListItem
	cursor       int
	scrollOffset int
	width        int
	height       int
	focused      bool
	showSize     bool
	multiSelect  bool
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
		// Scroll adjustment is handled in View() to account for section headers
	}
}

// MoveDown moves cursor down
func (l *List) MoveDown() {
	if l.cursor < len(l.Items)-1 {
		l.cursor++
		// Scroll adjustment is handled in View() to account for section headers
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
func (l *List) View() string {
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

	// Use persisted scroll offset
	start := l.scrollOffset

	// Ensure start is valid
	if start < 0 {
		start = 0
	}

	// Count how many extra lines are needed for section headers from start to cursor
	// This ensures the cursor is always visible
	extraLinesNeeded := 0
	if l.cursor >= start {
		currentSection := ""
		for i := start; i <= l.cursor && i < len(l.Items); i++ {
			if l.Items[i].Section != "" && l.Items[i].Section != currentSection {
				if currentSection != "" {
					extraLinesNeeded++ // Extra spacing between sections
				}
				extraLinesNeeded++ // Section header
				currentSection = l.Items[i].Section
			}
		}
	}

	// Adjust visible items count to account for section headers
	effectiveVisibleItems := visibleHeight - extraLinesNeeded
	if effectiveVisibleItems < 5 {
		effectiveVisibleItems = 5
	}

	// Keep cursor in view with scrolling
	if l.cursor < start {
		start = l.cursor
	} else if l.cursor >= start+effectiveVisibleItems {
		start = l.cursor - effectiveVisibleItems + 1
	}

	// Re-validate start after adjustment
	if start < 0 {
		start = 0
	}
	if start > len(l.Items)-1 {
		start = len(l.Items) - 1
	}

	// Sync scroll offset for next render
	l.scrollOffset = start

	// Calculate end - show all remaining items from start
	end := len(l.Items)

	currentSection := ""
	linesRendered := 0
	for i := start; i < end && linesRendered < visibleHeight; i++ {
		item := l.Items[i]
		isActive := i == l.cursor && l.focused

		// Render section header if section changed
		if item.Section != "" && item.Section != currentSection {
			currentSection = item.Section
			if b.Len() > 0 {
				b.WriteString("\n") // Extra spacing between sections
				linesRendered++
				if linesRendered >= visibleHeight {
					break
				}
			}
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#cad3f5")).
				MarginBottom(0)
			// Section-specific colors
			switch item.Section {
			case "Orphaned":
				headerStyle = headerStyle.Foreground(lipgloss.Color("#f4b8e4"))
				b.WriteString(headerStyle.Render("⚠️  "+item.Section) + "\n")
			case "Critical":
				headerStyle = headerStyle.Foreground(lipgloss.Color("#ed8796"))
				b.WriteString(headerStyle.Render("💾 "+item.Section) + "\n")
			default:
				headerStyle = headerStyle.Foreground(lipgloss.Color("#a6da95"))
				b.WriteString(headerStyle.Render("   "+item.Section) + "\n")
			}
			linesRendered++
			if linesRendered >= visibleHeight {
				break
			}
		}

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
		if item.IsOrphaned {
			if isActive {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#f4b8e4")).Bold(true)
			} else {
				titleStyle = titleStyle.Foreground(lipgloss.Color("#f4b8e4"))
			}
		} else if item.IsCritical {
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
		// Pad title to fixed width for size alignment
		paddedTitle := lipgloss.NewStyle().Width(titleWidth).Render(title)
		line.WriteString(titleStyle.Render(paddedTitle))

		// Size - display right after title with small spacing
		if l.showSize {
			sizeStr := formatSizeCompact(item.Size)
			line.WriteString(" ") // Single space between title and size

			sizeStyle := lipgloss.NewStyle()
			if item.IsCritical {
				sizeStyle = sizeStyle.Foreground(lipgloss.Color("#ed8796"))
			} else {
				sizeStyle = sizeStyle.Foreground(lipgloss.Color("#a6da95"))
			}
			line.WriteString(sizeStyle.Render(sizeStr))

			// Fire emoji for >10GB
			if item.Size > 10<<30 {
				line.WriteString(" 💾")
			}
		}

		b.WriteString(line.String())
		linesRendered++
		if linesRendered < visibleHeight {
			b.WriteString("\n")
		}
	}

	// Pad to consistent height to prevent layout shifts
	for linesRendered < visibleHeight {
		if linesRendered > 0 {
			b.WriteString("\n")
		}
		linesRendered++
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
