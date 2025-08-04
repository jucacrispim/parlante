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
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHeader(t *testing.T) {
	h := NewHeader()
	cmd := h.Init()
	if cmd != nil {
		t.Fatal("header Init should return nil")
	}
	msg := tea.WindowSizeMsg{Width: 80}
	h.Update(msg)
	view := h.View()
	if !strings.Contains(view, "Parlante TUI") {
		t.Fatalf("missing header title")
	}
	if h.width != 80 {
		t.Fatalf("bad width %d", h.width)
	}
}

func TestCustomKeyMapList(t *testing.T) {

	var tests = []struct {
		testName string
		msg      tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test items render",
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, "item 1") {
					t.Fatalf("item missing from list view")
				}
			},
		},
		{
			"test short help",
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, MESSAGE_KEY_HELP_QUIT) ||
					strings.Contains(view, MESSAGE_KEY_HELP_PREV_SCREEN) {
					t.Fatalf("bad short help %s", view)
				}
			},
		},
		{
			"test full help",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, MESSAGE_KEY_HELP_QUIT) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_PREV_SCREEN) {
					t.Fatalf("bad full help %s", view)
				}
			},
		},
		{
			"test without help",
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				nm := m.(CustomKeyMapList)
				nm.SetShowHelp(false)
				view := nm.View()
				if strings.Contains(view, "quit") ||
					strings.Contains(view, "previous screen") {
					t.Fatalf("bad no help %s", view)
				}
			},
		},

		{
			"test base actions",
			tea.KeyMsg{Type: tea.KeyDown},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(CustomKeyMapList)
				if !ok {
					t.Fatalf("bad model for base actions update")
				}
				if nm.Cursor() != 1 {
					t.Fatalf("Cursor did not move on KeyDown %d", nm.Cursor())
				}
			},
		},
		{
			"test Init",
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				if m.Init() != nil {
					t.Fatalf("Init should return nil")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			items := []list.Item{
				listItem("item 1"),
				listItem("item 2"),
			}
			keys := DefaultAddRemoveItemListKeyMap()
			opts := ListOpts{
				Title:         "Test",
				ShowHelp:      true,
				ShowStatusBar: true,
			}
			m := NewCustomKeyMapList(opts, items, &keys)
			nm, cmd := m.Update(test.msg)
			test.checkFn(nm, cmd)
		})
	}

}

func TestAddRemoveItemScreen(t *testing.T) {

	var tests = []struct {
		testName string
		msg      tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test load items",
			ItemListMsg{
				Items: []list.Item{
					listItem("item 1"),
					listItem("item 2"),
				},
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad model for AddRemoveItemScreen loading items")
				}
				if len(nm.List.Items()) != 2 {
					t.Fatalf("bad load items %d", len(nm.List.Items()))
				}

				view := nm.View()
				if !strings.Contains(view, "item 1") ||
					!strings.Contains(view, "item 2") {
					t.Fatalf("Item missing in view")
				}
			},
		},
		{
			"test load items with error",
			ItemListMsg{
				Items: []list.Item{},
				Err:   errors.New("Error loading"),
			},
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, "Error loading") {
					t.Fatalf("Error missing in view")
				}
			},
		},
		{
			"test get add screen",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(testAddScreen)
				if !ok {
					t.Fatalf("Bad model for AddRemoveItemScreen GetAddScreen")
				}
			},
		},
		{
			"test get remove screen",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(testRemoveScreen)
				if !ok {
					t.Fatalf("Bad model for AddRemoveItemScreen GetRemoveScreen")
				}
			},
		},
		{
			"test get previous screen",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(testPreviousScreen)
				if !ok {
					t.Fatalf(
						"Bad model for AddRemoveItemScreen GetPreviousScreen")
				}
			},
		},
		{
			"test short help with items",
			ItemListMsg{
				Items: []list.Item{
					listItem("item 1"),
					listItem("item 2"),
				},
			},
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, "d remove") {
					t.Fatalf("remove key not in help with items %s", view)
				}
			},
		},
		{
			"test full help",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}},
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, MESSAGE_KEY_HELP_PREV_SCREEN) {
					t.Fatalf("bad full help %s", view)
				}
			},
		},
		{
			"test short help without items",
			ItemListMsg{
				Items: []list.Item{},
			},
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if strings.Contains(view, MESSAGE_KEY_HELP_REMOVE) {
					t.Fatalf("remove key present in help without items")
				}
			},
		},
		{
			"test get help key",
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				nm, _ := m.(AddRemoveItemScreen)
				k := nm.GetHelpKey()

				if k.Keys()[0] != "?" {
					t.Fatalf("bad help key")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			screen := newTestScreen()
			m, cmd := screen.Update(test.msg)
			test.checkFn(m, cmd)
		})
	}
}

func TestRemoveItemScreenFilter(t *testing.T) {
	var tests = []struct {
		testName string
		msg      tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test action while filtering",
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad model for AddRemoveItemScreen filtering")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			screen := newTestScreen()
			screen.List.SetFilterState(list.Filtering)
			m, cmd := screen.Update(test.msg)
			test.checkFn(m, cmd)
		})
	}
}

func TestConfirmCancelKeyMap(t *testing.T) {
	kb := NewConfirmCancelKeyMap()

	if !kb.Confirm.Enabled() || !kb.Cancel.Enabled() {
		t.Fatalf("Keys not enable")
	}

	short := kb.ShortHelp()
	expected_short := []key.Binding{kb.Confirm, kb.Cancel}
	if !reflect.DeepEqual(short, expected_short) {
		t.Fatalf("bad short help %+v", short)
	}
	long := kb.FullHelp()
	expected_long := [][]key.Binding{{kb.Confirm, kb.Cancel}}
	if !reflect.DeepEqual(long, expected_long) {
		t.Fatalf("bad long help %+v", short)
	}

}

func TestHackHeader(t *testing.T) {
	_ = os.Setenv("INSIDE_EMACS", "vterm")
	hacked := hackHeader("hello")
	if !strings.HasPrefix(hacked, "\n") {
		t.Error("expected newline added in vterm")
	}
	_ = os.Unsetenv("INSIDE_EMACS")
	hacked = hackHeader("hello")
	if strings.HasPrefix(hacked, "\n") {
		t.Error("did not expect newline outside vterm")
	}
}

func newTestScreen() AddRemoveItemScreen {
	header := NewHeader()
	opts := ListOpts{
		Title:    "Test",
		ShowHelp: true,
	}
	load := func() tea.Cmd {
		return func() tea.Msg {
			return nil
		}
	}
	return NewAddRemoveItemScreen(&header, opts, testNav{}, load)
}

type listItem string

func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return string(i) }

type testNav struct {
}

func (m testNav) GetAddScreen() tea.Model             { return testAddScreen{} }
func (m testNav) GetRemoveScreen(list.Item) tea.Model { return testRemoveScreen{} }
func (m testNav) GetPreviousScreen() tea.Model        { return testPreviousScreen{} }

type testAddScreen struct {
}

func (s testAddScreen) Init() tea.Cmd {
	return nil
}

func (s testAddScreen) Update(tea.Msg) (m tea.Model, cmd tea.Cmd) {
	return nil, nil
}

func (s testAddScreen) View() string {
	return ""
}

type testRemoveScreen struct {
	testAddScreen
}

type testPreviousScreen struct {
	testAddScreen
}
