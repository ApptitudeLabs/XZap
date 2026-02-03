package views

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"xclean/tui/components"
	"xclean/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// SimulatorsModel is the model for the simulators view
type SimulatorsModel struct {
	list      components.List
	confirm   components.Confirm
	spinner   spinner.Model
	loading   bool
	deleting  bool
	width     int
	height    int
	err       error
	message   string // Status message after actions
}

// SimulatorsLoadedMsg is sent when simulators finish loading
type SimulatorsLoadedMsg struct {
	Items []components.ListItem
	Err   error
}

// SimulatorsDeletedMsg is sent when deletion completes
type SimulatorsDeletedMsg struct {
	Count int
	Size  int64
	Err   error
}

// NewSimulatorsModel creates a new simulators view model
func NewSimulatorsModel() SimulatorsModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4"))

	return SimulatorsModel{
		list:    components.NewList(nil, true, true),
		confirm: components.NewDangerConfirm("Delete Simulators", ""),
		spinner: s,
		loading: true,
	}
}

// Init initializes the model
func (m SimulatorsModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, LoadSimulators)
}

// LoadSimulators loads simulator data asynchronously
func LoadSimulators() tea.Msg {
	base := utils.ExpandPath("~/Library/Developer/CoreSimulator/Devices")
	entries, err := os.ReadDir(base)
	if err != nil {
		return SimulatorsLoadedMsg{Err: err}
	}

	nameMap := utils.GetSimFriendlyNames()
	var items []components.ListItem

	for _, e := range entries {
		uuid := e.Name()
		dataPath := filepath.Join(base, uuid, "data")
		size := utils.CalculateDirSize(dataPath)

		// Skip very small simulators (< 10MB)
		if size < 10<<20 {
			continue
		}

		friendlyName, found := nameMap[uuid]
		if !found {
			friendlyName = uuid
		}

		isCritical := size >= 3<<30 // 3GB threshold

		items = append(items, components.ListItem{
			ID:         uuid,
			Title:      friendlyName,
			Size:       size,
			IsCritical: isCritical,
		})
	}

	// Sort by size descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Size > items[j].Size
	})

	return SimulatorsLoadedMsg{Items: items}
}

// deleteSelectedSimulators deletes selected simulators
func deleteSelectedSimulators(items []components.ListItem) tea.Cmd {
	return func() tea.Msg {
		var totalSize int64
		count := 0
		base := utils.ExpandPath("~/Library/Developer/CoreSimulator/Devices")

		for _, item := range items {
			if item.Selected {
				// Use simctl to delete
				cmd := exec.Command("xcrun", "simctl", "delete", item.ID)
				if err := cmd.Run(); err != nil {
					// If simctl fails, try direct removal
					devicePath := filepath.Join(base, item.ID)
					os.RemoveAll(devicePath)
				}
				totalSize += item.Size
				count++
			}
		}

		return SimulatorsDeletedMsg{Count: count, Size: totalSize}
	}
}

// SetSize sets the view dimensions
func (m *SimulatorsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-8) // Reserve space for header/footer
	m.confirm.SetSize(50)
}

// Update handles messages
func (m SimulatorsModel) Update(msg tea.Msg) (SimulatorsModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case SimulatorsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.list = components.NewList(msg.Items, true, true)
			m.list.SetSize(m.width, m.height-8)
		}
		return m, nil

	case SimulatorsDeletedMsg:
		m.deleting = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.message = fmt.Sprintf("Deleted %d simulator(s), freed %.2f GB", msg.Count, float64(msg.Size)/(1<<30))
			// Reload the list
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, LoadSimulators)
		}
		return m, nil

	case components.ConfirmResult:
		if msg.Confirmed && msg.ID == "delete-sims" {
			m.deleting = true
			m.message = ""
			return m, deleteSelectedSimulators(m.list.Items)
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading || m.deleting {
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
				for _, item := range selected {
					totalSize += item.Size
				}
				m.confirm = components.NewDangerConfirm(
					"Delete Simulators",
					fmt.Sprintf("Delete %d simulator(s) and free %.2f GB?",
						len(selected), float64(totalSize)/(1<<30)),
				)
				m.confirm.ID = "delete-sims"
				m.confirm.SetSize(50)
				m.confirm.Show()
				return m, nil
			}
		case "r":
			// Refresh
			m.loading = true
			m.message = ""
			return m, tea.Batch(m.spinner.Tick, LoadSimulators)
		default:
			m.list.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the simulators view
func (m SimulatorsModel) View() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#8aadf4")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("Xcode Simulators"))
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		b.WriteString(m.spinner.View())
		b.WriteString(" Calculating simulator sizes...")
		return b.String()
	}

	// Deleting state
	if m.deleting {
		b.WriteString(m.spinner.View())
		b.WriteString(" Deleting simulators...")
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
	if len(m.list.Items) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d"))
		b.WriteString(mutedStyle.Render("No simulators with significant data found."))
	} else {
		b.WriteString(m.list.View())
	}

	// Confirmation overlay
	if m.confirm.IsVisible() {
		// Clear and show dialog centered
		return m.confirm.CenterInView(m.width, m.height)
	}

	return b.String()
}

// Footer returns the footer/status line for this view
func (m SimulatorsModel) Footer() string {
	if m.loading || m.deleting {
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
		Render(fmt.Sprintf("Total: %d items (%.2f GB)", len(m.list.Items), float64(totalSize)/(1<<30))))

	summary := strings.Join(parts, "  |  ")

	// Help
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render(
		"j/k navigate  space select  a critical  A all  d clear  enter delete  r refresh",
	)

	return summary + "\n" + help
}
