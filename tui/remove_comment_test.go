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
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestRemoveCommentScreen(t *testing.T) {
	cs := parlante.NewClientStorageInMemory()
	ds := parlante.NewClientDomainStorageInMemory()
	cmts := parlante.NewCommentStorageInMemory()
	main := newMainScreen(&cs, &ds, &cmts)

	client, _, _ := cs.CreateClient("client")
	domain, _ := ds.AddClientDomain(client, "domain.net")

	comm, _ := cmts.CreateComment(client, domain, "zé", "comment", "http://bla.net")

	tests := []struct {
		testName string
		screenFn func() removeCommentScreen
		msgFn    func(removeCommentScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"remove comment successfully",
			func() removeCommentScreen {
				return newRemoveCommentScreen(main, comm)
			},
			func(m removeCommentScreen) tea.Msg {
				return m.removeComment()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("expected AddRemoveItemScreen, got %T", m)
				}
				c := cmts.GetComment()
				if c != (parlante.Comment{}) {
					t.Fatal("comment was not removed")
				}
			},
		},
		{
			"remove comment with error",
			func() removeCommentScreen {
				return newRemoveCommentScreen(main, comm)
			},
			func(m removeCommentScreen) tea.Msg {
				cmts.ForceRemoveError(true)
				return m.removeComment()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				cmts.ForceRemoveError(false)
				nm, ok := m.(removeCommentScreen)
				if !ok {
					t.Fatalf("expected removeCommentScreen, got %T", m)
				}
				if nm.err == nil {
					t.Fatal("expected error to be set")
				}
			},
		},
		{
			"confirm comment removal via enter",
			func() removeCommentScreen {
				return newRemoveCommentScreen(main, comm)
			},
			func(m removeCommentScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(removeCommentScreen)
				if !ok {
					t.Fatalf("expected removeCommentScreen, got %T", m)
				}
				msg := cmd()
				_, ok = msg.(removeCommentMsg)
				if !ok {
					t.Fatalf("expected removeCommentMsg, got %T", msg)
				}
			},
		},
		{
			"cancel comment removal via esc",
			func() removeCommentScreen {
				return newRemoveCommentScreen(main, comm)
			},
			func(m removeCommentScreen) tea.Msg {
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
			func() removeCommentScreen {
				return newRemoveCommentScreen(main, comm)
			},
			func(m removeCommentScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, _ tea.Cmd) {
				nm, ok := m.(removeCommentScreen)
				if !ok {
					t.Fatalf("expected removeCommentScreen, got %T", m)
				}
				view := nm.View()
				data := map[string]any{
					"name": comm.Author,
					"url":  comm.PageURL,
				}
				expected := parlante.Tprintf(MESSAGE_REMOVE_COMMENT_CONFIRM, data)

				if !strings.Contains(view, MESSAGE_REMOVE_COMMENT) ||
					!strings.Contains(view, expected) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CONFIRM) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CANCEL) {
					t.Fatalf("view missing expected content: %s", view)
				}
			},
		},
		{
			"render view with error",
			func() removeCommentScreen {
				s := newRemoveCommentScreen(main, comm)
				s.err = errors.New("failed to remove comment")
				return s
			},
			func(m removeCommentScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, _ tea.Cmd) {
				nm, ok := m.(removeCommentScreen)
				if !ok {
					t.Fatalf("expected removeCommentScreen, got %T", m)
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
