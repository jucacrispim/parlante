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
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type domainItem struct {
	domain parlante.ClientDomain
}

func (i domainItem) Title() string { return i.domain.Domain }
func (i domainItem) Description() string {
	data := make(map[string]any)
	data["clientName"] = i.domain.Client.Name
	return parlante.Tprintf(MESSAGE_DOMAIN_DESCRIPTION, data)
}
func (i domainItem) FilterValue() string { return i.domain.Domain }

type DomainListNavigation struct {
	MainScreen *mainScreen
}

func (n DomainListNavigation) GetAddScreen() tea.Model {
	s := newAddDomainScreen(n.MainScreen)
	return s
}

func (n DomainListNavigation) GetRemoveScreen(item list.Item) tea.Model {
	i := item.(domainItem)
	s := newRemoveDomainScreen(n.MainScreen, &i.domain)
	return s
}

func (n DomainListNavigation) GetPreviousScreen() tea.Model {
	return *n.MainScreen
}

type DomainLoader struct {
	Storage parlante.ClientDomainStorage
}

func (l DomainLoader) Load() tea.Cmd {
	return func() tea.Msg {
		domains, err := l.Storage.ListDomains()

		if err != nil {
			msg := ItemListMsg{
				Err: err,
			}
			return msg
		}

		items := make([]list.Item, 0)
		for _, d := range domains {
			item := domainItem{
				domain: d,
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

func newDomainListScreen(mainScreen *mainScreen) AddRemoveItemScreen {

	nav := DomainListNavigation{
		MainScreen: mainScreen,
	}
	l := DomainLoader{
		Storage: mainScreen.domainStorage,
	}
	h := mainScreen.header
	opts := ListOpts{
		Title:           MESSAGE_DOMAINS,
		ShowDescription: true,
		ShowStatusBar:   true,
		ShowHelp:        true,
	}
	s := NewAddRemoveItemScreen(&h, opts, nav, l.Load)
	return s
}
