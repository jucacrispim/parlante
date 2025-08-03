package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jucacrispim/parlante"
)

type removeCommentMsg struct {
	Comment parlante.Comment
	err     error
}

type removeCommentScreen struct {
	mainScreen     mainScreen
	CommentStorage parlante.CommentStorage
	Comment        parlante.Comment
	help           help.Model
	keys           ConfirmCancelKeyMap
	err            error
}

func (m removeCommentScreen) Init() tea.Cmd {
	return nil
}

func (m removeCommentScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mainScreen.header.Update(msg)
	switch msg := msg.(type) {
	case removeCommentMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		model := newCommentListScreen(&m.mainScreen)
		return model, model.Init()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.removeComment()

		case "esc":
			model := newCommentListScreen(&m.mainScreen)
			return model, model.Init()

		}
	}
	return m, nil
}

func (m removeCommentScreen) View() string {
	s := m.mainScreen.header.View()
	title := "  " + titleStyle.Render(MESSAGE_REMOVE_COMMENT)
	s += title + "\n\n\n"
	var content string
	if m.err != nil {
		content = m.err.Error()
	} else {
		data := make(map[string]any, 0)
		data["name"] = m.Comment.Author
		data["url"] = m.Comment.PageURL
		content = fmt.Sprintf(
			parlante.Tprintf(MESSAGE_REMOVE_COMMENT_CONFIRM, data))
	}
	s += defaultTextStyle.Render(content)

	lines := strings.Split(s, "\n")
	rest := m.mainScreen.list.Height() - len(lines) + 2

	helpView := m.help.View(m.keys)
	s += strings.Repeat("\n", rest) + helpViewStyle.Render(helpView)

	return s
}

func (m removeCommentScreen) removeComment() tea.Cmd {
	return func() tea.Msg {
		err := m.CommentStorage.RemoveComment(m.Comment)
		msg := removeCommentMsg{
			Comment: m.Comment,
			err:     err,
		}
		return msg
	}
}

func newRemoveCommentScreen(
	mainScreen mainScreen,
	comment parlante.Comment) removeCommentScreen {
	m := removeCommentScreen{
		mainScreen:     mainScreen,
		CommentStorage: mainScreen.CommentStorage,
		Comment:        comment,
		keys:           NewConfirmCancelKeyMap(),
		help:           createHelp(),
	}
	return m
}
