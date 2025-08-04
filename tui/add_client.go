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
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type addClientMsg struct {
	client parlante.Client
	err    error
}

type addClientScreen struct {
	mainScreen    mainScreen
	textinput     textinput.Model
	clientStorage parlante.ClientStorage
	success       bool
	err           error
	keys          ConfirmCancelKeyMap
	help          help.Model
}

func (m addClientScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m addClientScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mainScreen.header.Update(msg)
	switch msg := msg.(type) {
	case addClientMsg:
		m.err = msg.err
		if m.err != nil {
			return m, nil
		}
		model := newClientListScreen(m.mainScreen)
		return model, model.Init()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			return m, m.addClient()

		case key.Matches(msg, m.keys.Cancel):
			model := newClientListScreen(m.mainScreen)
			return model, model.Init()
		}

	}
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m addClientScreen) View() string {
	s := m.mainScreen.header.View()
	title := "  " + titleStyle.Render(MESSAGE_ADD_CLIENT)
	s += title + "\n\n"
	var content string
	if m.err != nil {
		content = m.err.Error()
	} else {
		content = m.textinput.View()
	}

	s += content + "\n\n"
	lines := strings.Split(s, "\n")
	rest := m.mainScreen.list.Height() - len(lines) + 2

	helpView := m.help.View(m.keys)
	s += strings.Repeat("\n", rest) + helpViewStyle.Render(helpView)
	return s
}

func (m addClientScreen) addClient() tea.Cmd {
	return func() tea.Msg {
		client, _, err := m.clientStorage.CreateClient(m.textinput.Value())
		msg := addClientMsg{
			client: client,
			err:    err,
		}
		return msg
	}
}

func newAddClientScreen(mainScreen mainScreen) addClientScreen {
	ti := textinput.New()
	ti.Width = 20
	ti.Placeholder = MESSAGE_CLIENT_NAME
	ti.TextStyle = defaultTextStyle
	ti.PromptStyle = defaultTextStyle
	ti.Focus()
	m := addClientScreen{
		mainScreen:    mainScreen,
		textinput:     ti,
		clientStorage: mainScreen.clientStorage,
		keys:          NewConfirmCancelKeyMap(),
		help:          createHelp(),
	}
	return m
}
