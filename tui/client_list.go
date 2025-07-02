package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type clientItem struct {
	client parlante.Client
}

func (i clientItem) Title() string { return i.client.Name }
func (i clientItem) Description() string {
	return fmt.Sprintf("uuid: %s", i.client.UUID)
}
func (i clientItem) FilterValue() string { return i.client.Name }

type ClientListNavigation struct {
	MainScreen *mainScreen
}

func (n ClientListNavigation) GetAddScreen() tea.Model {
	s := newAddClientScreen(*n.MainScreen)
	return s
}

func (n ClientListNavigation) GetRemoveScreen(item list.Item) tea.Model {
	i := item.(clientItem)
	s := newRemoveClientScreen(*n.MainScreen, i.client)
	return s
}

func (n ClientListNavigation) GetPreviousScreen() tea.Model {
	return *n.MainScreen
}

type ClientLoader struct {
	Storage parlante.ClientStorage
}

func (l ClientLoader) LoadClients() tea.Cmd {
	return func() tea.Msg {
		clients, err := l.Storage.ListClients()

		if err != nil {
			msg := ItemListMsg{
				Err: err,
			}
			return msg
		}

		items := make([]list.Item, 0)
		for _, c := range clients {
			item := clientItem{
				client: c,
			}
			items = append(items, item)
		}
		msg := ItemListMsg{
			Items: items,
			Err:   nil,
		}
		return msg
	}
}

func newClientListScreen(mainScreen mainScreen) AddRemoveItemScreen {

	nav := ClientListNavigation{
		MainScreen: &mainScreen,
	}
	l := ClientLoader{
		Storage: mainScreen.clientStorage,
	}
	h := mainScreen.header
	opts := ListOpts{
		Title:           "Clients",
		ShowDescription: true,
		ShowStatusBar:   true,
		ShowHelp:        true,
	}
	s := NewAddRemoveItemScreen(&h, opts, nav, l.LoadClients)
	return s
}
