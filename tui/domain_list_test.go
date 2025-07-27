package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

func TestDomainItem(t *testing.T) {
	c := parlante.NewClientStorageInMemory()
	d := parlante.NewClientDomainStorageInMemory()
	client, _, _ := c.CreateClient("test client")
	domain, _ := d.AddClientDomain(client, "bla.net")
	defer func() {
		c.RemoveClient(client.UUID)
		d.RemoveClientDomain(client, domain.Domain)
	}()

	item := domainItem{domain: domain}

	if item.Title() != domain.Domain {
		t.Fatalf("Bad title for item %s", item.Title())
	}

	if item.Description() != fmt.Sprintf("client: %s", client.Name) {
		t.Fatalf("Bad description for item %s", item.Description())
	}

	if item.FilterValue() != domain.Domain {
		t.Fatalf("Bad filter value for item %s", item.FilterValue())
	}
}

func TestDomainListScreen(t *testing.T) {

	c := parlante.NewClientStorageInMemory()
	cd := parlante.NewClientDomainStorageInMemory()
	comm := parlante.NewCommentStorageInMemory()
	main := newMainScreen(&c, &cd, &comm)

	c1, _, _ := c.CreateClient("a client")
	d1, _ := cd.AddClientDomain(c1, "bla.net")
	d2, _ := cd.AddClientDomain(c1, "ble.net")

	defer func() {
		c.RemoveClient(c1.UUID)
		cd.RemoveClientDomain(c1, d1.Domain)
		cd.RemoveClientDomain(c1, d2.Domain)

	}()

	var tests = []struct {
		testName string
		screenFn func() AddRemoveItemScreen
		msgFn    func(AddRemoveItemScreen) tea.Msg
		checkFn  func(tea.Model, tea.Cmd)
	}{
		{
			"test load domains",
			func() AddRemoveItemScreen {
				return newDomainListScreen(&main)
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
				if !strings.Contains(view, d1.Domain) ||
					!strings.Contains(view, d2.Domain) {
					t.Fatalf("clients not loaded")
				}

			},
		},
		{
			"test load domains with error",
			func() AddRemoveItemScreen {
				cd.ForceListError(true)
				return newDomainListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return m.Init()()
			},
			func(m tea.Model, cmd tea.Cmd) {
				cd.ForceListError(false)
				nm, ok := m.(AddRemoveItemScreen)
				if !ok {
					t.Fatalf("bad model loading domains")
				}
				if nm.err == nil {
					t.Fatalf("No error with load domain error")
				}
			},
		},
		{
			"test GetAddScreen",
			func() AddRemoveItemScreen {
				return newDomainListScreen(&main)
			},
			func(m AddRemoveItemScreen) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
			},
			func(m tea.Model, cmd tea.Cmd) {
				_, ok := m.(addDomainScreen)
				if !ok {
					t.Fatalf("bad model for add domain")
				}

			},
		},
		{
			"test GetRemoveScreen",
			func() AddRemoveItemScreen {
				s := newDomainListScreen(&main)
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
				nm, ok := m.(removeDomainScreen)
				if !ok {
					t.Fatalf("bad model for remove client")
				}

				if nm.domain.ID != d2.ID {
					t.Fatalf("bad client on remove")
				}

			},
		},
		{
			"test GetPreviousScreen",
			func() AddRemoveItemScreen {
				return newDomainListScreen(&main)
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
