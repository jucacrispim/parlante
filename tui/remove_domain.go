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
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type removeDomainMsg struct {
	domain *parlante.ClientDomain
	err    error
}

type removeDomainScreen struct {
	mainScreen    *mainScreen
	domainStorage parlante.ClientDomainStorage
	domain        *parlante.ClientDomain
	help          help.Model
	keys          ConfirmCancelKeyMap
	err           error
}

func (m removeDomainScreen) Init() tea.Cmd {
	return nil
}

func (m removeDomainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mainScreen.header.Update(msg)
	switch msg := msg.(type) {
	case removeDomainMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		model := newDomainListScreen(m.mainScreen)
		return model, model.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.removeDomain()

		case "esc":
			model := newDomainListScreen(m.mainScreen)
			return model, model.Init()

		}
	}

	return m, nil
}

func (m removeDomainScreen) View() string {
	s := m.mainScreen.header.View()
	title := "  " + titleStyle.Render(MESSAGE_REMOVE_DOMAIN)
	s += title + "\n\n\n"
	var content string
	if m.err != nil {
		content = m.err.Error()
	} else {
		d := make(map[string]any)
		d["domain"] = m.domain.Domain
		content = parlante.Tprintf(MESSAGE_REMOVE_DOMAIN_CONFIRM, d)
	}
	s += defaultTextStyle.Render(content)

	lines := strings.Split(s, "\n")
	rest := m.mainScreen.list.Height() - len(lines) + 2

	helpView := m.help.View(m.keys)
	s += strings.Repeat("\n", rest) + helpViewStyle.Render(helpView)

	return s
}

func (m removeDomainScreen) removeDomain() tea.Cmd {
	return func() tea.Msg {
		err := m.domainStorage.RemoveClientDomain(*m.domain.Client, m.domain.Domain)
		msg := removeDomainMsg{
			domain: m.domain,
			err:    err,
		}
		return msg
	}
}

func newRemoveDomainScreen(main *mainScreen,
	domain *parlante.ClientDomain) removeDomainScreen {
	m := removeDomainScreen{
		mainScreen:    main,
		domainStorage: main.domainStorage,
		domain:        domain,
		keys:          NewConfirmCancelKeyMap(),
		help:          createHelp(),
	}
	return m
}
