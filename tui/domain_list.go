package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type domainItem struct {
	domain parlante.ClientDomain
}

func (i domainItem) Title() string { return i.domain.Domain }
func (i domainItem) Description() string {
	return fmt.Sprintf("client: %s", i.domain.Client.Name)
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
		Title:           "Domains",
		ShowDescription: true,
		ShowStatusBar:   true,
		ShowHelp:        true,
	}
	s := NewAddRemoveItemScreen(&h, opts, nav, l.Load)
	return s
}
