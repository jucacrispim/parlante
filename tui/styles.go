package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

const INNER_HEIGHT = 20
const INNER_WIDTH = 50

var headerStyle = lipgloss.NewStyle().Background(
	lipgloss.AdaptiveColor{
		Light: "#C0C0C0",
		Dark:  "#C0C0C0",
	},
).Foreground(
	lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#000000",
	},
).Align(lipgloss.Center)

var defaultTextStyle = lipgloss.NewStyle().Foreground(
	lipgloss.AdaptiveColor{
		Light: "#2A2A2A",
		Dark:  "#E5E5E5",
	},
)

var helpViewStyle = lipgloss.NewStyle().Padding(1, 0, 0, 2)
var helpKeyStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#909090",
	Dark:  "#626262",
}).Align(lipgloss.Left)

var helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#B2B2B2",
	Dark:  "#4A4A4A",
}).Align(lipgloss.Left)

var titleStyle = list.DefaultStyles().Title

var highlightTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#b58900",
	Dark:  "#ffcc00",
}).Background(lipgloss.AdaptiveColor{
	Light: "#e0ecff",
	Dark:  "#5f5fd7",
})
