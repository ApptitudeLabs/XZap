package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"xclean/tui/components"
	"xclean/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// CacheDir represents a cache directory
type CacheDir struct {
	Name string
	Path string
}

var cacheDirs = []CacheDir{
	{Name: "DerivedData", Path: "~/Library/Developer/Xcode/DerivedData"},
	{Name: "Archives", Path: "~/Library/Developer/Xcode/Archives"},
	{Name: "ModuleCache", Path: "~/Library/Developer/Xcode/ModuleCache.noindex"},
	{Name: "SwiftPM Cache", Path: "~/Library/Caches/org.swift.swiftpm"},
	{Name: "iOS DeviceSupport", Path: "~/Library/Developer/Xcode/iOS DeviceSupport"},
	{Name: "watchOS DeviceSupport", Path: "~/Library/Developer/Xcode/watchOS DeviceSupport"},
	{Name: "Products", Path: "~/Library/Developer/Xcode/Products"},
	{Name: "Xcode Cache", Path: "~/Library/Caches/com.apple.dt.Xcode"},
}

// CachesModel is the model for the caches view
type CachesModel struct {
	list     components.List
	confirm  components.Confirm
	spinner  spinner.Model
	loading  bool
	cleaning bool
	width    int
	height   int
	err      error
	message  string
}

// CachesLoadedMsg is sent when cache sizes are calculated
type CachesLoadedMsg struct {
	Items []components.ListItem
	Err   error
}

// CachesCleanedMsg is sent when cleaning completes
type CachesCleanedMsg struct {
	Count int
	Size  int64
	Err   error
}

// NewCachesModel creates a new caches view model
func NewCachesModel() CachesModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4"))

	return CachesModel{
		list:    components.NewList(nil, true, true),
		confirm: components.NewDangerConfirm("Clean Caches", ""),
		spinner: s,
		loading: true,
	}
}

// Init initializes the model
func (m CachesModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, LoadCaches)
}

// LoadCaches loads cache directory sizes asynchronously
func LoadCaches() tea.Msg {
	var items []components.ListItem

	for _, cache := range cacheDirs {
		fullPath := utils.ExpandPath(cache.Path)
		size := utils.CalculateDirSize(fullPath)

		// Check if directory exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			size = 0
		}

		items = append(items, components.ListItem{
			ID:         cache.Path,
			Title:      cache.Name,
			Subtitle:   cache.Path,
			Size:       size,
			IsCritical: size > 5<<30, // > 5GB is critical
		})
	}

	return CachesLoadedMsg{Items: items}
}

// cleanSelectedCaches cleans selected cache directories
func cleanSelectedCaches(items []components.ListItem) tea.Cmd {
	return func() tea.Msg {
		var totalSize int64
		count := 0

		for _, item := range items {
			if item.Selected {
				fullPath := utils.ExpandPath(item.ID)

				// Read directory contents and delete each item
				entries, err := os.ReadDir(fullPath)
				if err != nil {
					continue
				}

				for _, entry := range entries {
					fp := filepath.Join(fullPath, entry.Name())
					os.RemoveAll(fp)
				}

				totalSize += item.Size
				count++
			}
		}

		return CachesCleanedMsg{Count: count, Size: totalSize}
	}
}

// SetSize sets the view dimensions
func (m *CachesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-8)
	m.confirm.SetSize(50)
}

// Update handles messages
func (m CachesModel) Update(msg tea.Msg) (CachesModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case CachesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.list = components.NewList(msg.Items, true, true)
			m.list.SetSize(m.width, m.height-8)
		}
		return m, nil

	case CachesCleanedMsg:
		m.cleaning = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.message = fmt.Sprintf("Cleaned %d cache(s), freed %.2f GB", msg.Count, float64(msg.Size)/(1<<30))
			// Reload the list
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, LoadCaches)
		}
		return m, nil

	case components.ConfirmResult:
		if msg.Confirmed && msg.ID == "clean-caches" {
			m.cleaning = true
			m.message = ""
			return m, cleanSelectedCaches(m.list.Items)
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading || m.cleaning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// Handle confirmation dialog first if visible
		if m.confirm.IsVisible() {
			cmd := m.confirm.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			// Show confirmation if items are selected
			if m.list.SelectedCount() > 0 {
				selected := m.list.SelectedItems()
				var totalSize int64
				var names []string
				for _, item := range selected {
					totalSize += item.Size
					names = append(names, item.Title)
				}
				m.confirm = components.NewDangerConfirm(
					"Clean Caches",
					fmt.Sprintf("Clean %s and free %.2f GB?",
						strings.Join(names, ", "), float64(totalSize)/(1<<30)),
				)
				m.confirm.ID = "clean-caches"
				m.confirm.SetSize(55)
				m.confirm.Show()
				return m, nil
			}
		case "r":
			// Refresh
			m.loading = true
			m.message = ""
			return m, tea.Batch(m.spinner.Tick, LoadCaches)
		default:
			m.list.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the caches view
func (m CachesModel) View() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#8aadf4")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("Xcode Caches"))
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		b.WriteString(m.spinner.View())
		b.WriteString(" Calculating cache sizes...")
		return b.String()
	}

	// Cleaning state
	if m.cleaning {
		b.WriteString(m.spinner.View())
		b.WriteString(" Cleaning caches...")
		return b.String()
	}

	// Error state
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796"))
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		return b.String()
	}

	// Status message
	if m.message != "" {
		msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95"))
		b.WriteString(msgStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	// List
	b.WriteString(m.list.View())

	// Path info for current item
	if item := m.list.CurrentItem(); item != nil {
		b.WriteString("\n\n")
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d"))
		b.WriteString(pathStyle.Render(item.ID))
	}

	// Confirmation overlay
	if m.confirm.IsVisible() {
		return m.confirm.CenterInView(m.width, m.height)
	}

	return b.String()
}

// Footer returns the footer/status line for this view
func (m CachesModel) Footer() string {
	if m.loading || m.cleaning {
		return ""
	}

	if m.confirm.IsVisible() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render(
			"<-/-> select  enter confirm  esc cancel  y yes  n no",
		)
	}

	var parts []string

	// Selection info
	selectedCount := m.list.SelectedCount()
	if selectedCount > 0 {
		selectedSize := m.list.SelectedSize()
		parts = append(parts, lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eed49f")).
			Bold(true).
			Render(fmt.Sprintf("Selected: %d (%.2f GB)", selectedCount, float64(selectedSize)/(1<<30))))
	}

	// Total info
	totalSize := m.list.TotalSize()
	parts = append(parts, lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8aadf4")).
		Render(fmt.Sprintf("Total: %.2f GB", float64(totalSize)/(1<<30))))

	summary := strings.Join(parts, "  |  ")

	// Help
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render(
		"j/k navigate  space select  A all  d clear  enter clean  r refresh",
	)

	return summary + "\n" + help
}
