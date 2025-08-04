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

func TestRemoveDomainScreen(t *testing.T) {
	cs := parlante.NewClientStorageInMemory()
	ds := parlante.NewClientDomainStorageInMemory()
	main := newMainScreen(&cs, &ds, nil)

	client, _, _ := cs.CreateClient("client")
	ds.AddClientDomain(client, "to-be-removed")
	domain, _ := ds.GetClientDomain(client, "to-be-removed")

	tests := []struct {
		testName string
		screenFn func() removeDomainScreen
		msgFn    func(removeDomainScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"remove domain successfully",
			func() removeDomainScreen {
				return newRemoveDomainScreen(&main, &domain)
			},
			func(m removeDomainScreen) tea.Msg {
				return m.removeDomain()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("expected AddRemoveItemScreen, got %T", m)
				}
				d, _ := ds.GetClientDomain(*domain.Client, domain.Domain)
				zeroDomain := parlante.ClientDomain{}
				if d != zeroDomain {
					t.Fatal("domain was not removed")
				}
			},
		},
		{
			"remove domain with error",
			func() removeDomainScreen {
				return newRemoveDomainScreen(&main, &domain)
			},
			func(m removeDomainScreen) tea.Msg {
				ds.ForceRemoveError(true)
				return m.removeDomain()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				ds.ForceRemoveError(false)
				nm, ok := m.(removeDomainScreen)
				if !ok {
					t.Fatalf("expected removeDomainScreen, got %T", m)
				}
				if nm.err == nil {
					t.Fatal("expected error to be set")
				}
			},
		},
		{
			"confirm domain removal via enter",
			func() removeDomainScreen {
				return newRemoveDomainScreen(&main, &domain)
			},
			func(m removeDomainScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(removeDomainScreen)
				if !ok {
					t.Fatalf("expected removeDomainScreen, got %T", m)
				}
				msg := cmd()
				_, ok = msg.(removeDomainMsg)
				if !ok {
					t.Fatalf("expected removeDomainMsg, got %T", msg)
				}
			},
		},
		{
			"cancel domain removal via esc",
			func() removeDomainScreen {
				return newRemoveDomainScreen(&main, &domain)
			},
			func(m removeDomainScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("expected AddRemoveItemScreen, got %T", m)
				}
			},
		},
		{
			"render view without error",
			func() removeDomainScreen {
				return newRemoveDomainScreen(&main, &domain)
			},
			func(m removeDomainScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, _ tea.Cmd) {
				nm, ok := m.(removeDomainScreen)
				if !ok {
					t.Fatalf("expected removeDomainScreen, got %T", m)
				}
				view := nm.View()
				data := make(map[string]any)
				data["domain"] = domain.Domain
				expected_msg := parlante.Tprintf(
					MESSAGE_REMOVE_DOMAIN_CONFIRM, data)

				if !strings.Contains(view, MESSAGE_REMOVE_DOMAIN) ||
					!strings.Contains(view, expected_msg) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CONFIRM) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CANCEL) {
					t.Fatalf("view missing expected content: %s", view)
				}
			},
		},
		{
			"render view with error",
			func() removeDomainScreen {
				s := newRemoveDomainScreen(&main, &domain)
				s.err = errors.New("failed to remove domain")
				return s
			},
			func(m removeDomainScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, _ tea.Cmd) {
				nm, ok := m.(removeDomainScreen)
				if !ok {
					t.Fatalf("expected removeDomainScreen, got %T", m)
				}
				view := nm.View()
				if !strings.Contains(view, nm.err.Error()) {
					t.Fatalf("expected error message in view, got: %s", view)
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
