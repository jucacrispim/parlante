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
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestRemoveClientScreen(t *testing.T) {

	cs := parlante.NewClientStorageInMemory()
	main := newMainScreen(&cs, nil, nil)

	clientToRemove, _, _ := cs.CreateClient("client-to-remove")

	var tests = []struct {
		testName string
		screenFn func() removeClientScreen
		msgFn    func(removeClientScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test remove client",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				return m.removeClient()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad model for remove")
				}

				c, _ := cs.GetClientByUUID(clientToRemove.UUID)
				if c.UUID == clientToRemove.UUID {
					t.Fatal("client not removed")
				}

			},
		},
		{
			"test remove client with error",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				cs.ForceRemoveError(true)
				return m.removeClient()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				cs.ForceRemoveError(false)
				nm, ok := m.(removeClientScreen)
				if !ok {
					t.Fatalf("Bad model for remove with error")
				}

				if nm.err == nil {
					t.Fatal("no error for client remove error")
				}

			},
		},
		{
			"test confirm remove",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(removeClientScreen)
				if !ok {
					t.Fatalf("bad model for confirm remove")
				}

				r := cmd()
				_, ok = r.(removeClientMsg)
				if !ok {
					t.Fatalf("bad msg for confirm remove")
				}

			},
		},
		{
			"test remove cancel",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model fo cancel remove")
				}
			},
		},
		{
			"test View",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(removeClientScreen)
				if !ok {
					t.Fatalf("bad model for View test")
				}
				view := nm.View()
				data := make(map[string]any)
				data["name"] = clientToRemove.Name
				expected_msg := parlante.Tprintf(
					MESSAGE_REMOVE_CLIENT_CONFIRM, data)
				if !strings.Contains(view, MESSAGE_REMOVE_CLIENT) ||
					!strings.Contains(view, expected_msg) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CANCEL) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CONFIRM) {
					t.Fatalf("missing view elements %s", view)
				}
			},
		},
		{
			"test View with error",
			func() removeClientScreen {
				return newRemoveClientScreen(main, clientToRemove)
			},
			func(m removeClientScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(removeClientScreen)
				if !ok {
					t.Fatalf("bad model for View test")
				}
				nm.err = errors.New("a error happend!")
				view := nm.View()
				if !strings.Contains(view, nm.err.Error()) {
					t.Fatalf("missing view elements with error %s", view)
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
