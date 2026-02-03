package tui

import (
	"fmt"
	"os"
	"strings"

	"xclean/cmd"
	"xclean/tui/views"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
)

// Tab indices
const (
	TabSimulators = iota
	TabCaches
	TabRuntimes
)

var tabs = []string{"Simulators", "Caches", "Runtimes"}

// Model is the main TUI model
type Model struct {
	activeTab  int
	simulators views.SimulatorsModel
	caches     views.CachesModel
	runtimes   views.RuntimesModel
	width      int
	height     int
	ready      bool
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

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.simulators.Init(),
		m.caches.Init(),
		m.runtimes.Init(),
	)
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
	banner := figure.NewFigure("xclean", "slant", true)
	// Catppuccin Macchiato colors
	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8aadf4")) // Blue

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6e738d")) // Overlay0

	byLineStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a5adcb")) // Subtext0

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8087a2")). // Overlay1
		Italic(true)

	return bannerStyle.Render(banner.String()) + versionStyle.Render("  v"+cmd.Version) + "\n" +
		byLineStyle.Render("        from Apptitude Labs") + "\n\n" +
		subtitleStyle.Render("  The fastest way to clean your Xcode workspace") + "\n"
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
