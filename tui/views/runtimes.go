package views

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"xclean/tui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// RuntimeInfo holds runtime details
type RuntimeInfo struct {
	Name       string
	Identifier string
	Version    string
}

// RuntimesModel is the model for the runtimes view
type RuntimesModel struct {
	list     components.List
	confirm  components.Confirm
	spinner  spinner.Model
	loading  bool
	deleting bool
	width    int
	height   int
	err      error
	message  string
}

// RuntimesLoadedMsg is sent when runtimes are loaded
type RuntimesLoadedMsg struct {
	Items []components.ListItem
	Err   error
}

// RuntimeDeletedMsg is sent when deletion completes
type RuntimeDeletedMsg struct {
	Name string
	Err  error
}

// NewRuntimesModel creates a new runtimes view model
func NewRuntimesModel() RuntimesModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4"))

	return RuntimesModel{
		list:    components.NewList(nil, false, false), // Single select, no size
		confirm: components.NewDangerConfirm("Delete Runtime", ""),
		spinner: s,
		loading: true,
	}
}

// Init initializes the model
func (m RuntimesModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, LoadRuntimes)
}

// LoadRuntimes loads runtime information asynchronously
func LoadRuntimes() tea.Msg {
	out, err := exec.Command("xcrun", "simctl", "list", "--json", "runtimes").Output()
	if err != nil {
		return RuntimesLoadedMsg{Err: err}
	}

	var result struct {
		Runtimes []struct {
			Name         string `json:"name"`
			Identifier   string `json:"identifier"`
			Version      string `json:"version"`
			BuildVersion string `json:"buildversion"`
			IsAvailable  bool   `json:"isAvailable"`
		} `json:"runtimes"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return RuntimesLoadedMsg{Err: err}
	}

	var items []components.ListItem
	for _, runtime := range result.Runtimes {
		if runtime.IsAvailable {
			items = append(items, components.ListItem{
				ID:       runtime.Identifier,
				Title:    fmt.Sprintf("%s (%s)", runtime.Name, runtime.Version),
				Subtitle: runtime.Identifier,
			})
		}
	}

	return RuntimesLoadedMsg{Items: items}
}

// deleteRuntime deletes a runtime (requires sudo)
func deleteRuntime(identifier, name string) tea.Cmd {
	return func() tea.Msg {
		// Get runtime folder name from identifier
		parts := strings.Split(identifier, ".")
		if len(parts) == 0 {
			return RuntimeDeletedMsg{Name: name, Err: fmt.Errorf("invalid identifier")}
		}
		runtimeFolder := parts[len(parts)-1] + ".simruntime"
		runtimePath := "/Library/Developer/CoreSimulator/Profiles/Runtimes/" + runtimeFolder

		cmd := exec.Command("sudo", "rm", "-rf", runtimePath)
		if err := cmd.Run(); err != nil {
			return RuntimeDeletedMsg{Name: name, Err: err}
		}

		// Log the deletion
		logDeletion(name)

		return RuntimeDeletedMsg{Name: name}
	}
}

func logDeletion(name string) {
	f, err := os.OpenFile(os.Getenv("HOME")+"/.xclean.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	logLine := fmt.Sprintf("%s Deleted runtime: %s\n", time.Now().Format(time.RFC3339), name)
	f.WriteString(logLine)
}

// SetSize sets the view dimensions
func (m *RuntimesModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-8)
	m.confirm.SetSize(55)
}

// Update handles messages
func (m RuntimesModel) Update(msg tea.Msg) (RuntimesModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case RuntimesLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.list = components.NewList(msg.Items, false, false)
			m.list.SetSize(m.width, m.height-8)
		}
		return m, nil

	case RuntimeDeletedMsg:
		m.deleting = false
		if msg.Err != nil {
			m.err = msg.Err
			m.message = fmt.Sprintf("Failed to delete %s: %v", msg.Name, msg.Err)
		} else {
			m.message = fmt.Sprintf("Deleted runtime: %s", msg.Name)
			// Reload the list
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, LoadRuntimes)
		}
		return m, nil

	case components.ConfirmResult:
		if msg.Confirmed && strings.HasPrefix(msg.ID, "delete-runtime:") {
			item := m.list.CurrentItem()
			if item != nil {
				m.deleting = true
				m.message = ""
				return m, deleteRuntime(item.ID, item.Title)
			}
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
		case "enter", "d":
			// Show confirmation for current item
			item := m.list.CurrentItem()
			if item != nil {
				m.confirm = components.NewDangerConfirm(
					"Delete Runtime",
					fmt.Sprintf("Delete %s?\n\nThis requires sudo permission.", item.Title),
				)
				m.confirm.ID = "delete-runtime:" + item.ID
				m.confirm.SetSize(55)
				m.confirm.Show()
				return m, nil
			}
		case "r":
			// Refresh
			m.loading = true
			m.message = ""
			return m, tea.Batch(m.spinner.Tick, LoadRuntimes)
		default:
			m.list.Update(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the runtimes view
func (m RuntimesModel) View() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#8aadf4")).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("Xcode Runtimes"))
	b.WriteString("\n\n")

	// Loading state
	if m.loading {
		b.WriteString(m.spinner.View())
		b.WriteString(" Loading runtimes...")
		return b.String()
	}

	// Deleting state
	if m.deleting {
		b.WriteString(m.spinner.View())
		b.WriteString(" Deleting runtime (sudo required)...")
		return b.String()
	}

	// Error state
	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796"))
		b.WriteString(errStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Status message
	if m.message != "" {
		var msgStyle lipgloss.Style
		if strings.HasPrefix(m.message, "Failed") {
			msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796"))
		} else {
			msgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6da95"))
		}
		b.WriteString(msgStyle.Render(m.message))
		b.WriteString("\n\n")
	}

	// List
	if len(m.list.Items) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d"))
		b.WriteString(mutedStyle.Render("No runtimes found."))
	} else {
		b.WriteString(m.list.View())
	}

	// Info for current item
	if item := m.list.CurrentItem(); item != nil {
		b.WriteString("\n\n")
		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d"))
		b.WriteString(pathStyle.Render(item.Subtitle))
	}

	// Sudo warning
	b.WriteString("\n\n")
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#f5a97f"))
	b.WriteString(warnStyle.Render("Note: Deleting runtimes requires sudo permission."))

	// Confirmation overlay
	if m.confirm.IsVisible() {
		return m.confirm.CenterInView(m.width, m.height)
	}

	return b.String()
}

// Footer returns the footer/status line for this view
func (m RuntimesModel) Footer() string {
	if m.loading || m.deleting {
		return ""
	}

	if m.confirm.IsVisible() {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render(
			"<-/-> select  enter confirm  esc cancel  y yes  n no",
		)
	}

	// Help
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d")).Render(
		fmt.Sprintf("%d runtime(s)  |  j/k navigate  enter/d delete  r refresh", len(m.list.Items)),
	)

	return help
}
