package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestMainScreenItem(t *testing.T) {
	item := mainScreenItem{name: "the item", descr: "something"}

	if item.Title() != item.name {
		t.Fatalf("bad title for item %s", item.Title())
	}

	if item.Description() != item.descr {
		t.Fatalf("bad description for item %s", item.Description())
	}

	if item.FilterValue() != item.name {
		t.Fatalf("bad filter value for item %s", item.FilterValue())
	}
}

func TestMainScreen(t *testing.T) {

	c := parlante.NewClientStorageInMemory()
	cd := parlante.NewClientDomainStorageInMemory()
	comm := parlante.NewCommentStorageInMemory()

	var tests = []struct {
		testName string
		screenFn func() mainScreen
		msg      tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test screen instance",
			func() mainScreen {
				return newMainScreen(&c, &cd, &comm)
			},
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				nm, ok := m.(mainScreen)
				if !ok {
					t.Fatalf("bad model for init")
				}
				if nm.list.KeyMap.GoToEnd.Enabled() ||
					nm.list.KeyMap.GoToStart.Enabled() {
					t.Fatalf("Bad key enabled")
				}
			},
		},
		{
			"test help",
			func() mainScreen {
				return newMainScreen(&c, &cd, &comm)
			},
			nil,
			func(m tea.Model, cmd tea.Cmd) {
				view := m.View()
				if !strings.Contains(view, MESSAGE_KEY_HELP_SELECT) {
					t.Fatalf("missing key on help %s", view)
				}
			},
		},
		{
			"test select client",
			func() mainScreen {
				return newMainScreen(&c, &cd, &comm)
			},
			tea.KeyMsg{Type: tea.KeyEnter},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad screen for clients")
				}
				r := cmd()
				_, ok = r.(ItemListMsg)

				if !ok {
					t.Fatalf("bad load fn return for client")
				}

			},
		},
		{
			"test select domain",
			func() mainScreen {
				s := newMainScreen(&c, &cd, &comm)
				s.list.CursorDown()
				return s
			},
			tea.KeyMsg{Type: tea.KeyEnter},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad screen for domains")
				}
				r := cmd()
				_, ok = r.(ItemListMsg)

				if !ok {
					t.Fatalf("bad load fn return for domain")
				}

			},
		},
		{
			"test select comment",
			func() mainScreen {
				s := newMainScreen(&c, &cd, &comm)
				s.list.CursorDown()
				s.list.CursorDown()
				return s
			},
			tea.KeyMsg{Type: tea.KeyEnter},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("Bad screen for domains")
				}
				r := cmd()
				_, ok = r.(ItemListMsg)

				if !ok {
					t.Fatalf("bad load fn return for client")
				}

			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			m, cmd := test.screenFn().Update(test.msg)
			test.checkFn(m, cmd)
		})
	}
}
