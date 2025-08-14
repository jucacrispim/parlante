// Copyright 2025 Juca Crispim <juca@poraodojuca.dev>

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
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestChooseDomainKeyMap_ShortHelp(t *testing.T) {
	k := newChooseDomainKeyMap()

	if !k.Confirm.Enabled() || !k.Cancel.Enabled() {
		t.Fatalf("Confirm or Cancel keys not enabled")
	}

	short := k.ShortHelp()
	expectedShort := []key.Binding{
		k.CursorUp,
		k.CursorDown,
		k.Confirm,
		k.Cancel,
	}
	if !reflect.DeepEqual(short, expectedShort) {
		t.Fatalf("bad short help\n got: %+v\nwant: %+v", short, expectedShort)
	}
}

func TestChooseDomainKeyMap_GetHelpKey(t *testing.T) {
	k := newChooseDomainKeyMap()

	if !k.Confirm.Enabled() || !k.Cancel.Enabled() {
		t.Fatalf("Confirm or Cancel keys not enabled")
	}

	short := k.GetHelpKey()

	if !reflect.DeepEqual(short, k.ShowFullHelp) {
		t.Fatalf("bad short help\n got: %+v\nwant: %+v", short, short)
	}
}

func TestChooseDomainKeyMap_FullHelp(t *testing.T) {

	k := newChooseDomainKeyMap()

	long := k.FullHelp()

	expectedLong := [][]key.Binding{
		{
			k.CursorUp,
			k.CursorDown,
			k.NextPage,
			k.PrevPage,
			k.GoToStart,
			k.GoToEnd,
		},
		{
			k.Confirm,
			k.Cancel,
		},
		{
			k.Filter,
			k.ClearFilter,
			k.AcceptWhileFiltering,
			k.CancelWhileFiltering,
		},
		{
			k.Quit,
			k.CloseFullHelp,
		},
	}

	if !reflect.DeepEqual(long, expectedLong) {
		t.Fatalf("bad full help\n got: %+v\nwant: %+v", long, expectedLong)
	}
}

func TestAddDomainScreen(t *testing.T) {

	c := parlante.NewClientStorageInMemory()
	ds := parlante.NewClientDomainStorageInMemory()
	main := newMainScreen(&c, &ds, nil)

	c1, _, _ := c.CreateClient("a client")
	c2, _, _ := c.CreateClient("another client")
	defer func() {
		c.RemoveClient(c1.UUID)
		c.RemoveClient(c2.UUID)

	}()

	var tests = []struct {
		testName string
		screenFn func() addDomainScreen
		msgFn    func(addDomainScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test add domain select screen load clients",
			func() addDomainScreen {
				s := newAddDomainScreen(&main)
				s.textinput.SetValue("a test client")
				return s
			},
			func(m addDomainScreen) tea.Msg {
				return m.clientLoader.Load()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for add add domain select client")
				}
				view := nm.View()
				if !strings.Contains(view, c1.Name) ||
					!strings.Contains(view, c2.Name) {
					t.Fatalf("clients not loaded")
				}

			},
		},
		{
			"test add domain select screen load clients error",
			func() addDomainScreen {
				s := newAddDomainScreen(&main)
				s.textinput.SetValue("a test client")
				return s
			},
			func(m addDomainScreen) tea.Msg {
				c.ForceListError(true)
				return m.clientLoader.Load()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				c.ForceListError(false)
				nm, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for add add domain select client")
				}
				view := nm.View()
				if strings.Contains(view, c1.Name) ||
					strings.Contains(view, c2.Name) {
					t.Fatalf("clients  loaded when error")
				}

			},
		},
		{
			"test confirm select client",
			func() addDomainScreen {
				s := newAddDomainScreen(&main)
				items := s.Init()()
				i := items.(ItemListMsg)
				s.clients.SetItems(i.Items)
				s.clients.CursorDown()

				return s
			},
			func(m addDomainScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for confirm add")
				}

				if cmd == nil {
					t.Fatalf("cmd should not be nil")
				}

			},
		},

		{
			"test cancel select client",
			func() addDomainScreen {
				return newAddDomainScreen(&main)
			},
			func(m addDomainScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model for cancel add")
				}

			},
		},
		{
			"test select client View",
			func() addDomainScreen {
				m := newAddDomainScreen(&main)
				m.textinput.SetValue("Test Client Name")
				return m
			},
			func(m addDomainScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for View test %+v", m)
				}
				view := nm.View()
				if !strings.Contains(view, MESSAGE_CHOOSE_CLIENT) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CANCEL) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CONFIRM) {
					t.Fatalf("missing expected elements %s", view)
				}
			},
		},
		{
			"test add domain View",
			func() addDomainScreen {
				s := newAddDomainScreen(&main)
				s.step = addDomain
				items := s.Init()()
				i := items.(ItemListMsg)
				sel := i.Items[0].(clientItem)
				s.selectedClient = &sel.client
				s.clients.SetItems(i.Items)
				s.clients.CursorDown()

				return s
			},
			func(m addDomainScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addDomainScreen)
				nm.Update(nil)
				if !ok {
					t.Fatalf("bad model for confirm add")
				}

				if cmd == nil {
					t.Fatalf("cmd should not be nil")
				}

				view := nm.View()
				if !strings.Contains(view, nm.textinput.View()) {
					t.Fatalf("Missing expected elements %s", view)
				}

			},
		},
		{
			"test add domain  with error",
			func() addDomainScreen {
				m := newAddDomainScreen(&main)
				return m
			},
			func(m addDomainScreen) tea.Msg {
				return addDomainMsg{
					err: errors.New("bad"),
				}
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for View test")
				}
				view := nm.View()
				if !strings.Contains(view, nm.err.Error()) {
					t.Fatalf("missing expected elements %s", view)
				}
			},
		},
		{
			"test add domain ok",
			func() addDomainScreen {
				m := newAddDomainScreen(&main)
				m.textinput.SetValue("some domain")
				m.selectedClient = &c1
				m.step = addDomain

				return m
			},
			func(m addDomainScreen) tea.Msg {
				return m.addDomain()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model for View test")
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
