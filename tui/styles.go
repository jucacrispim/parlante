// Copyright 2025 Juca Crispim <juca@poraodojuca.net>

// This file is part of parlante.

// parlante is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// parlante is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with parlante. If not, see <http://www.gnu.org/licenses/>.

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
	Dark:  "#dcdcdc",
}).Align(lipgloss.Left)

var helpDescStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#B2B2B2",
	Dark:  "#8b8682",
}).Align(lipgloss.Left)

var titleStyle = list.DefaultStyles().Title

var highlightTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
	Light: "#b58900",
	Dark:  "#ffcc00",
}).Background(lipgloss.AdaptiveColor{
	Light: "#e0ecff",
	Dark:  "#5f5fd7",
})
