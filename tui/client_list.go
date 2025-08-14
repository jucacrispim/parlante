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

func (l ClientLoader) Load() tea.Cmd {
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
		Title:           MESSAGE_CLIENTS,
		ShowDescription: true,
		ShowStatusBar:   true,
		ShowHelp:        true,
	}
	s := NewAddRemoveItemScreen(&h, opts, nav, l.Load)
	return s
}
