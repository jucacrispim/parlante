package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type addDomainStep int

const (
	selectClient addDomainStep = iota
	addDomain
)

type addDomainMsg struct {
	domain parlante.ClientDomain
	err    error
}

type chooseDomainKeyMap struct {
	list.KeyMap
	ConfirmCancelKeyMap
}

func (k chooseDomainKeyMap) ShortHelp() []key.Binding {
	kb := []key.Binding{
		k.CursorUp, k.CursorDown,
		k.Confirm, k.Cancel}
	return kb
}

func (k chooseDomainKeyMap) FullHelp() [][]key.Binding {
	kb := [][]key.Binding{
		{
			k.KeyMap.CursorUp,
			k.KeyMap.CursorDown,
			k.KeyMap.NextPage,
			k.KeyMap.PrevPage,
			k.KeyMap.GoToStart,
			k.KeyMap.GoToEnd,
		},
		{
			k.Confirm,
			k.Cancel,
		},
	}

	listLevelBindings := []key.Binding{
		k.KeyMap.Filter,
		k.KeyMap.ClearFilter,
		k.KeyMap.AcceptWhileFiltering,
		k.KeyMap.CancelWhileFiltering,
	}

	return append(kb,
		listLevelBindings,
		[]key.Binding{
			k.KeyMap.Quit,
			k.KeyMap.CloseFullHelp,
		})
}

func (k chooseDomainKeyMap) GetHelpKey() key.Binding {
	return k.ShowFullHelp
}

func newChooseDomainKeyMap() chooseDomainKeyMap {
	k := chooseDomainKeyMap{
		KeyMap:              baseListKeyMap(),
		ConfirmCancelKeyMap: NewConfirmCancelKeyMap(),
	}
	return k
}

type addDomainScreen struct {
	mainScreen     *mainScreen
	domainStorage  parlante.ClientDomainStorage
	clientStorage  parlante.ClientStorage
	step           addDomainStep
	clientLoader   *ClientLoader
	clients        CustomKeyMapList
	selectedClient *parlante.Client
	textinput      textinput.Model
	err            error
	keys           chooseDomainKeyMap
	help           help.Model
}

func (m addDomainScreen) Init() tea.Cmd {
	return m.clientLoader.Load()
}

func (m addDomainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.mainScreen.header.Update(msg)
	switch msg := msg.(type) {

	case ItemListMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.clients.SetItems(msg.Items)

	case addDomainMsg:
		m.err = msg.err
		if m.err != nil {
			return m, nil
		}
		model := newDomainListScreen(m.mainScreen)
		return model, model.Init()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			if m.step == selectClient {
				m.step = addDomain
				m.textinput.Focus()
				i := m.clients.SelectedItem()
				item := i.(clientItem)
				m.selectedClient = &item.client
				return m, textinput.Blink
			}
			return m, m.addDomain()

		case key.Matches(msg, m.keys.Cancel):
			model := newDomainListScreen(m.mainScreen)
			return model, model.Init()
		}

	}

	if m.step == selectClient {
		var l tea.Model
		l, cmd = m.clients.Update(msg)
		nl, _ := l.(CustomKeyMapList)
		m.clients = nl

	} else {
		m.textinput, cmd = m.textinput.Update(msg)
	}
	return m, cmd
}

func (m addDomainScreen) View() string {
	var s string
	var title string
	var content string
	helpView := m.help.View(m.keys)
	help := helpViewStyle.Render(helpView)
	if m.err != nil {
		s += m.mainScreen.header.View()
		content = m.err.Error()
		s += content
	} else if m.step == selectClient {
		s = hackHeader(m.mainScreen.header.View())
		content = m.clients.View()
		s += content
	} else {
		s += m.mainScreen.header.View()
		title = "New domain for "
		title = titleStyle.Render(title) + highlightTitleStyle.Render(m.selectedClient.Name)
		content = m.textinput.View()
		s += title + "\n\n" + content
	}

	lines := strings.Split(s, "\n")
	rest := m.mainScreen.list.Height() - len(lines)
	if rest < 0 {
		rest = 0
	}

	s += strings.Repeat("\n", rest) + help

	return s
}

func (m addDomainScreen) addDomain() tea.Cmd {
	return func() tea.Msg {
		domain, err := m.domainStorage.AddClientDomain(
			*m.selectedClient, m.textinput.Value())

		msg := addDomainMsg{
			domain: domain,
			err:    err,
		}
		return msg

	}
}

func newAddDomainScreen(main *mainScreen) addDomainScreen {
	l := ClientLoader{
		Storage: main.clientStorage,
	}

	m := addDomainScreen{
		mainScreen:    main,
		domainStorage: main.domainStorage,
		clientStorage: main.clientStorage,
		step:          selectClient,
		help:          createHelp(),
		keys:          newChooseDomainKeyMap(),
		clientLoader:  &l,
	}
	listOpts := ListOpts{
		ShowDescription: false,
		ShowStatusBar:   false,
		Title:           "Choose a client",
	}
	m.clients = NewCustomKeyMapList(listOpts, []list.Item{}, m.keys)
	ti := textinput.New()
	ti.Width = 20
	ti.Placeholder = "Domain name"
	ti.TextStyle = defaultTextStyle
	ti.PromptStyle = defaultTextStyle
	m.textinput = ti
	return m
}
