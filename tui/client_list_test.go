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
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestClientItem(t *testing.T) {
	c := parlante.NewClientStorageInMemory()
	client, _, _ := c.CreateClient("test client")
	defer func() {
		c.RemoveClient(client.UUID)
	}()

	item := clientItem{client: client}

	if item.Title() != client.Name {
		t.Fatalf("Bad title for item %s", item.Title())
	}

	if item.Description() != fmt.Sprintf("uuid: %s", client.UUID) {
		t.Fatalf("Bad description for item %s", item.Description())
	}

	if item.FilterValue() != client.Name {
		t.Fatalf("Bad filter value for item %s", item.FilterValue())
	}
}

func TestClientListScreen(t *testing.T) {

	c := parlante.NewClientStorageInMemory()
	cd := parlante.NewClientDomainStorageInMemory()
	comm := parlante.NewCommentStorageInMemory()
	main := newMainScreen(&c, &cd, &comm)

	c1, _, _ := c.CreateClient("a client")
	c2, _, _ := c.CreateClient("another client")
	defer func() {
		c.RemoveClient(c1.UUID)
		c.RemoveClient(c2.UUID)

	}()

	var tests = []struct {
		testName string
		screenFn func() AddRemoveItemScreen
		msgFn    func(AddRemoveItemScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test load clients",
			func() AddRemoveItemScreen {
				return newClientListScreen(main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model loading clients")
				}
				view := nm.View()
				if !strings.Contains(view, c1.Name) ||
					!strings.Contains(view, c2.Name) {
					t.Fatalf("clients not loaded")
				}

			},
		},
		{
			"test load clients with error",
			func() AddRemoveItemScreen {
				c.ForceListError(true)
				return newClientListScreen(main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				c.ForceListError(false)
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model loading clients")
				}
				if nm.err == nil {
					t.Fatalf("No error with load client error")
				}
			},
		},
		{
			"test GetAddScreen",
			func() AddRemoveItemScreen {
				return newClientListScreen(main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(addClientScreen)
				if !ok {
					t.Fatalf("bad model for add client")
				}

			},
		},
		{
			"test GetRemoveScreen",
			func() AddRemoveItemScreen {
				s := newClientListScreen(main)
				items := s.Init()()
				i := items.(ItemListMsg)
				s.List.SetItems(i.Items)
				s.List.CursorDown()
				return s
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(removeClientScreen)
				if !ok {
					t.Fatalf("bad model for remove client")
				}

				if nm.client.UUID != c2.UUID {
					t.Fatalf("bad client on remove")
				}

			},
		},
		{
			"test GetPreviousScreen",
			func() AddRemoveItemScreen {
				return newClientListScreen(main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(mainScreen)
				if !ok {
					t.Fatalf("bad model for previous screen")
				}

			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			screen := test.screenFn()
			msg := test.msgFn(screen)
			m, cmd := screen.Update(msg)
			test.checkFn(m, cmd)
		})
	}
}
