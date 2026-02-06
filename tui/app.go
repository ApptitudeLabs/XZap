package tui

import (
	"fmt"
	"os"
	"strings"

	"xzap/cmd"
	"xzap/tui/components"
	"xzap/tui/views"
	"xzap/utils"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Tab indices
const (
	TabSimulators = iota
	TabCaches
	TabRuntimes
)

var tabs = []string{"Simulators", "Caches", "Runtimes"}

// XcrunCheckMsg is sent after checking for xcrun availability
type XcrunCheckMsg struct {
	Available bool
}

// Model is the main TUI model
type Model struct {
	activeTab   int
	simulators  views.SimulatorsModel
	caches      views.CachesModel
	runtimes    views.RuntimesModel
	width       int
	height      int
	ready       bool
	errorDialog components.ErrorDialog
	hasXcrun    bool
	xcrunChecked bool
}

// NewModel creates the main TUI model
func NewModel() Model {
	return Model{
		activeTab:  TabSimulators,
		simulators: views.NewSimulatorsModel(),
		caches:     views.NewCachesModel(),
		runtimes:   views.NewRuntimesModel(),
	}
}

// checkXcrun checks if xcrun is available
func checkXcrun() tea.Msg {
	return XcrunCheckMsg{Available: utils.IsXcrunAvailable()}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		checkXcrun,
		m.simulators.Init(),
		m.caches.Init(),
		m.runtimes.Init(),
	)
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case XcrunCheckMsg:
		m.xcrunChecked = true
		m.hasXcrun = msg.Available
		if !msg.Available {
			m.errorDialog = components.NewErrorDialog(
				"Xcode Command Line Tools Required",
				"xcrun was not found on your system.\n\n"+
					"Please install Xcode Command Line Tools:\n"+
					"  xcode-select --install\n\n"+
					"Press any key to exit.",
			)
			m.errorDialog.SetSize(55)
		}
		return m, nil

	case components.ErrorDismissed:
		// If xcrun is missing, quit the app
		if !m.hasXcrun {
			return m, tea.Quit
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update all view sizes
		contentHeight := m.height - 18 // Reserve for header, tabs, footer
		m.simulators.SetSize(m.width-4, contentHeight)
		m.caches.SetSize(m.width-4, contentHeight)
		m.runtimes.SetSize(m.width-4, contentHeight)

	case tea.KeyMsg:
		// Handle error dialog first
		if m.errorDialog.IsVisible() {
			cmd := m.errorDialog.Update(msg)
			return m, cmd
		}

		// Global keys
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabs)
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
			return m, nil
		case "1":
			m.activeTab = TabSimulators
			return m, nil
		case "2":
			m.activeTab = TabCaches
			return m, nil
		case "3":
			m.activeTab = TabRuntimes
			return m, nil
		}
	}

	// Route spinner ticks and loaded messages to ALL views
	switch msg.(type) {
	case spinner.TickMsg,
		views.SimulatorsLoadedMsg, views.SimulatorsDeletedMsg,
		views.CachesLoadedMsg, views.CachesCleanedMsg,
		views.RuntimesLoadedMsg, views.RuntimeDeletedMsg:
		var cmd1, cmd2, cmd3 tea.Cmd
		m.simulators, cmd1 = m.simulators.Update(msg)
		m.caches, cmd2 = m.caches.Update(msg)
		m.runtimes, cmd3 = m.runtimes.Update(msg)
		cmds = append(cmds, cmd1, cmd2, cmd3)
		return m, tea.Batch(cmds...)
	}

	// Route other messages to active view only
	switch m.activeTab {
	case TabSimulators:
		var cmd tea.Cmd
		m.simulators, cmd = m.simulators.Update(msg)
		cmds = append(cmds, cmd)
	case TabCaches:
		var cmd tea.Cmd
		m.caches, cmd = m.caches.Update(msg)
		cmds = append(cmds, cmd)
	case TabRuntimes:
		var cmd tea.Cmd
		m.runtimes, cmd = m.runtimes.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the entire TUI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Show error dialog if xcrun is not available
	if m.errorDialog.IsVisible() {
		return m.errorDialog.CenterInView(m.width, m.height)
	}

	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Content
	b.WriteString(m.renderContent())

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	// Catppuccin Macchiato colors for each letter
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color("#8aadf4"))    // X
	magenta := lipgloss.NewStyle().Foreground(lipgloss.Color("#c6a0f6")) // ZAP

	versionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6e738d"))
	byLineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a5adcb"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#cad3f5"))

	// Build colorful ASCII banner
	var b strings.Builder
	b.WriteString(blue.Render(" ██╗  ██╗") + magenta.Render("███████╗") + magenta.Render(" █████╗ ") + magenta.Render("██████╗ ") + versionStyle.Render(" v"+cmd.Version) + "\n")
	b.WriteString(blue.Render(" ╚██╗██╔╝") + magenta.Render("╚══███╔╝") + magenta.Render("██╔══██╗") + magenta.Render("██╔══██╗") + "\n")
	b.WriteString(blue.Render("  ╚███╔╝ ") + magenta.Render("  ███╔╝ ") + magenta.Render("███████║") + magenta.Render("██████╔╝") + "\n")
	b.WriteString(blue.Render("  ██╔██╗ ") + magenta.Render(" ███╔╝  ") + magenta.Render("██╔══██║") + magenta.Render("██╔═══╝ ") + "\n")
	b.WriteString(blue.Render(" ██╔╝ ██╗") + magenta.Render("███████╗") + magenta.Render("██║  ██║") + magenta.Render("██║     ") + "\n")
	b.WriteString(blue.Render(" ╚═╝  ╚═╝") + magenta.Render("╚══════╝") + magenta.Render("╚═╝  ╚═╝") + magenta.Render("╚═╝     ") + "\n")
	b.WriteString("\n")
	b.WriteString(byLineStyle.Render("            from Apptitude Labs") + "\n\n")
	b.WriteString(subtitleStyle.Render("      The Ultimate Xcode Cleaner") + "\n")

	return b.String()
}

func (m Model) renderTabs() string {
	var renderedTabs []string

	for i, tab := range tabs {
		var style lipgloss.Style
		if i == m.activeTab {
			// Catppuccin Macchiato - active tab
			style = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#24273a")). // Base
				Background(lipgloss.Color("#8aadf4")). // Blue
				Padding(0, 2)
		} else {
			// Inactive tab
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6e738d")). // Overlay0
				Padding(0, 2)
		}
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Add separator line
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#494d64")). // Surface1
		Render(strings.Repeat("─", m.width))

	return tabBar + "\n" + separator
}

func (m Model) renderContent() string {
	contentStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Height(m.height - 18). // Reserve space for header, tabs, footer
		Width(m.width - 4)

	var content string
	switch m.activeTab {
	case TabSimulators:
		content = m.simulators.View()
	case TabCaches:
		content = m.caches.View()
	case TabRuntimes:
		content = m.runtimes.View()
	}

	return contentStyle.Render(content)
}

func (m Model) renderFooter() string {
	// Separator - Catppuccin Macchiato
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#494d64")). // Surface1
		Render(strings.Repeat("─", m.width))

	// View-specific footer
	var viewFooter string
	switch m.activeTab {
	case TabSimulators:
		viewFooter = m.simulators.Footer()
	case TabCaches:
		viewFooter = m.caches.Footer()
	case TabRuntimes:
		viewFooter = m.runtimes.Footer()
	}

	// Global help
	globalHelp := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6e738d")). // Overlay0
		Render("tab switch views  1/2/3 jump to view  q quit")

	footerStyle := lipgloss.NewStyle().Padding(0, 2)

	return separator + "\n" + footerStyle.Render(viewFooter) + "\n" + footerStyle.Render(globalHelp)
}

// Run starts the TUI application
func Run() {
	p := tea.NewProgram(
		NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
