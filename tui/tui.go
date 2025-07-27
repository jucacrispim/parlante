package tui

import (
	"os"
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jucacrispim/parlante"
)

type Header struct {
	Text  string
	width int
	Style lipgloss.Style
}

func (m Header) Init() tea.Cmd {
	return nil
}

func (m *Header) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	}
	return nil, nil
}

func (m Header) View() string {
	padded := lipgloss.PlaceHorizontal(m.width, lipgloss.Left, " "+m.Text)
	return m.Style.Render(padded) + "\n\n"
}

func NewHeader() Header {
	h := Header{
		Text:  "Parlante TUI",
		Style: headerStyle,
	}
	return h
}

type CustomKeyMap interface {
	help.KeyMap
	GetHelpKey() key.Binding
}

// List that displays the help for a custom KeyMap
type CustomKeyMapList struct {
	list.Model
	Keys     CustomKeyMap
	Help     help.Model
	showHelp bool
}

func (m *CustomKeyMapList) SetShowHelp(showHelp bool) {
	// the base list help is always hidden because
	// we use our own help
	m.Model.SetShowHelp(false)
	m.showHelp = showHelp
}

func (m CustomKeyMapList) Init() tea.Cmd {
	return nil
}

func (m CustomKeyMapList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.GetHelpKey()):
			m.Help.ShowAll = !m.Help.ShowAll
			return m, nil
		}
	}
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m CustomKeyMapList) View() string {
	content := m.Model.View()
	full := content
	if m.showHelp {
		help := m.Help.View(m.Keys)
		full += helpViewStyle.Render(help)

	}
	return full
}

type ListOpts struct {
	Title           string
	ShowDescription bool
	ShowStatusBar   bool
	ShowHelp        bool
}

func NewCustomKeyMapList(
	opts ListOpts,
	items []list.Item,
	keys CustomKeyMap) CustomKeyMapList {

	width := INNER_WIDTH
	height := INNER_HEIGHT
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = opts.ShowDescription

	l := list.New(items, delegate, width, height)

	m := CustomKeyMapList{
		Model: l,
		Keys:  keys,
		Help:  createHelp(),
	}
	m.Title = opts.Title
	m.SetShowStatusBar(opts.ShowStatusBar)
	m.SetShowHelp(opts.ShowHelp)
	return m
}

// Navigation between screens
type AddRemoveScreenNavigation interface {
	GetAddScreen() tea.Model
	GetRemoveScreen(list.Item) tea.Model
	GetPreviousScreen() tea.Model
}

type ItemListMsg struct {
	Items []list.Item
	Err   error
}

// A function used to load the items of a list
type LoadItemsFn func() tea.Cmd

// A screen with a header and a list in it. It has key bindings to
// add/remove items from the list and for navigation between screens
type AddRemoveItemScreen struct {
	Header     *Header
	ListOpts   ListOpts
	List       CustomKeyMapList
	Navigation AddRemoveScreenNavigation
	KeyMap     ListKeyMap
	LoadItems  LoadItemsFn
	ShowAll    bool
	keys       *ListKeyMap
	err        error
}

func (m AddRemoveItemScreen) Init() tea.Cmd {
	return m.LoadItems()
}

func (m AddRemoveItemScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Header.Update(msg)

	switch msg := msg.(type) {
	case ItemListMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.List.SetItems(msg.Items)
	case tea.KeyMsg:
		if m.List.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, m.keys.Add):
			screen := m.Navigation.GetAddScreen()
			return screen, screen.Init()

		case key.Matches(msg, m.keys.Remove):
			screen := m.Navigation.GetRemoveScreen(m.List.SelectedItem())
			return screen, screen.Init()

		case key.Matches(msg, m.keys.PrevScreen, m.keys.Quit):
			screen := m.Navigation.GetPreviousScreen()
			return screen, screen.Init()
		}
	}
	l, cmd := m.List.Update(msg)
	nl, _ := l.(CustomKeyMapList)
	m.List = nl

	return m, cmd
}

func (m AddRemoveItemScreen) View() string {
	s := hackHeader(m.Header.View())

	if m.err != nil {
		return m.err.Error()
	}

	// change the base list KeyMap so
	// we can use our own help functions
	origKeys := m.List.Keys
	m.List.Keys = &m
	s += m.List.View()
	m.List.Keys = origKeys
	return s
}

func (m AddRemoveItemScreen) ShortHelp() []key.Binding {
	kb := []key.Binding{
		m.keys.CursorUp, m.keys.CursorDown,
		m.keys.Add}
	if len(m.List.Items()) > 0 {
		kb = append(kb, m.keys.Remove)
	}
	kb = append(kb,
		[]key.Binding{m.keys.PrevScreen,
			m.keys.Quit, m.keys.ShowFullHelp}...)
	return kb
}

func (m AddRemoveItemScreen) FullHelp() [][]key.Binding {
	h := m.List.FullHelp()
	col := []key.Binding{
		m.keys.Add,
		m.keys.Remove,
	}
	h = slices.Insert(h, 2, col)
	h[3] = slices.Insert(h[3], 0, m.keys.PrevScreen)
	return h
}

func (m AddRemoveItemScreen) GetHelpKey() key.Binding {
	return m.keys.ShowFullHelp
}

func NewAddRemoveItemScreen(
	header *Header,
	opts ListOpts,
	nav AddRemoveScreenNavigation,
	loadfn LoadItemsFn) AddRemoveItemScreen {

	kb := DefaultAddRemoveItemListKeyMap()
	m := AddRemoveItemScreen{
		Header:     header,
		ListOpts:   opts,
		Navigation: nav,
		LoadItems:  loadfn,
		KeyMap:     kb,
		keys:       &kb,
	}
	m.List = NewCustomKeyMapList(m.ListOpts, []list.Item{}, m.keys)
	return m

}

type ListKeyMap struct {
	list.KeyMap
	PrevScreen key.Binding

	// Items management
	Add     key.Binding
	Remove  key.Binding
	ShowAll bool
}

// DefaultAddRemoveItemListKeyMap returns a default set of keybindings
// for list screens.
func DefaultAddRemoveItemListKeyMap() ListKeyMap {
	basemap := baseListKeyMap()
	return ListKeyMap{
		KeyMap: basemap,

		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),

		Remove: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "remove"),
		),
		PrevScreen: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "previous screen"),
		),
	}
}

func baseListKeyMap() list.KeyMap {
	basemap := list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "u"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),

		// Filtering.
		CancelWhileFiltering: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
	return basemap
}

func (k ListKeyMap) GetHelpKey() key.Binding {
	return k.ShowFullHelp
}

// note that these help functions are here just
// to implement the KeyMap interface, but they are
// not used. In lists that use this KeyMap the
// help functions from the list are used
func (k ListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Quit}
}

func (k ListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.CursorUp, k.CursorDown, k.Quit, k.PrevScreen},
		{k.GoToEnd}}
}

type ConfirmCancelKeyMap struct {
	Cancel  key.Binding
	Confirm key.Binding
}

func (k ConfirmCancelKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

func (k ConfirmCancelKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Confirm, k.Cancel}}
}

func NewConfirmCancelKeyMap() ConfirmCancelKeyMap {
	return ConfirmCancelKeyMap{
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
	}
}

func createHelp() help.Model {
	help := help.New()
	help.Styles.ShortKey = helpKeyStyle
	help.Styles.ShortDesc = helpDescStyle
	help.Styles.FullKey = helpKeyStyle
	help.Styles.ShortDesc = helpDescStyle
	return help
}

func hackHeader(s string) string {
	if isVterm() {
		s = "\n" + s
	}
	return s
}

func isVterm() bool {
	return os.Getenv("INSIDE_EMACS") == "vterm"
}

func NewTui(
	cs parlante.ClientStorage,
	ds parlante.ClientDomainStorage,
	cos parlante.CommentStorage) *tea.Program {
	// notest
	m := newMainScreen(cs, ds, cos)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p
}
