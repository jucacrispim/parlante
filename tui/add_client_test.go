package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestAddClientScreen(t *testing.T) {

	cs := parlante.NewClientStorageInMemory()
	main := newMainScreen(cs, nil, nil)

	var tests = []struct {
		testName string
		screenFn func() addClientScreen
		msgFn    func(addClientScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test add client",
			func() addClientScreen {
				s := newAddClientScreen(main)
				s.textinput.SetValue("a test client")
				return s
			},
			func(m addClientScreen) tea.Msg {
				return m.addClient()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model for add client")
				}
			},
		},
		{
			"test confirm add",
			func() addClientScreen {
				return newAddClientScreen(main)
			},
			func(m addClientScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEnter}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(addClientScreen)
				if !ok {
					t.Fatalf("bad model for confirm add")
				}

				r := cmd()
				_, ok = r.(addClientMsg)
				if !ok {
					t.Fatalf("bad msg for confirm add")
				}
			},
		},
		{
			"test cancel add",
			func() addClientScreen {
				return newAddClientScreen(main)
			},
			func(m addClientScreen) tea.Msg {
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
			"test View",
			func() addClientScreen {
				m := newAddClientScreen(main)
				m.textinput.SetValue("Test Client Name")
				return m
			},
			func(m addClientScreen) tea.Msg {
				return nil
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addClientScreen)
				if !ok {
					t.Fatalf("bad model for View test")
				}
				view := nm.View()
				if !strings.Contains(view, MESSAGE_ADD_CLIENT) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CANCEL) ||
					!strings.Contains(view, MESSAGE_KEY_HELP_CONFIRM) {
					t.Fatalf("missing expected elements %s", view)
				}
			},
		},
		{
			"test View with error",
			func() addClientScreen {
				m := newAddClientScreen(main)
				return m
			},
			func(m addClientScreen) tea.Msg {
				return addClientMsg{
					err: errors.New("bad"),
				}
			},
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(addClientScreen)
				if !ok {
					t.Fatalf("bad model for View test")
				}
				view := nm.View()
				if !strings.Contains(view, nm.err.Error()) {
					t.Fatalf("missing expected elements %s", view)
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
