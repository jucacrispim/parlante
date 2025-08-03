package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type nextScreenType int

const (
	screenClient nextScreenType = iota
	screenDomain
	screenComment
)

type mainScreenKeyMap struct {
	Select     key.Binding
	CursorUp   key.Binding
	CursorDown key.Binding
	Quit       key.Binding
}

func (k mainScreenKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Select, k.Quit}
}

// there is no full help for the main screen
func (k mainScreenKeyMap) FullHelp() [][]key.Binding {
	// notest
	return [][]key.Binding{{}}
}

func (k mainScreenKeyMap) GetHelpKey() key.Binding {
	// notest
	return key.NewBinding()
}

func newMainScreenKeyMap() mainScreenKeyMap {
	return mainScreenKeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", MESSAGE_KEY_HELP_UP),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", MESSAGE_KEY_HELP_DOWN),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", MESSAGE_KEY_HELP_SELECT),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", MESSAGE_KEY_HELP_QUIT),
		),
	}
}

type mainScreenItem struct {
	name       string
	descr      string
	screenType nextScreenType
}

func (i mainScreenItem) Title() string       { return i.name }
func (i mainScreenItem) Description() string { return i.descr }
func (i mainScreenItem) FilterValue() string { return i.name }

type mainScreen struct {
	list   CustomKeyMapList
	header Header

	// Database stuff. The main screen has
	// references to all kinds of storage so it can
	// pass along to the specific screens
	clientStorage  parlante.ClientStorage
	domainStorage  parlante.ClientDomainStorage
	CommentStorage parlante.CommentStorage
	keys           *mainScreenKeyMap
}

func (m mainScreen) Init() tea.Cmd {
	return nil
}

func (m mainScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.header.Update(msg)
	l, cmd := m.list.Update(msg)
	nl, _ := l.(CustomKeyMapList)
	m.list = nl

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Select):
			model, cmd := m.getNextAction()
			return model, cmd
		}
	}
	return m, cmd
}

func (m mainScreen) View() string {

	s := hackHeader(m.header.View())

	s += m.list.View()
	return s
}

func (m mainScreen) getNextAction() (tea.Model, tea.Cmd) {
	choice := m.list.SelectedItem().(mainScreenItem)

	switch choice.screenType {
	case screenClient:
		c := newClientListScreen(m)
		return c, c.Init()

	case screenDomain:
		c := newDomainListScreen(&m)
		return c, c.Init()
	case screenComment:
		c := newCommentListScreen(&m)
		return c, c.Init()
	}
	return m, nil // notest
}

func newMainScreen(
	cs parlante.ClientStorage,
	ds parlante.ClientDomainStorage,
	cos parlante.CommentStorage) mainScreen {
	items := []list.Item{
		mainScreenItem{
			MESSAGE_CLIENTS,
			MESSAGE_CLIENTS_SCREEN_DESCR,
			screenClient,
		},
		mainScreenItem{
			MESSAGE_DOMAINS,
			MESSAGE_DOMAINS_SCREEN_DESCR,
			screenDomain,
		},
		mainScreenItem{
			MESSAGE_COMMENTS,
			MESSAGE_COMMENTS_SCREEN_DESCR,
			screenComment,
		},
	}

	opts := ListOpts{
		Title:           MESSAGE_CHOOSE_ONE,
		ShowDescription: true,
		ShowStatusBar:   false,
		ShowHelp:        true,
	}
	keys := newMainScreenKeyMap()
	l := NewCustomKeyMapList(opts, items, keys)
	l.KeyMap.GoToStart.SetEnabled(false)
	l.KeyMap.GoToEnd.SetEnabled(false)
	h := NewHeader()
	m := mainScreen{
		list:           l,
		header:         h,
		clientStorage:  cs,
		domainStorage:  ds,
		CommentStorage: cos,
		keys:           &keys,
	}
	return m
}
