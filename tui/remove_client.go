package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type removeClientMsg struct {
	client parlante.Client
	err    error
}

type removeClientScreen struct {
	mainScreen    mainScreen
	clientStorage parlante.ClientStorage
	client        parlante.Client
	help          help.Model
	keys          ConfirmCancelKeyMap
	err           error
}

func (m removeClientScreen) Init() tea.Cmd {
	return nil
}

func (m removeClientScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mainScreen.header.Update(msg)
	switch msg := msg.(type) {
	case removeClientMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		model := newClientListScreen(m.mainScreen)
		return model, model.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.removeClient()

		case "esc":
			model := newClientListScreen(m.mainScreen)
			return model, model.Init()

		}
	}
	return m, nil
}

func (m removeClientScreen) View() string {
	s := m.mainScreen.header.View()
	title := "  " + titleStyle.Render(MESSAGE_REMOVE_CLIENT)
	s += title + "\n\n\n"
	var content string
	if m.err != nil {
		content = m.err.Error()
	} else {
		data := make(map[string]any, 0)
		data["name"] = m.client.Name
		content = fmt.Sprintf(
			parlante.Tprintf(MESSAGE_REMOVE_CLIENT_CONFIRM, data))
	}
	s += defaultTextStyle.Render(content)

	lines := strings.Split(s, "\n")
	rest := m.mainScreen.list.Height() - len(lines) + 2

	helpView := m.help.View(m.keys)
	s += strings.Repeat("\n", rest) + helpViewStyle.Render(helpView)

	return s
}

func (m removeClientScreen) removeClient() tea.Cmd {
	return func() tea.Msg {
		err := m.clientStorage.RemoveClient(m.client.UUID)
		msg := removeClientMsg{
			client: m.client,
			err:    err,
		}
		return msg
	}
}

func newRemoveClientScreen(
	mainScreen mainScreen,
	client parlante.Client) removeClientScreen {
	m := removeClientScreen{
		mainScreen:    mainScreen,
		clientStorage: mainScreen.clientStorage,
		client:        client,
		keys:          NewConfirmCancelKeyMap(),
		help:          createHelp(),
	}
	return m
}
